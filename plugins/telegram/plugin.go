// Telegram plugin for M3M
// Build with: go build -buildmode=plugin -o ../telegram.so
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"m3m/pkg/schema"
)

// TelegramPlugin provides Telegram bot functionality to M3M runtime
type TelegramPlugin struct {
	initialized bool
	bots        map[string]*BotInstance
	mu          sync.RWMutex
	storagePath string
}

// BotInstance represents a running Telegram bot
type BotInstance struct {
	bot            *bot.Bot
	ctx            context.Context
	cancel         context.CancelFunc
	runtime        *goja.Runtime
	handlers       map[string]goja.Callable
	callbacks      map[string]goja.Callable
	defaultHandler goja.Callable
}

// UpdateContext provides context for handler callbacks
type UpdateContext struct {
	instance *BotInstance
	update   *models.Update
	runtime  *goja.Runtime
}

func (p *TelegramPlugin) Name() string {
	return "$telegram"
}

func (p *TelegramPlugin) Version() string {
	return "1.0.0"
}

func (p *TelegramPlugin) Init(config map[string]interface{}) error {
	p.bots = make(map[string]*BotInstance)
	p.initialized = true
	if path, ok := config["storage_path"].(string); ok {
		p.storagePath = path
	}
	return nil
}

func (p *TelegramPlugin) RegisterModule(runtime *goja.Runtime) error {
	return runtime.Set("$telegram", map[string]interface{}{
		"startBot": p.createStartBot(runtime),
		"stopBot":  p.stopBot,
		"stopAll":  p.stopAll,
	})
}

func (p *TelegramPlugin) Shutdown() error {
	p.stopAll()
	p.initialized = false
	return nil
}

// createStartBot creates the startBot function with runtime context
func (p *TelegramPlugin) createStartBot(runtime *goja.Runtime) func(string, goja.Callable) error {
	return func(token string, callback goja.Callable) error {
		return p.startBot(runtime, token, callback)
	}
}

// startBot starts a new Telegram bot with the given token
func (p *TelegramPlugin) startBot(runtime *goja.Runtime, token string, callback goja.Callable) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Stop existing bot with same token
	if existing, ok := p.bots[token]; ok {
		existing.cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())

	instance := &BotInstance{
		ctx:       ctx,
		cancel:    cancel,
		runtime:   runtime,
		handlers:  make(map[string]goja.Callable),
		callbacks: make(map[string]goja.Callable),
	}

	// Create bot options with default handler
	opts := []bot.Option{
		bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, update *models.Update) {
			instance.handleUpdate(ctx, b, update)
		}),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create bot: %w", err)
	}

	instance.bot = b
	p.bots[token] = instance

	// Create instance object for JavaScript
	instanceObj := p.createInstanceObject(runtime, instance)

	// Call the setup callback
	if callback != nil {
		_, err := callback(goja.Undefined(), runtime.ToValue(instanceObj))
		if err != nil {
			cancel()
			delete(p.bots, token)
			return fmt.Errorf("setup callback failed: %w", err)
		}
	}

	// Start the bot in background
	go func() {
		b.Start(ctx)
	}()

	return nil
}

// createInstanceObject creates the $instance object for JavaScript
func (p *TelegramPlugin) createInstanceObject(runtime *goja.Runtime, instance *BotInstance) map[string]interface{} {
	return map[string]interface{}{
		// Handler registration
		"handle":         instance.createHandle(),
		"handleCallback": instance.createHandleCallback(),
		"handleDefault":  instance.createHandleDefault(),

		// Message sending
		"sendMessage":  instance.createSendMessage(),
		"sendPhoto":    p.createSendPhoto(instance),
		"sendDocument": p.createSendDocument(instance),
		"sendSticker":  instance.createSendSticker(),
		"sendVideo":    p.createSendVideo(instance),
		"sendAudio":    p.createSendAudio(instance),
		"sendVoice":    p.createSendVoice(instance),

		// Message editing
		"editMessage":      instance.createEditMessage(),
		"editMessageMedia": p.createEditMessageMedia(instance),
		"deleteMessage":    instance.createDeleteMessage(),

		// Callback answers
		"answerCallback": instance.createAnswerCallback(),

		// Bot info
		"getMe": instance.createGetMe(),

		// Utilities
		"getChatMember": instance.createGetChatMember(),
	}
}

