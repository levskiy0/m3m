package main

import (
	"github.com/go-telegram/bot/models"
)

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
