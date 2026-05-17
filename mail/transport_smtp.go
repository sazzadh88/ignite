package mail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

// SMTPConfig holds SMTP transport configuration.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	UseTLS   bool
}

// SMTPTransport sends emails via SMTP using net/smtp.
type SMTPTransport struct {
	config SMTPConfig
}

// NewSMTPTransport creates a new SMTP transport with the given configuration.
func NewSMTPTransport(config SMTPConfig) *SMTPTransport {
	return &SMTPTransport{
		config: config,
	}
}

// Send delivers the message via SMTP.
func (t *SMTPTransport) Send(message *Message) error {
	if len(message.To) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	if len(message.From) == 0 {
		return fmt.Errorf("no sender specified")
	}

	// Build MIME message
	body := t.buildMIMEMessage(message)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", t.config.Host, t.config.Port)
	auth := smtp.PlainAuth("", t.config.Username, t.config.Password, t.config.Host)

	// Collect all recipients
	recipients := make([]string, 0)
	for _, addr := range message.To {
		recipients = append(recipients, addr.Address)
	}
	for _, addr := range message.Cc {
		recipients = append(recipients, addr.Address)
	}
	for _, addr := range message.Bcc {
		recipients = append(recipients, addr.Address)
	}

	// Send the email
	return smtp.SendMail(addr, auth, message.From[0].Address, recipients, []byte(body))
}

// buildMIMEMessage constructs a MIME-formatted email message.
func (t *SMTPTransport) buildMIMEMessage(message *Message) string {
	var buf bytes.Buffer

	// Headers
	buf.WriteString(fmt.Sprintf("From: %s\r\n", message.From[0].String()))

	if len(message.To) > 0 {
		toAddrs := make([]string, len(message.To))
		for i, addr := range message.To {
			toAddrs[i] = addr.String()
		}
		buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(toAddrs, ", ")))
	}

	if len(message.Cc) > 0 {
		ccAddrs := make([]string, len(message.Cc))
		for i, addr := range message.Cc {
			ccAddrs[i] = addr.String()
		}
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(ccAddrs, ", ")))
	}

	if len(message.ReplyTo) > 0 {
		buf.WriteString(fmt.Sprintf("Reply-To: %s\r\n", message.ReplyTo[0].String()))
	}

	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", message.Subject))
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	buf.WriteString("MIME-Version: 1.0\r\n")

	// Custom headers
	for key, value := range message.Headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	// Determine content structure
	hasAttachments := len(message.Attachments) > 0
	hasHTML := message.HTMLBody != ""
	hasText := message.Body != ""

	if hasAttachments {
		boundary := generateBoundary()
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", boundary))

		// Message body part
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		if hasHTML && hasText {
			altBoundary := generateBoundary()
			buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", altBoundary))

			// Plain text
			buf.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
			buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
			buf.WriteString(message.Body)
			buf.WriteString("\r\n")

			// HTML
			buf.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
			buf.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
			buf.WriteString(message.HTMLBody)
			buf.WriteString("\r\n")

			buf.WriteString(fmt.Sprintf("--%s--\r\n", altBoundary))
		} else if hasHTML {
			buf.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
			buf.WriteString(message.HTMLBody)
			buf.WriteString("\r\n")
		} else {
			buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
			buf.WriteString(message.Body)
			buf.WriteString("\r\n")
		}

		// Attachments
		for _, att := range message.Attachments {
			buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			buf.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", att.MimeType, att.Name))
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", att.Name))

			encoded := base64.StdEncoding.EncodeToString(att.Content)
			for i := 0; i < len(encoded); i += 76 {
				end := i + 76
				if end > len(encoded) {
					end = len(encoded)
				}
				buf.WriteString(encoded[i:end])
				buf.WriteString("\r\n")
			}
		}

		buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if hasHTML && hasText {
		boundary := generateBoundary()
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundary))

		// Plain text
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
		buf.WriteString(message.Body)
		buf.WriteString("\r\n")

		// HTML
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
		buf.WriteString(message.HTMLBody)
		buf.WriteString("\r\n")

		buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else if hasHTML {
		buf.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
		buf.WriteString(message.HTMLBody)
		buf.WriteString("\r\n")
	} else {
		buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
		buf.WriteString(message.Body)
		buf.WriteString("\r\n")
	}

	return buf.String()
}

// generateBoundary generates a MIME boundary string.
func generateBoundary() string {
	return fmt.Sprintf("boundary_%d", time.Now().UnixNano())
}
