package main

import (
	"github.com/go-telegram/bot/models"

	"m3m/pkg/plugin"
)

// convertToKeyboardRows converts GOJA array to typed keyboard rows
func convertToKeyboardRows(value interface{}) [][]map[string]interface{} {
	rows, ok := value.([]interface{})
	if !ok {
		// Try direct type (might work in some cases)
		if typed, ok := value.([][]map[string]interface{}); ok {
			return typed
		}
		return nil
	}

	result := make([][]map[string]interface{}, len(rows))
	for i, row := range rows {
		rowSlice, ok := row.([]interface{})
		if !ok {
			continue
		}
		result[i] = make([]map[string]interface{}, len(rowSlice))
		for j, btn := range rowSlice {
			if btnMap, ok := btn.(map[string]interface{}); ok {
				result[i][j] = btnMap
			}
		}
	}
	return result
}

// buildInlineKeyboard builds an inline keyboard markup from JS array
func buildInlineKeyboard(keyboard [][]map[string]interface{}) *models.InlineKeyboardMarkup {
	rows := make([][]models.InlineKeyboardButton, len(keyboard))
	for i, row := range keyboard {
		buttons := make([]models.InlineKeyboardButton, len(row))
		for j, btn := range row {
			button := models.InlineKeyboardButton{
				Text: plugin.GetString(btn, "text"),
			}
			if url := plugin.GetString(btn, "url"); url != "" {
				button.URL = url
			}
			if data := plugin.GetString(btn, "callbackData"); data != "" {
				button.CallbackData = data
			} else if data := plugin.GetString(btn, "callback_data"); data != "" {
				button.CallbackData = data
			}
			buttons[j] = button
		}
		rows[i] = buttons
	}
	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// buildReplyKeyboard builds a reply keyboard markup from JS array
func buildReplyKeyboard(keyboard [][]map[string]interface{}, options map[string]interface{}) *models.ReplyKeyboardMarkup {
	rows := make([][]models.KeyboardButton, len(keyboard))
	for i, row := range keyboard {
		buttons := make([]models.KeyboardButton, len(row))
		for j, btn := range row {
			button := models.KeyboardButton{
				Text: plugin.GetString(btn, "text"),
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
