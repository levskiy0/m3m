// Telegram plugin for M3M
// Build with: go build -buildmode=plugin -o ../telegram.so
package main

import (
	"context"
	"fmt"

	"github.com/dop251/goja"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (p *TelegramPlugin) Name() string {
	return "$telegram"
}

func (p *TelegramPlugin) Version() string {
	return "1.0.1"
}

func (p *TelegramPlugin) Description() string {
	return "Telegram Bot API plugin for building Telegram bots"
}

func (p *TelegramPlugin) Author() string {
	return "M3M Team"
}

func (p *TelegramPlugin) URL() string {
	return "https://github.com/m3m/plugins/telegram"
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
		ctx:         ctx,
		cancel:      cancel,
		runtime:     runtime,
		handlers:    make(map[string]goja.Callable),
		callbacks:   make(map[string]goja.Callable),
		storagePath: p.storagePath,
		plugin:      p,
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

// NewPlugin is the exported function that returns a new plugin instance
func NewPlugin() interface{} {
	return &TelegramPlugin{}
}