// handleUpdate processes incoming updates
func (instance *BotInstance) handleUpdate(ctx context.Context, b *bot.Bot, update *models.Update) {
	uctx := &UpdateContext{
		instance: instance,
		update:   update,
		runtime:  instance.runtime,
	}

	// Handle callback queries
	if update.CallbackQuery != nil {
		if handler, ok := instance.callbacks[update.CallbackQuery.Data]; ok {
			instance.callHandler(handler, uctx)
			return
		}
		// Try prefix match for callbacks with data
		for pattern, handler := range instance.callbacks {
			if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
				prefix := pattern[:len(pattern)-1]
				if len(update.CallbackQuery.Data) >= len(prefix) && update.CallbackQuery.Data[:len(prefix)] == prefix {
					instance.callHandler(handler, uctx)
					return
				}
			}
		}
		if instance.defaultHandler != nil {
			instance.callHandler(instance.defaultHandler, uctx)
		}
		return
	}

	// Handle messages
	if update.Message != nil {
		text := update.Message.Text

		// Try exact match first
		if handler, ok := instance.handlers[text]; ok {
			instance.callHandler(handler, uctx)
			return
		}

		// Try command match (e.g., "/start" matches "/start@botname")
		for pattern, handler := range instance.handlers {
			if len(pattern) > 0 && pattern[0] == '/' {
				// Command pattern
				cmdLen := len(pattern)
				if len(text) >= cmdLen && text[:cmdLen] == pattern {
					if len(text) == cmdLen || text[cmdLen] == ' ' || text[cmdLen] == '@' {
						instance.callHandler(handler, uctx)
						return
					}
				}
			}
		}

		// Default handler
		if instance.defaultHandler != nil {
			instance.callHandler(instance.defaultHandler, uctx)
		}
	}
}

// callHandler safely calls a JavaScript handler
func (instance *BotInstance) callHandler(handler goja.Callable, uctx *UpdateContext) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Handler panic: %v\n", r)
		}
	}()

	ctxObj := instance.createContextObject(uctx)
	handler(goja.Undefined(), instance.runtime.ToValue(ctxObj))
}

// createContextObject creates the context object passed to handlers
func (instance *BotInstance) createContextObject(uctx *UpdateContext) map[string]interface{} {
	ctx := map[string]interface{}{
		"update":                  uctx.convertUpdate(),
		"reply":                   uctx.createReply(),
		"replyPhoto":              uctx.createReplyPhoto(),
		"replyWithKeyboard":       uctx.createReplyWithKeyboard(),
		"replyWithInlineKeyboard": uctx.createReplyWithInlineKeyboard(),
		"answerCallback":          uctx.createAnswerCallback(),
		"editMessage":             uctx.createEditMessage(),
		"deleteMessage":           uctx.createDeleteMessage(),
	}
	return ctx
}

// convertUpdate converts models.Update to JavaScript-friendly object
func (uctx *UpdateContext) convertUpdate() map[string]interface{} {
	u := uctx.update
	result := map[string]interface{}{
		"updateId": u.ID,
	}

	if u.Message != nil {
		result["message"] = uctx.convertMessage(u.Message)
	}
	if u.CallbackQuery != nil {
		result["callbackQuery"] = map[string]interface{}{
			"id":           u.CallbackQuery.ID,
			"from":         uctx.convertUser(&u.CallbackQuery.From),
			"data":         u.CallbackQuery.Data,
			"chatInstance": u.CallbackQuery.ChatInstance,
		}
		if u.CallbackQuery.Message.Message != nil {
			result["callbackQuery"].(map[string]interface{})["message"] = uctx.convertMessage(u.CallbackQuery.Message.Message)
		}
	}

	return result
}

func (uctx *UpdateContext) convertMessage(m *models.Message) map[string]interface{} {
	msg := map[string]interface{}{
		"messageId": m.ID,
		"date":      m.Date,
		"text":      m.Text,
		"chat":      uctx.convertChat(m.Chat),
	}
	if m.From != nil {
		msg["from"] = uctx.convertUser(m.From)
	}
	if m.Photo != nil && len(m.Photo) > 0 {
		photos := make([]map[string]interface{}, len(m.Photo))
		for i, p := range m.Photo {
			photos[i] = map[string]interface{}{
				"fileId":       p.FileID,
				"fileUniqueId": p.FileUniqueID,
				"width":        p.Width,
				"height":       p.Height,
				"fileSize":     p.FileSize,
			}
		}
		msg["photo"] = photos
	}
	if m.Document != nil {
		msg["document"] = map[string]interface{}{
			"fileId":       m.Document.FileID,
			"fileUniqueId": m.Document.FileUniqueID,
			"fileName":     m.Document.FileName,
			"mimeType":     m.Document.MimeType,
			"fileSize":     m.Document.FileSize,
		}
	}
	return msg
}

func (uctx *UpdateContext) convertChat(c models.Chat) map[string]interface{} {
	return map[string]interface{}{
		"id":        c.ID,
		"type":      c.Type,
		"title":     c.Title,
		"username":  c.Username,
		"firstName": c.FirstName,
		"lastName":  c.LastName,
	}
}

func (uctx *UpdateContext) convertUser(u *models.User) map[string]interface{} {
	if u == nil {
		return nil
	}
	return map[string]interface{}{
		"id":           u.ID,
		"isBot":        u.IsBot,
		"firstName":    u.FirstName,
		"lastName":     u.LastName,
		"username":     u.Username,
		"languageCode": u.LanguageCode,
	}
}

