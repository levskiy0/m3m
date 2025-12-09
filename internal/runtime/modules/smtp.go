package modules

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/dop251/goja"
	"github.com/levskiy0/m3m/pkg/schema"
)

type SMTPModule struct {
	envModule *EnvModule
}

// Name returns the module name for JavaScript
func (m *SMTPModule) Name() string {
	return "$smtp"
}

// Register registers the module into the JavaScript VM
func (m *SMTPModule) Register(vm interface{}) {
	vm.(*goja.Runtime).Set(m.Name(), map[string]interface{}{
		"send":     m.Send,
		"sendHTML": m.SendHTML,
	})
}

type EmailOptions struct {
	From        string            `json:"from"`
	ReplyTo     string            `json:"reply_to"`
	CC          []string          `json:"cc"`
	BCC         []string          `json:"bcc"`
	Headers     map[string]string `json:"headers"`
	ContentType string            `json:"content_type"` // "text/plain" or "text/html"
}

type SMTPResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func NewSMTPModule(envModule *EnvModule) *SMTPModule {
	return &SMTPModule{
		envModule: envModule,
	}
}

// Send sends an email using SMTP configuration from environment
// Required env vars: SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, SMTP_FROM
func (m *SMTPModule) Send(to, subject, body string, options *EmailOptions) *SMTPResult {
	// Get SMTP configuration from environment
	host := m.getEnvString("SMTP_HOST")
	port := m.getEnvString("SMTP_PORT")
	user := m.getEnvString("SMTP_USER")
	pass := m.getEnvString("SMTP_PASS")
	defaultFrom := m.getEnvString("SMTP_FROM")

	if host == "" || port == "" {
		return &SMTPResult{
			Success: false,
			Error:   "SMTP_HOST and SMTP_PORT environment variables are required",
		}
	}

	// Determine sender
	from := defaultFrom
	if options != nil && options.From != "" {
		from = options.From
	}
	if from == "" {
		return &SMTPResult{
			Success: false,
			Error:   "sender address is required (SMTP_FROM env or options.from)",
		}
	}

	// Build recipients list
	recipients := []string{to}
	if options != nil {
		recipients = append(recipients, options.CC...)
		recipients = append(recipients, options.BCC...)
	}

	// Build email message
	message := m.buildMessage(from, to, subject, body, options)

	// Send email
	addr := fmt.Sprintf("%s:%s", host, port)
	var auth smtp.Auth
	if user != "" && pass != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}

	var err error
	if port == "465" {
		// SSL/TLS connection
		err = m.sendTLS(addr, auth, from, recipients, message)
	} else {
		// STARTTLS or plain
		err = smtp.SendMail(addr, auth, from, recipients, message)
	}

	if err != nil {
		return &SMTPResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	return &SMTPResult{Success: true}
}

// SendHTML is a convenience method for sending HTML emails
func (m *SMTPModule) SendHTML(to, subject, body string) *SMTPResult {
	return m.Send(to, subject, body, &EmailOptions{ContentType: "text/html"})
}

func (m *SMTPModule) buildMessage(from, to, subject, body string, options *EmailOptions) []byte {
	var msg strings.Builder

	// Headers
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))

	if options != nil {
		if options.ReplyTo != "" {
			msg.WriteString(fmt.Sprintf("Reply-To: %s\r\n", options.ReplyTo))
		}
		if len(options.CC) > 0 {
			msg.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(options.CC, ", ")))
		}
		for key, value := range options.Headers {
			msg.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
		}

		contentType := options.ContentType
		if contentType == "" {
			contentType = "text/plain"
		}
		msg.WriteString(fmt.Sprintf("Content-Type: %s; charset=UTF-8\r\n", contentType))
	} else {
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}

	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)

	return []byte(msg.String())
}

func (m *SMTPModule) sendTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         strings.Split(addr, ":")[0],
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS dial error: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, strings.Split(addr, ":")[0])
	if err != nil {
		return fmt.Errorf("SMTP client error: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth error: %w", err)
		}
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL error: %w", err)
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("SMTP RCPT error: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA error: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("SMTP write error: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("SMTP close error: %w", err)
	}

	return client.Quit()
}

func (m *SMTPModule) getEnvString(key string) string {
	val := m.envModule.Get(key)
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

// GetSchema implements JSSchemaProvider
func (m *SMTPModule) GetSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$smtp",
		Description: "Email sending via SMTP (requires SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, SMTP_FROM env vars)",
		Types: []schema.TypeSchema{
			{
				Name:        "EmailOptions",
				Description: "Options for sending email",
				Fields: []schema.ParamSchema{
					{Name: "from", Type: "string", Description: "Sender email address (overrides SMTP_FROM)", Optional: true},
					{Name: "reply_to", Type: "string", Description: "Reply-to address", Optional: true},
					{Name: "cc", Type: "string[]", Description: "CC recipients", Optional: true},
					{Name: "bcc", Type: "string[]", Description: "BCC recipients", Optional: true},
					{Name: "headers", Type: "{ [key: string]: string }", Description: "Custom headers", Optional: true},
					{Name: "content_type", Type: "string", Description: "Content type (text/plain or text/html)", Optional: true},
				},
			},
			{
				Name:        "SMTPResult",
				Description: "Result of sending email",
				Fields: []schema.ParamSchema{
					{Name: "success", Type: "boolean", Description: "Whether the email was sent successfully"},
					{Name: "error", Type: "string", Description: "Error message if failed", Optional: true},
				},
			},
		},
		Methods: []schema.MethodSchema{
			{
				Name:        "send",
				Description: "Send an email",
				Params: []schema.ParamSchema{
					{Name: "to", Type: "string", Description: "Recipient email address"},
					{Name: "subject", Type: "string", Description: "Email subject"},
					{Name: "body", Type: "string", Description: "Email body content"},
					{Name: "options", Type: "EmailOptions", Description: "Additional email options", Optional: true},
				},
				Returns: &schema.ParamSchema{Type: "SMTPResult"},
			},
			{
				Name:        "sendHTML",
				Description: "Send an HTML email",
				Params: []schema.ParamSchema{
					{Name: "to", Type: "string", Description: "Recipient email address"},
					{Name: "subject", Type: "string", Description: "Email subject"},
					{Name: "body", Type: "string", Description: "HTML email body"},
				},
				Returns: &schema.ParamSchema{Type: "SMTPResult"},
			},
		},
	}
}

// GetSMTPSchema returns the smtp schema (static version)
func GetSMTPSchema() schema.ModuleSchema {
	return (&SMTPModule{}).GetSchema()
}
