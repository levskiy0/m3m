package modules

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/dop251/goja"
	"github.com/jordan-wright/email"
	"github.com/levskiy0/m3m/pkg/schema"
)

type MailModule struct {
	envModule *EnvModule
}

// Name returns the module name for JavaScript
func (m *MailModule) Name() string {
	return "$mail"
}

// Register registers the module into the JavaScript VM
func (m *MailModule) Register(vm interface{}) {
	vm.(*goja.Runtime).Set(m.Name(), map[string]interface{}{
		"send":     m.Send,
		"sendHTML": m.SendHTML,
	})
}

type MailOptions struct {
	From        string            `json:"from"`
	ReplyTo     []string          `json:"reply_to"`
	CC          []string          `json:"cc"`
	BCC         []string          `json:"bcc"`
	Headers     map[string]string `json:"headers"`
	Attachments []MailAttachment  `json:"attachments"`
}

type MailAttachment struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`      // base64 encoded content
	ContentType string `json:"content_type"` // MIME type
}

type MailResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func NewMailModule(envModule *EnvModule) *MailModule {
	return &MailModule{
		envModule: envModule,
	}
}

// Send sends a plain text email using SMTP configuration from environment
// Required env vars: SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, SMTP_FROM
func (m *MailModule) Send(to, subject, body string, options *MailOptions) *MailResult {
	return m.sendEmail(to, subject, body, "", options)
}

// SendHTML sends an HTML email using SMTP configuration from environment
func (m *MailModule) SendHTML(to, subject, htmlBody string, options *MailOptions) *MailResult {
	return m.sendEmail(to, subject, "", htmlBody, options)
}

func (m *MailModule) sendEmail(to, subject, textBody, htmlBody string, options *MailOptions) *MailResult {
	// Get SMTP configuration from environment
	host := m.getEnvString("SMTP_HOST")
	port := m.getEnvString("SMTP_PORT")
	user := m.getEnvString("SMTP_USER")
	pass := m.getEnvString("SMTP_PASS")
	defaultFrom := m.getEnvString("SMTP_FROM")

	if host == "" || port == "" {
		return &MailResult{
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
		return &MailResult{
			Success: false,
			Error:   "sender address is required (SMTP_FROM env or options.from)",
		}
	}

	// Create email
	e := email.NewEmail()
	e.From = from
	e.To = []string{to}
	e.Subject = subject

	if textBody != "" {
		e.Text = []byte(textBody)
	}
	if htmlBody != "" {
		e.HTML = []byte(htmlBody)
	}

	if options != nil {
		if len(options.ReplyTo) > 0 {
			e.ReplyTo = options.ReplyTo
		}
		if len(options.CC) > 0 {
			e.Cc = options.CC
		}
		if len(options.BCC) > 0 {
			e.Bcc = options.BCC
		}
		for key, value := range options.Headers {
			e.Headers.Add(key, value)
		}

		// Handle attachments
		for _, att := range options.Attachments {
			content, err := decodeBase64(att.Content)
			if err != nil {
				return &MailResult{
					Success: false,
					Error:   fmt.Sprintf("failed to decode attachment %s: %v", att.Filename, err),
				}
			}

			contentType := att.ContentType
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			_, err = e.Attach(strings.NewReader(string(content)), att.Filename, contentType)
			if err != nil {
				return &MailResult{
					Success: false,
					Error:   fmt.Sprintf("failed to attach %s: %v", att.Filename, err),
				}
			}
		}
	}

	// Send email
	addr := fmt.Sprintf("%s:%s", host, port)
	var auth smtp.Auth
	if user != "" && pass != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}

	var err error
	if port == "465" {
		// SSL/TLS connection
		tlsConfig := &tls.Config{
			ServerName: host,
		}
		err = e.SendWithTLS(addr, auth, tlsConfig)
	} else {
		// STARTTLS or plain
		err = e.Send(addr, auth)
	}

	if err != nil {
		return &MailResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	return &MailResult{Success: true}
}

func (m *MailModule) getEnvString(key string) string {
	val := m.envModule.Get(key)
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

func decodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// GetSchema implements JSSchemaProvider
func (m *MailModule) GetSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$mail",
		Description: "Email sending via SMTP (requires SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, SMTP_FROM env vars)",
		Types: []schema.TypeSchema{
			{
				Name:        "MailAttachment",
				Description: "Email attachment",
				Fields: []schema.ParamSchema{
					{Name: "filename", Type: "string", Description: "Attachment filename"},
					{Name: "content", Type: "string", Description: "Base64 encoded content"},
					{Name: "content_type", Type: "string", Description: "MIME type (default: application/octet-stream)", Optional: true},
				},
			},
			{
				Name:        "MailOptions",
				Description: "Options for sending email",
				Fields: []schema.ParamSchema{
					{Name: "from", Type: "string", Description: "Sender email address (overrides SMTP_FROM)", Optional: true},
					{Name: "reply_to", Type: "string[]", Description: "Reply-to addresses", Optional: true},
					{Name: "cc", Type: "string[]", Description: "CC recipients", Optional: true},
					{Name: "bcc", Type: "string[]", Description: "BCC recipients", Optional: true},
					{Name: "headers", Type: "{ [key: string]: string }", Description: "Custom headers", Optional: true},
					{Name: "attachments", Type: "MailAttachment[]", Description: "File attachments", Optional: true},
				},
			},
			{
				Name:        "MailResult",
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
				Description: "Send a plain text email",
				Params: []schema.ParamSchema{
					{Name: "to", Type: "string", Description: "Recipient email address"},
					{Name: "subject", Type: "string", Description: "Email subject"},
					{Name: "body", Type: "string", Description: "Plain text email body"},
					{Name: "options", Type: "MailOptions", Description: "Additional email options", Optional: true},
				},
				Returns: &schema.ParamSchema{Type: "MailResult"},
			},
			{
				Name:        "sendHTML",
				Description: "Send an HTML email",
				Params: []schema.ParamSchema{
					{Name: "to", Type: "string", Description: "Recipient email address"},
					{Name: "subject", Type: "string", Description: "Email subject"},
					{Name: "htmlBody", Type: "string", Description: "HTML email body"},
					{Name: "options", Type: "MailOptions", Description: "Additional email options", Optional: true},
				},
				Returns: &schema.ParamSchema{Type: "MailResult"},
			},
		},
	}
}

// GetMailSchema returns the mail schema (static version)
func GetMailSchema() schema.ModuleSchema {
	return (&MailModule{}).GetSchema()
}