// Reply functions
func (uctx *UpdateContext) createReply() func(string) (map[string]interface{}, error) {
	return func(text string) (map[string]interface{}, error) {
		chatID := uctx.getChatID()
		if chatID == 0 {
			return nil, fmt.Errorf("no chat ID available")
		}

		msg, err := uctx.instance.bot.SendMessage(uctx.instance.ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeHTML,
		})
		if err != nil {
			return nil, err
		}
		return uctx.convertMessage(msg), nil
	}
}

func (uctx *UpdateContext) createReplyPhoto() func(string, string) (map[string]interface{}, error) {
	return func(photo string, caption string) (map[string]interface{}, error) {
		chatID := uctx.getChatID()
		if chatID == 0 {
			return nil, fmt.Errorf("no chat ID available")
		}

		params := &bot.SendPhotoParams{
			ChatID:    chatID,
			Caption:   caption,
			ParseMode: models.ParseModeHTML,
		}

		// Check if it's a file path, URL or file_id
		if isFilePath(photo) {
			data, err := os.ReadFile(photo)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}
			params.Photo = &models.InputFileUpload{
				Filename: filepath.Base(photo),
				Data:     bytes.NewReader(data),
			}
		} else {
			params.Photo = &models.InputFileString{Data: photo}
		}

		msg, err := uctx.instance.bot.SendPhoto(uctx.instance.ctx, params)
		if err != nil {
			return nil, err
		}
		return uctx.convertMessage(msg), nil
	}
}

func (uctx *UpdateContext) createReplyWithKeyboard() func(string, [][]map[string]interface{}, map[string]interface{}) (map[string]interface{}, error) {
	return func(text string, keyboard [][]map[string]interface{}, options map[string]interface{}) (map[string]interface{}, error) {
		chatID := uctx.getChatID()
		if chatID == 0 {
			return nil, fmt.Errorf("no chat ID available")
		}

		kb := buildReplyKeyboard(keyboard, options)

		msg, err := uctx.instance.bot.SendMessage(uctx.instance.ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: kb,
		})
		if err != nil {
			return nil, err
		}
		return uctx.convertMessage(msg), nil
	}
}

func (uctx *UpdateContext) createReplyWithInlineKeyboard() func(string, [][]map[string]interface{}) (map[string]interface{}, error) {
	return func(text string, keyboard [][]map[string]interface{}) (map[string]interface{}, error) {
		chatID := uctx.getChatID()
		if chatID == 0 {
			return nil, fmt.Errorf("no chat ID available")
		}

		kb := buildInlineKeyboard(keyboard)

		msg, err := uctx.instance.bot.SendMessage(uctx.instance.ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: kb,
		})
		if err != nil {
			return nil, err
		}
		return uctx.convertMessage(msg), nil
	}
}

func (uctx *UpdateContext) createAnswerCallback() func(string, bool) error {
	return func(text string, showAlert bool) error {
		if uctx.update.CallbackQuery == nil {
			return nil
		}
		_, err := uctx.instance.bot.AnswerCallbackQuery(uctx.instance.ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: uctx.update.CallbackQuery.ID,
			Text:            text,
			ShowAlert:       showAlert,
		})
		return err
	}
}

func (uctx *UpdateContext) createEditMessage() func(string, map[string]interface{}) (map[string]interface{}, error) {
	return func(text string, options map[string]interface{}) (map[string]interface{}, error) {
		var chatID int64
		var messageID int

		if uctx.update.CallbackQuery != nil && uctx.update.CallbackQuery.Message.Message != nil {
			msg := uctx.update.CallbackQuery.Message.Message
			chatID = msg.Chat.ID
			messageID = msg.ID
		}

		if chatID == 0 || messageID == 0 {
			return nil, fmt.Errorf("no message to edit")
		}

		params := &bot.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: messageID,
			Text:      text,
			ParseMode: models.ParseModeHTML,
		}

		// Handle inline keyboard in options
		if keyboard, ok := options["inlineKeyboard"].([][]map[string]interface{}); ok {
			params.ReplyMarkup = buildInlineKeyboard(keyboard)
		}

		msg, err := uctx.instance.bot.EditMessageText(uctx.instance.ctx, params)
		if err != nil {
			return nil, err
		}
		return uctx.convertMessage(msg), nil
	}
}

func (uctx *UpdateContext) createDeleteMessage() func() error {
	return func() error {
		var chatID int64
		var messageID int

		if uctx.update.Message != nil {
			chatID = uctx.update.Message.Chat.ID
			messageID = uctx.update.Message.ID
		} else if uctx.update.CallbackQuery != nil && uctx.update.CallbackQuery.Message.Message != nil {
			msg := uctx.update.CallbackQuery.Message.Message
			chatID = msg.Chat.ID
			messageID = msg.ID
		}

		if chatID == 0 || messageID == 0 {
			return fmt.Errorf("no message to delete")
		}

		_, err := uctx.instance.bot.DeleteMessage(uctx.instance.ctx, &bot.DeleteMessageParams{
			ChatID:    chatID,
			MessageID: messageID,
		})
		return err
	}
}

