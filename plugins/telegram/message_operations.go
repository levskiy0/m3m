package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"m3m/pkg/plugin"
)

// Context reply methods

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

		// Resolve path relative to storage
		photo = plugin.MustResolvePath(uctx.instance.storagePath, photo)

		// Check if it's a file path, URL or file_id
		if plugin.IsFilePath(photo) {
			data, err := os.ReadFile(photo)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}
			params.Photo = &models.InputFileUpload{
				Filename: filepath.Base(photo),
				Data:     bytes.NewReader(data),
			}
		} else if plugin.IsBase64(photo) {
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

		msg, err := uctx.instance.bot.SendPhoto(uctx.instance.ctx, params)
		if err != nil {
			return nil, err
		}
		return uctx.convertMessage(msg), nil
	}
}

func (uctx *UpdateContext) createReplyWithKeyboard() func(string, interface{}, map[string]interface{}) (map[string]interface{}, error) {
	return func(text string, keyboardRaw interface{}, options map[string]interface{}) (map[string]interface{}, error) {
		chatID := uctx.getChatID()
		if chatID == 0 {
			return nil, fmt.Errorf("no chat ID available")
		}

		keyboard := convertToKeyboardRows(keyboardRaw)
		if keyboard == nil {
			return nil, fmt.Errorf("invalid keyboard format")
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

func (uctx *UpdateContext) createReplyWithInlineKeyboard() func(string, interface{}) (map[string]interface{}, error) {
	return func(text string, keyboardRaw interface{}) (map[string]interface{}, error) {
		chatID := uctx.getChatID()
		if chatID == 0 {
			return nil, fmt.Errorf("no chat ID available")
		}

		keyboard := convertToKeyboardRows(keyboardRaw)
		if keyboard == nil {
			return nil, fmt.Errorf("invalid keyboard format")
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
		if kb := options["inlineKeyboard"]; kb != nil {
			if keyboard := convertToKeyboardRows(kb); keyboard != nil {
				params.ReplyMarkup = buildInlineKeyboard(keyboard)
			}
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

// Instance direct send methods

func (instance *BotInstance) createSendMessage() func(int64, string, map[string]interface{}) (map[string]interface{}, error) {
	return func(chatID int64, text string, options map[string]interface{}) (map[string]interface{}, error) {
		params := &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeHTML,
		}

		if options != nil {
			if kb := options["inlineKeyboard"]; kb != nil {
				if keyboard := convertToKeyboardRows(kb); keyboard != nil {
					params.ReplyMarkup = buildInlineKeyboard(keyboard)
				}
			} else if kb := options["keyboard"]; kb != nil {
				if keyboard := convertToKeyboardRows(kb); keyboard != nil {
					keyboardOpts := make(map[string]interface{})
					if resize, ok := options["resizeKeyboard"].(bool); ok {
						keyboardOpts["resize"] = resize
					}
					if oneTime, ok := options["oneTimeKeyboard"].(bool); ok {
						keyboardOpts["oneTime"] = oneTime
					}
					params.ReplyMarkup = buildReplyKeyboard(keyboard, keyboardOpts)
				}
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
			if kb := options["inlineKeyboard"]; kb != nil {
				if keyboard := convertToKeyboardRows(kb); keyboard != nil {
					params.ReplyMarkup = buildInlineKeyboard(keyboard)
				}
			}
		}

		// Determine photo source
		photo = plugin.MustResolvePath(p.storagePath, photo)
		if plugin.IsFilePath(photo) {
			data, err := os.ReadFile(photo)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}
			params.Photo = &models.InputFileUpload{
				Filename: filepath.Base(photo),
				Data:     bytes.NewReader(data),
			}
		} else if plugin.IsBase64(photo) {
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
			if kb := options["inlineKeyboard"]; kb != nil {
				if keyboard := convertToKeyboardRows(kb); keyboard != nil {
					params.ReplyMarkup = buildInlineKeyboard(keyboard)
				}
			}
		}

		document = plugin.MustResolvePath(p.storagePath, document)
		if plugin.IsFilePath(document) {
			data, err := os.ReadFile(document)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}
			params.Document = &models.InputFileUpload{
				Filename: filepath.Base(document),
				Data:     bytes.NewReader(data),
			}
		} else if plugin.IsBase64(document) {
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
			if kb := options["inlineKeyboard"]; kb != nil {
				if keyboard := convertToKeyboardRows(kb); keyboard != nil {
					params.ReplyMarkup = buildInlineKeyboard(keyboard)
				}
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
			if kb := options["inlineKeyboard"]; kb != nil {
				if keyboard := convertToKeyboardRows(kb); keyboard != nil {
					params.ReplyMarkup = buildInlineKeyboard(keyboard)
				}
			}
		}

		video = plugin.MustResolvePath(p.storagePath, video)
		if plugin.IsFilePath(video) {
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
			if kb := options["inlineKeyboard"]; kb != nil {
				if keyboard := convertToKeyboardRows(kb); keyboard != nil {
					params.ReplyMarkup = buildInlineKeyboard(keyboard)
				}
			}
		}

		audio = plugin.MustResolvePath(p.storagePath, audio)
		if plugin.IsFilePath(audio) {
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
			if kb := options["inlineKeyboard"]; kb != nil {
				if keyboard := convertToKeyboardRows(kb); keyboard != nil {
					params.ReplyMarkup = buildInlineKeyboard(keyboard)
				}
			}
		}

		voice = plugin.MustResolvePath(p.storagePath, voice)
		if plugin.IsFilePath(voice) {
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
			if kb := options["inlineKeyboard"]; kb != nil {
				if keyboard := convertToKeyboardRows(kb); keyboard != nil {
					params.ReplyMarkup = buildInlineKeyboard(keyboard)
				}
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
		photo = plugin.MustResolvePath(p.storagePath, photo)
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
			if kb := options["inlineKeyboard"]; kb != nil {
				if keyboard := convertToKeyboardRows(kb); keyboard != nil {
					params.ReplyMarkup = buildInlineKeyboard(keyboard)
				}
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
