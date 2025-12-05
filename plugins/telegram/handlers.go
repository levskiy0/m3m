package main

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/dop251/goja"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

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

// callHandler safely calls a JavaScript handler with panic recovery
func (instance *BotInstance) callHandler(handler goja.Callable, uctx *UpdateContext) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[ERROR] Handler panic: %v\n%s\n", r, debug.Stack())
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

// Handler registration methods
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

func (uctx *UpdateContext) getChatID() int64 {
	if uctx.update.Message != nil {
		return uctx.update.Message.Chat.ID
	}
	if uctx.update.CallbackQuery != nil && uctx.update.CallbackQuery.Message.Message != nil {
		return uctx.update.CallbackQuery.Message.Message.Chat.ID
	}
	return 0
}