func (uctx *UpdateContext) getChatID() int64 {
	if uctx.update.Message != nil {
		return uctx.update.Message.Chat.ID
	}
	if uctx.update.CallbackQuery != nil && uctx.update.CallbackQuery.Message.Message != nil {
		return uctx.update.CallbackQuery.Message.Message.Chat.ID
	}
	return 0
}

// Instance methods
func (instance *BotInstance) createHandle() func(string, goja.Callable) {
	return func(pattern string, handler goja.Callable) {
		instance.handlers[pattern] = handler
	}
}

func (instance *BotInstance) createHandleCallback() func(string, goja.Callable) {
	return func(data string, handler goja.Callable) {
		instance.callbacks[data] = handler
	}
}

func (instance *BotInstance) createHandleDefault() func(goja.Callable) {
	return func(handler goja.Callable) {
		instance.defaultHandler = handler
	}
}

func (instance *BotInstance) createSendMessage() func(int64, string, map[string]interface{}) (map[string]interface{}, error) {
	return func(chatID int64, text string, options map[string]interface{}) (map[string]interface{}, error) {
		params := &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeHTML,
		}

		if options != nil {
			if keyboard, ok := options["inlineKeyboard"].([][]map[string]interface{}); ok {
				params.ReplyMarkup = buildInlineKeyboard(keyboard)
			} else if keyboard, ok := options["keyboard"].([][]map[string]interface{}); ok {
				keyboardOpts := make(map[string]interface{})
				if resize, ok := options["resizeKeyboard"].(bool); ok {
					keyboardOpts["resize"] = resize
				}
				if oneTime, ok := options["oneTimeKeyboard"].(bool); ok {
					keyboardOpts["oneTime"] = oneTime
				}
				params.ReplyMarkup = buildReplyKeyboard(keyboard, keyboardOpts)
			} else if removeKb, ok := options["removeKeyboard"].(bool); ok && removeKb {
				params.ReplyMarkup = &models.ReplyKeyboardRemove{RemoveKeyboard: true}
			}

			if parseMode, ok := options["parseMode"].(string); ok {
				params.ParseMode = models.ParseMode(parseMode)
			}
			if disablePreview, ok := options["disableWebPagePreview"].(bool); ok && disablePreview {
				disabled := true
				params.LinkPreviewOptions = &models.LinkPreviewOptions{
					IsDisabled: &disabled,
				}
			}
		}

		msg, err := instance.bot.SendMessage(instance.ctx, params)
		if err != nil {
			return nil, err
		}
		return (&UpdateContext{instance: instance}).convertMessage(msg), nil
	}
}

func (p *TelegramPlugin) createSendPhoto(instance *BotInstance) func(int64, string, map[string]interface{}) (map[string]interface{}, error) {
	return func(chatID int64, photo string, options map[string]interface{}) (map[string]interface{}, error) {
		params := &bot.SendPhotoParams{
			ChatID:    chatID,
			ParseMode: models.ParseModeHTML,
		}

		if options != nil {
			if caption, ok := options["caption"].(string); ok {
				params.Caption = caption
			}
			if keyboard, ok := options["inlineKeyboard"].([][]map[string]interface{}); ok {
				params.ReplyMarkup = buildInlineKeyboard(keyboard)
			}
		}

		// Determine photo source
		photo = p.resolvePath(photo)
		if isFilePath(photo) {
			data, err := os.ReadFile(photo)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}
			params.Photo = &models.InputFileUpload{
				Filename: filepath.Base(photo),
				Data:     bytes.NewReader(data),
			}
		} else if isBase64(photo) {
			data, err := base64.StdEncoding.DecodeString(photo)
			if err != nil {
				return nil, fmt.Errorf("invalid base64: %w", err)
			}
			params.Photo = &models.InputFileUpload{
				Filename: "image.png",
				Data:     bytes.NewReader(data),
			}
		} else {
			params.Photo = &models.InputFileString{Data: photo}
		}

		msg, err := instance.bot.SendPhoto(instance.ctx, params)
		if err != nil {
			return nil, err
		}
		return (&UpdateContext{instance: instance}).convertMessage(msg), nil
	}
}

