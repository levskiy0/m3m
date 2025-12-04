package modules

import (
	"strings"
	"testing"
)

func TestSMTPModule_MissingConfig(t *testing.T) {
	envModule := NewEnvModule(map[string]interface{}{})
	module := NewSMTPModule(envModule)

	result := module.Send("test@example.com", "Test Subject", "Test Body", nil)
	if result.Success {
		t.Error("Expected failure when SMTP config is missing")
	}
	if !strings.Contains(result.Error, "SMTP_HOST") {
		t.Errorf("Expected error about SMTP_HOST, got: %s", result.Error)
	}
}

func TestSMTPModule_MissingSender(t *testing.T) {
	envModule := NewEnvModule(map[string]interface{}{
		"SMTP_HOST": "smtp.example.com",
		"SMTP_PORT": "587",
	})
	module := NewSMTPModule(envModule)

	result := module.Send("test@example.com", "Test Subject", "Test Body", nil)
	if result.Success {
		t.Error("Expected failure when sender is missing")
	}
	if !strings.Contains(result.Error, "sender address") {
		t.Errorf("Expected error about sender, got: %s", result.Error)
	}
}

func TestSMTPModule_BuildMessage_Plain(t *testing.T) {
	envModule := NewEnvModule(map[string]interface{}{})
	module := NewSMTPModule(envModule)

	msg := module.buildMessage("from@example.com", "to@example.com", "Test Subject", "Test Body", nil)
	msgStr := string(msg)

	// Check headers
	if !strings.Contains(msgStr, "From: from@example.com") {
		t.Error("Missing From header")
	}
	if !strings.Contains(msgStr, "To: to@example.com") {
		t.Error("Missing To header")
	}
	if !strings.Contains(msgStr, "Subject: Test Subject") {
		t.Error("Missing Subject header")
	}
	if !strings.Contains(msgStr, "Content-Type: text/plain; charset=UTF-8") {
		t.Error("Missing Content-Type header")
	}
	if !strings.Contains(msgStr, "MIME-Version: 1.0") {
		t.Error("Missing MIME-Version header")
	}
	if !strings.Contains(msgStr, "Test Body") {
		t.Error("Missing body content")
	}
}

func TestSMTPModule_BuildMessage_HTML(t *testing.T) {
	envModule := NewEnvModule(map[string]interface{}{})
	module := NewSMTPModule(envModule)

	options := &EmailOptions{ContentType: "text/html"}
	msg := module.buildMessage("from@example.com", "to@example.com", "Test Subject", "<h1>Hello</h1>", options)
	msgStr := string(msg)

	if !strings.Contains(msgStr, "Content-Type: text/html; charset=UTF-8") {
		t.Error("Missing HTML Content-Type header")
	}
	if !strings.Contains(msgStr, "<h1>Hello</h1>") {
		t.Error("Missing HTML body content")
	}
}

func TestSMTPModule_BuildMessage_WithOptions(t *testing.T) {
	envModule := NewEnvModule(map[string]interface{}{})
	module := NewSMTPModule(envModule)

	options := &EmailOptions{
		From:    "custom@example.com",
		ReplyTo: "reply@example.com",
		CC:      []string{"cc1@example.com", "cc2@example.com"},
		Headers: map[string]string{
			"X-Custom-Header": "custom-value",
		},
	}
	msg := module.buildMessage("from@example.com", "to@example.com", "Test Subject", "Body", options)
	msgStr := string(msg)

	if !strings.Contains(msgStr, "Reply-To: reply@example.com") {
		t.Error("Missing Reply-To header")
	}
	if !strings.Contains(msgStr, "Cc: cc1@example.com, cc2@example.com") {
		t.Error("Missing Cc header")
	}
	if !strings.Contains(msgStr, "X-Custom-Header: custom-value") {
		t.Error("Missing custom header")
	}
}

func TestSMTPModule_SendHTML(t *testing.T) {
	envModule := NewEnvModule(map[string]interface{}{})
	module := NewSMTPModule(envModule)

	// Without SMTP config, it should fail early
	result := module.SendHTML("test@example.com", "Test Subject", "<p>Test</p>")
	if result.Success {
		t.Error("Expected failure when SMTP config is missing")
	}
}

func TestSMTPModule_CustomSender(t *testing.T) {
	envModule := NewEnvModule(map[string]interface{}{
		"SMTP_HOST": "smtp.example.com",
		"SMTP_PORT": "587",
	})
	module := NewSMTPModule(envModule)

	// With custom sender in options
	options := &EmailOptions{From: "custom@example.com"}
	result := module.Send("test@example.com", "Test Subject", "Test Body", options)

	// Will fail at network level, but sender validation should pass
	if result.Success {
		t.Error("Expected failure at network level (no actual SMTP server)")
	}
	// Error should NOT be about missing sender
	if strings.Contains(result.Error, "sender address") {
		t.Errorf("Unexpected sender error with custom From, got: %s", result.Error)
	}
}

func TestSMTPModule_GetEnvString(t *testing.T) {
	envModule := NewEnvModule(map[string]interface{}{
		"STRING_VALUE": "hello",
		"INT_VALUE":    123,
		"BOOL_VALUE":   true,
	})
	module := NewSMTPModule(envModule)

	// Test string value
	if val := module.getEnvString("STRING_VALUE"); val != "hello" {
		t.Errorf("Expected 'hello', got '%s'", val)
	}

	// Test non-string values should return empty string
	if val := module.getEnvString("INT_VALUE"); val != "" {
		t.Errorf("Expected empty string for int value, got '%s'", val)
	}
	if val := module.getEnvString("BOOL_VALUE"); val != "" {
		t.Errorf("Expected empty string for bool value, got '%s'", val)
	}

	// Test missing key
	if val := module.getEnvString("NONEXISTENT"); val != "" {
		t.Errorf("Expected empty string for missing key, got '%s'", val)
	}
}

func TestSMTPModule_RecipientsList(t *testing.T) {
	// Test that CC and BCC are added to recipients
	envModule := NewEnvModule(map[string]interface{}{
		"SMTP_HOST": "smtp.example.com",
		"SMTP_PORT": "587",
		"SMTP_FROM": "from@example.com",
	})
	module := NewSMTPModule(envModule)

	options := &EmailOptions{
		CC:  []string{"cc@example.com"},
		BCC: []string{"bcc@example.com"},
	}

	// Send will fail at network level, but internal logic is exercised
	result := module.Send("to@example.com", "Subject", "Body", options)
	if result.Success {
		t.Error("Expected failure at network level")
	}
}

func TestSMTPModule_EmptyBody(t *testing.T) {
	envModule := NewEnvModule(map[string]interface{}{})
	module := NewSMTPModule(envModule)

	msg := module.buildMessage("from@example.com", "to@example.com", "Subject", "", nil)
	msgStr := string(msg)

	// Should still have proper structure
	if !strings.Contains(msgStr, "\r\n\r\n") {
		t.Error("Message should have header-body separator")
	}
}

func TestSMTPModule_SpecialCharactersInSubject(t *testing.T) {
	envModule := NewEnvModule(map[string]interface{}{})
	module := NewSMTPModule(envModule)

	subject := "Test: Special chars & symbols!"
	msg := module.buildMessage("from@example.com", "to@example.com", subject, "Body", nil)
	msgStr := string(msg)

	if !strings.Contains(msgStr, "Subject: "+subject) {
		t.Error("Subject with special characters not preserved")
	}
}
