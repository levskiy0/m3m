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
	// Forward origin (new API)
	if m.ForwardOrigin != nil {
		origin := map[string]interface{}{
			"type": string(m.ForwardOrigin.Type),
		}
		switch m.ForwardOrigin.Type {
		case "user":
			if m.ForwardOrigin.MessageOriginUser != nil {
				origin["date"] = m.ForwardOrigin.MessageOriginUser.Date
				origin["senderUser"] = uctx.convertUser(&m.ForwardOrigin.MessageOriginUser.SenderUser)
			}
		case "hidden_user":
			if m.ForwardOrigin.MessageOriginHiddenUser != nil {
				origin["date"] = m.ForwardOrigin.MessageOriginHiddenUser.Date
				origin["senderUserName"] = m.ForwardOrigin.MessageOriginHiddenUser.SenderUserName
			}
		case "chat":
			if m.ForwardOrigin.MessageOriginChat != nil {
				origin["date"] = m.ForwardOrigin.MessageOriginChat.Date
				origin["senderChat"] = uctx.convertChat(m.ForwardOrigin.MessageOriginChat.SenderChat)
				if m.ForwardOrigin.MessageOriginChat.AuthorSignature != nil {
					origin["authorSignature"] = *m.ForwardOrigin.MessageOriginChat.AuthorSignature
				}
			}
		case "channel":
			if m.ForwardOrigin.MessageOriginChannel != nil {
				origin["date"] = m.ForwardOrigin.MessageOriginChannel.Date
				origin["chat"] = uctx.convertChat(m.ForwardOrigin.MessageOriginChannel.Chat)
				origin["messageId"] = m.ForwardOrigin.MessageOriginChannel.MessageID
				if m.ForwardOrigin.MessageOriginChannel.AuthorSignature != nil {
					origin["authorSignature"] = *m.ForwardOrigin.MessageOriginChannel.AuthorSignature
				}
			}
		}
		msg["forwardOrigin"] = origin
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