func (p *TelegramPlugin) createSendDocument(instance *BotInstance) func(int64, string, map[string]interface{}) (map[string]interface{}, error) {
	return func(chatID int64, document string, options map[string]interface{}) (map[string]interface{}, error) {
		params := &bot.SendDocumentParams{
			ChatID:    chatID,
			ParseMode: models.ParseModeHTML,
		}

		filename := "document"
		if options != nil {
			if caption, ok := options["caption"].(string); ok {
				params.Caption = caption
			}
			if fn, ok := options["filename"].(string); ok {
				filename = fn
			}
			if keyboard, ok := options["inlineKeyboard"].([][]map[string]interface{}); ok {
				params.ReplyMarkup = buildInlineKeyboard(keyboard)
			}
		}

		document = p.resolvePath(document)
		if isFilePath(document) {
			data, err := os.ReadFile(document)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}
			params.Document = &models.InputFileUpload{
				Filename: filepath.Base(document),
				Data:     bytes.NewReader(data),
			}
		} else if isBase64(document) {
			data, err := base64.StdEncoding.DecodeString(document)
			if err != nil {
				return nil, fmt.Errorf("invalid base64: %w", err)
			}
			params.Document = &models.InputFileUpload{
				Filename: filename,
				Data:     bytes.NewReader(data),
			}
		} else {
			params.Document = &models.InputFileString{Data: document}
		}

		msg, err := instance.bot.SendDocument(instance.ctx, params)
		if err != nil {
			return nil, err
		}
		return (&UpdateContext{instance: instance}).convertMessage(msg), nil
	}
}

func (instance *BotInstance) createSendSticker() func(int64, string, map[string]interface{}) (map[string]interface{}, error) {
	return func(chatID int64, sticker string, options map[string]interface{}) (map[string]interface{}, error) {
		params := &bot.SendStickerParams{
			ChatID:  chatID,
			Sticker: &models.InputFileString{Data: sticker},
		}

		if options != nil {
			if keyboard, ok := options["inlineKeyboard"].([][]map[string]interface{}); ok {
				params.ReplyMarkup = buildInlineKeyboard(keyboard)
			}
		}

		msg, err := instance.bot.SendSticker(instance.ctx, params)
		if err != nil {
			return nil, err
		}
		return (&UpdateContext{instance: instance}).convertMessage(msg), nil
	}
}

func (p *TelegramPlugin) createSendVideo(instance *BotInstance) func(int64, string, map[string]interface{}) (map[string]interface{}, error) {
	return func(chatID int64, video string, options map[string]interface{}) (map[string]interface{}, error) {
		params := &bot.SendVideoParams{
			ChatID:    chatID,
			ParseMode: models.ParseModeHTML,
		}

		if options != nil {
			if caption, ok := options["caption"].(string); ok {
				params.Caption = caption
			}
			if keyboard, ok := options["inlineKeyboard"].([][]map[string]interface{}); ok {
				params.ReplyMarkup = buildInlineKeyboard(keyboard)
			}
		}

		video = p.resolvePath(video)
		if isFilePath(video) {
			data, err := os.ReadFile(video)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}
			params.Video = &models.InputFileUpload{
				Filename: filepath.Base(video),
				Data:     bytes.NewReader(data),
			}
		} else {
			params.Video = &models.InputFileString{Data: video}
		}

		msg, err := instance.bot.SendVideo(instance.ctx, params)
		if err != nil {
			return nil, err
		}
		return (&UpdateContext{instance: instance}).convertMessage(msg), nil
	}
}

func (p *TelegramPlugin) createSendAudio(instance *BotInstance) func(int64, string, map[string]interface{}) (map[string]interface{}, error) {
	return func(chatID int64, audio string, options map[string]interface{}) (map[string]interface{}, error) {
		params := &bot.SendAudioParams{
			ChatID:    chatID,
			ParseMode: models.ParseModeHTML,
		}

		if options != nil {
			if caption, ok := options["caption"].(string); ok {
				params.Caption = caption
			}
			if keyboard, ok := options["inlineKeyboard"].([][]map[string]interface{}); ok {
				params.ReplyMarkup = buildInlineKeyboard(keyboard)
			}
		}

		audio = p.resolvePath(audio)
		if isFilePath(audio) {
			data, err := os.ReadFile(audio)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}
			params.Audio = &models.InputFileUpload{
				Filename: filepath.Base(audio),
				Data:     bytes.NewReader(data),
			}
		} else {
			params.Audio = &models.InputFileString{Data: audio}
		}

		msg, err := instance.bot.SendAudio(instance.ctx, params)
		if err != nil {
			return nil, err
		}
		return (&UpdateContext{instance: instance}).convertMessage(msg), nil
	}
}

