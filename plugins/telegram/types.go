package main

import (
	"context"
	"sync"

	"github.com/dop251/goja"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
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
	storagePath    string
	plugin         *TelegramPlugin
}

// UpdateContext provides context for handler callbacks
type UpdateContext struct {
	instance *BotInstance
	update   *models.Update
	runtime  *goja.Runtime
}