func (p *TelegramPlugin) createSendVoice(instance *BotInstance) func(int64, string, map[string]interface{}) (map[string]interface{}, error) {
	return func(chatID int64, voice string, options map[string]interface{}) (map[string]interface{}, error) {
		params := &bot.SendVoiceParams{
			ChatID:    chatID,
			ParseMode: models.ParseModeHTML,
		}

		if options != nil {
			if caption, ok := options["caption"].(string); ok {
				params.Caption = caption
			}
			if keyboard, ok := options["inlineKeyboard"].([][]map[string]interface{}); ok {
				params.ReplyMarkup = buildInlineKeyboard(keyboard)
			}
		}

		voice = p.resolvePath(voice)
		if isFilePath(voice) {
			data, err := os.ReadFile(voice)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}
			params.Voice = &models.InputFileUpload{
				Filename: filepath.Base(voice),
				Data:     bytes.NewReader(data),
			}
		} else {
			params.Voice = &models.InputFileString{Data: voice}
		}

		msg, err := instance.bot.SendVoice(instance.ctx, params)
		if err != nil {
			return nil, err
		}
		return (&UpdateContext{instance: instance}).convertMessage(msg), nil
	}
}

func (instance *BotInstance) createEditMessage() func(int64, int, string, map[string]interface{}) (map[string]interface{}, error) {
	return func(chatID int64, messageID int, text string, options map[string]interface{}) (map[string]interface{}, error) {
		params := &bot.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: messageID,
			Text:      text,
			ParseMode: models.ParseModeHTML,
		}

		if options != nil {
			if keyboard, ok := options["inlineKeyboard"].([][]map[string]interface{}); ok {
				params.ReplyMarkup = buildInlineKeyboard(keyboard)
			}
		}

		msg, err := instance.bot.EditMessageText(instance.ctx, params)
		if err != nil {
			return nil, err
		}
		return (&UpdateContext{instance: instance}).convertMessage(msg), nil
	}
}

func (p *TelegramPlugin) createEditMessageMedia(instance *BotInstance) func(int64, int, string, map[string]interface{}) (map[string]interface{}, error) {
	return func(chatID int64, messageID int, photo string, options map[string]interface{}) (map[string]interface{}, error) {
		caption := ""
		if options != nil {
			if c, ok := options["caption"].(string); ok {
				caption = c
			}
		}

		// Note: EditMessageMedia only supports URL or file_id, not file uploads
		// For file uploads, delete the message and send a new one
		photo = p.resolvePath(photo)
		media := &models.InputMediaPhoto{
			Media:     photo,
			Caption:   caption,
			ParseMode: models.ParseModeHTML,
		}

		params := &bot.EditMessageMediaParams{
			ChatID:    chatID,
			MessageID: messageID,
			Media:     media,
		}

		if options != nil {
			if keyboard, ok := options["inlineKeyboard"].([][]map[string]interface{}); ok {
				params.ReplyMarkup = buildInlineKeyboard(keyboard)
			}
		}

		msg, err := instance.bot.EditMessageMedia(instance.ctx, params)
		if err != nil {
			return nil, err
		}
		return (&UpdateContext{instance: instance}).convertMessage(msg), nil
	}
}

func (instance *BotInstance) createDeleteMessage() func(int64, int) error {
	return func(chatID int64, messageID int) error {
		_, err := instance.bot.DeleteMessage(instance.ctx, &bot.DeleteMessageParams{
			ChatID:    chatID,
			MessageID: messageID,
		})
		return err
	}
}

func (instance *BotInstance) createAnswerCallback() func(string, string, bool) error {
	return func(callbackID string, text string, showAlert bool) error {
		_, err := instance.bot.AnswerCallbackQuery(instance.ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callbackID,
			Text:            text,
			ShowAlert:       showAlert,
		})
		return err
	}
}

func (instance *BotInstance) createGetMe() func() (map[string]interface{}, error) {
	return func() (map[string]interface{}, error) {
		user, err := instance.bot.GetMe(instance.ctx)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"id":           user.ID,
			"isBot":        user.IsBot,
			"firstName":    user.FirstName,
			"lastName":     user.LastName,
			"username":     user.Username,
			"languageCode": user.LanguageCode,
		}, nil
	}
}

func (instance *BotInstance) createGetChatMember() func(int64, int64) (map[string]interface{}, error) {
	return func(chatID int64, userID int64) (map[string]interface{}, error) {
		member, err := instance.bot.GetChatMember(instance.ctx, &bot.GetChatMemberParams{
			ChatID: chatID,
			UserID: userID,
		})
		if err != nil {
			return nil, err
		}

		result := map[string]interface{}{
			"status": string(member.Type),
		}

		uctx := &UpdateContext{instance: instance}
		switch member.Type {
		case models.ChatMemberTypeOwner:
			if member.Owner != nil {
				result["user"] = uctx.convertUser(member.Owner.User)
			}
		case models.ChatMemberTypeAdministrator:
			if member.Administrator != nil {
				result["user"] = uctx.convertUser(&member.Administrator.User)
			}
		case models.ChatMemberTypeMember:
			if member.Member != nil {
				result["user"] = uctx.convertUser(member.Member.User)
			}
		case models.ChatMemberTypeRestricted:
			if member.Restricted != nil {
				result["user"] = uctx.convertUser(member.Restricted.User)
			}
		case models.ChatMemberTypeLeft:
			if member.Left != nil {
				result["user"] = uctx.convertUser(member.Left.User)
			}
		case models.ChatMemberTypeBanned:
			if member.Banned != nil {
				result["user"] = uctx.convertUser(member.Banned.User)
			}
		}

		return result, nil
	}
}

// stopBot stops a bot by token
func (p *TelegramPlugin) stopBot(token string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if instance, ok := p.bots[token]; ok {
		instance.cancel()
		delete(p.bots, token)
	}
}

// stopAll stops all bots
func (p *TelegramPlugin) stopAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for token, instance := range p.bots {
		instance.cancel()
		delete(p.bots, token)
	}
}

// Helper functions
func buildInlineKeyboard(keyboard [][]map[string]interface{}) *models.InlineKeyboardMarkup {
	rows := make([][]models.InlineKeyboardButton, len(keyboard))
	for i, row := range keyboard {
		buttons := make([]models.InlineKeyboardButton, len(row))
		for j, btn := range row {
			button := models.InlineKeyboardButton{
				Text: getString(btn, "text"),
			}
			if url := getString(btn, "url"); url != "" {
				button.URL = url
			}
			if data := getString(btn, "callbackData"); data != "" {
				button.CallbackData = data
			} else if data := getString(btn, "callback_data"); data != "" {
				button.CallbackData = data
			}
			buttons[j] = button
		}
		rows[i] = buttons
	}
	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func buildReplyKeyboard(keyboard [][]map[string]interface{}, options map[string]interface{}) *models.ReplyKeyboardMarkup {
	rows := make([][]models.KeyboardButton, len(keyboard))
	for i, row := range keyboard {
		buttons := make([]models.KeyboardButton, len(row))
		for j, btn := range row {
			button := models.KeyboardButton{
				Text: getString(btn, "text"),
			}
			if contact, ok := btn["requestContact"].(bool); ok {
				button.RequestContact = contact
			}
			if location, ok := btn["requestLocation"].(bool); ok {
				button.RequestLocation = location
			}
			buttons[j] = button
		}
		rows[i] = buttons
	}

	kb := &models.ReplyKeyboardMarkup{
		Keyboard:       rows,
		ResizeKeyboard: true,
	}

	if options != nil {
		if resize, ok := options["resize"].(bool); ok {
			kb.ResizeKeyboard = resize
		}
		if oneTime, ok := options["oneTime"].(bool); ok {
			kb.OneTimeKeyboard = oneTime
		}
		if placeholder, ok := options["placeholder"].(string); ok {
			kb.InputFieldPlaceholder = placeholder
		}
	}

	return kb
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func isFilePath(s string) bool {
	// Check if it looks like a file path
	if len(s) == 0 {
		return false
	}
	// Absolute paths or relative paths with extension
	if s[0] == '/' || s[0] == '.' {
		return true
	}
	// Check for file extension and not URL
	if filepath.Ext(s) != "" && !isURL(s) {
		return true
	}
	return false
}

func isURL(s string) bool {
	return len(s) > 7 && (s[:7] == "http://" || s[:8] == "https://")
}

func isBase64(s string) bool {
	if len(s) < 100 {
		return false
	}
	// Try to decode first 100 chars to check if valid base64
	_, err := base64.StdEncoding.DecodeString(s[:100])
	return err == nil
}

func (p *TelegramPlugin) resolvePath(path string) string {
	if p.storagePath != "" && !filepath.IsAbs(path) && isFilePath(path) {
		return filepath.Join(p.storagePath, path)
	}
	return path
}

// GetSchema returns the schema for TypeScript generation
func (p *TelegramPlugin) GetSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$telegram",
		Description: "Telegram Bot API plugin",
		Methods: []schema.MethodSchema{
			{
				Name:        "startBot",
				Description: "Start a new Telegram bot",
				Params: []schema.ParamSchema{
					{Name: "token", Type: "string", Description: "Bot token from @BotFather"},
					{Name: "setup", Type: "(instance: TelegramBotInstance) => void", Description: "Setup callback"},
				},
			},
			{
				Name:        "stopBot",
				Description: "Stop a bot by token",
				Params: []schema.ParamSchema{
					{Name: "token", Type: "string", Description: "Bot token"},
				},
			},
			{
				Name:        "stopAll",
				Description: "Stop all running bots",
			},
		},
		RawTypes: `interface TelegramUser {
    id: number;
    isBot: boolean;
    firstName: string;
    lastName?: string;
    username?: string;
    languageCode?: string;
}

interface TelegramChat {
    id: number;
    type: string;
    title?: string;
    username?: string;
    firstName?: string;
    lastName?: string;
}

interface TelegramPhotoSize {
    fileId: string;
    fileUniqueId: string;
    width: number;
    height: number;
    fileSize?: number;
}

interface TelegramDocument {
    fileId: string;
    fileUniqueId: string;
    fileName?: string;
    mimeType?: string;
    fileSize?: number;
}

interface TelegramMessage {
    messageId: number;
    date: number;
    text?: string;
    chat: TelegramChat;
    from?: TelegramUser;
    photo?: TelegramPhotoSize[];
    document?: TelegramDocument;
}

interface TelegramCallbackQuery {
    id: string;
    from: TelegramUser;
    data?: string;
    chatInstance: string;
    message?: TelegramMessage;
}

interface TelegramUpdate {
    updateId: number;
    message?: TelegramMessage;
    callbackQuery?: TelegramCallbackQuery;
}

interface InlineKeyboardButton {
    text: string;
    url?: string;
    callbackData?: string;
    callback_data?: string;
}

interface KeyboardButton {
    text: string;
    requestContact?: boolean;
    requestLocation?: boolean;
}

interface SendMessageOptions {
    inlineKeyboard?: InlineKeyboardButton[][];
    keyboard?: KeyboardButton[][];
    removeKeyboard?: boolean;
    resizeKeyboard?: boolean;
    oneTimeKeyboard?: boolean;
    parseMode?: "HTML" | "Markdown" | "MarkdownV2";
    disableWebPagePreview?: boolean;
}

interface SendPhotoOptions {
    caption?: string;
    inlineKeyboard?: InlineKeyboardButton[][];
}

interface SendDocumentOptions {
    caption?: string;
    filename?: string;
    inlineKeyboard?: InlineKeyboardButton[][];
}

interface EditMessageOptions {
    inlineKeyboard?: InlineKeyboardButton[][];
}

interface TelegramContext {
    /** The raw update object */
    update: TelegramUpdate;
    /** Reply with a text message */
    reply(text: string): TelegramMessage;
    /** Reply with a photo */
    replyPhoto(photo: string, caption?: string): TelegramMessage;
    /** Reply with text and reply keyboard */
    replyWithKeyboard(text: string, keyboard: KeyboardButton[][], options?: { resize?: boolean; oneTime?: boolean; placeholder?: string }): TelegramMessage;
    /** Reply with text and inline keyboard */
    replyWithInlineKeyboard(text: string, keyboard: InlineKeyboardButton[][]): TelegramMessage;
    /** Answer callback query (for inline buttons) */
    answerCallback(text?: string, showAlert?: boolean): void;
    /** Edit the message (for callback queries) */
    editMessage(text: string, options?: EditMessageOptions): TelegramMessage;
    /** Delete the current message */
    deleteMessage(): void;
}

interface TelegramBotInstance {
    /** Register a handler for a command or text pattern */
    handle(pattern: string, handler: (ctx: TelegramContext) => void): void;
    /** Register a handler for callback query data */
    handleCallback(data: string, handler: (ctx: TelegramContext) => void): void;
    /** Register a default handler for unmatched messages */
    handleDefault(handler: (ctx: TelegramContext) => void): void;
    /** Send a text message */
    sendMessage(chatId: number, text: string, options?: SendMessageOptions): TelegramMessage;
    /** Send a photo (file path, URL, file_id, or base64) */
    sendPhoto(chatId: number, photo: string, options?: SendPhotoOptions): TelegramMessage;
    /** Send a document (file path, URL, file_id, or base64) */
    sendDocument(chatId: number, document: string, options?: SendDocumentOptions): TelegramMessage;
    /** Send a sticker */
    sendSticker(chatId: number, sticker: string, options?: SendMessageOptions): TelegramMessage;
    /** Send a video (file path, URL, or file_id) */
    sendVideo(chatId: number, video: string, options?: SendPhotoOptions): TelegramMessage;
    /** Send audio (file path, URL, or file_id) */
    sendAudio(chatId: number, audio: string, options?: SendPhotoOptions): TelegramMessage;
    /** Send voice message (file path, URL, or file_id) */
    sendVoice(chatId: number, voice: string, options?: SendPhotoOptions): TelegramMessage;
    /** Edit a message */
    editMessage(chatId: number, messageId: number, text: string, options?: EditMessageOptions): TelegramMessage;
    /** Edit message media (photo) */
    editMessageMedia(chatId: number, messageId: number, photo: string, options?: SendPhotoOptions & EditMessageOptions): TelegramMessage;
    /** Delete a message */
    deleteMessage(chatId: number, messageId: number): void;
    /** Answer a callback query */
    answerCallback(callbackId: string, text?: string, showAlert?: boolean): void;
    /** Get bot info */
    getMe(): TelegramUser;
    /** Get chat member info */
    getChatMember(chatId: number, userId: number): { status: string; user?: TelegramUser };
}`,
	}
}

// NewPlugin is the exported function that returns a new plugin instance
func NewPlugin() interface{} {
	return &TelegramPlugin{}
}

// Add a small delay for initialization
func init() {
	time.Sleep(10 * time.Millisecond)
}
