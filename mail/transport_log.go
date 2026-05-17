package mail

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// LogTransport writes email messages to a writer (typically for development).
type LogTransport struct {
	writer io.Writer
}

// NewLogTransport creates a new log transport that writes to the given writer.
// If writer is nil, it defaults to os.Stdout.
func NewLogTransport(writer io.Writer) *LogTransport {
	if writer == nil {
		writer = os.Stdout
	}
	return &LogTransport{
		writer: writer,
	}
}

// Send logs the message details to the configured writer.
func (t *LogTransport) Send(message *Message) error {
	var buf strings.Builder

	buf.WriteString("========================================\n")
	buf.WriteString("EMAIL MESSAGE\n")
	buf.WriteString("========================================\n")

	if len(message.From) > 0 {
		buf.WriteString(fmt.Sprintf("From: %s\n", message.From[0].String()))
	}

	if len(message.To) > 0 {
		toAddrs := make([]string, len(message.To))
		for i, addr := range message.To {
			toAddrs[i] = addr.String()
		}
		buf.WriteString(fmt.Sprintf("To: %s\n", strings.Join(toAddrs, ", ")))
	}

	if len(message.Cc) > 0 {
		ccAddrs := make([]string, len(message.Cc))
		for i, addr := range message.Cc {
			ccAddrs[i] = addr.String()
		}
		buf.WriteString(fmt.Sprintf("Cc: %s\n", strings.Join(ccAddrs, ", ")))
	}

	if len(message.Bcc) > 0 {
		bccAddrs := make([]string, len(message.Bcc))
		for i, addr := range message.Bcc {
			bccAddrs[i] = addr.String()
		}
		buf.WriteString(fmt.Sprintf("Bcc: %s\n", strings.Join(bccAddrs, ", ")))
	}

	if len(message.ReplyTo) > 0 {
		buf.WriteString(fmt.Sprintf("Reply-To: %s\n", message.ReplyTo[0].String()))
	}

	buf.WriteString(fmt.Sprintf("Subject: %s\n", message.Subject))

	if len(message.Headers) > 0 {
		buf.WriteString("\nCustom Headers:\n")
		for key, value := range message.Headers {
			buf.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	if len(message.Tags) > 0 {
		buf.WriteString(fmt.Sprintf("\nTags: %s\n", strings.Join(message.Tags, ", ")))
	}

	if len(message.Metadata) > 0 {
		buf.WriteString("\nMetadata:\n")
		for key, value := range message.Metadata {
			buf.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	if message.Body != "" {
		buf.WriteString("\n--- Plain Text ---\n")
		buf.WriteString(message.Body)
		buf.WriteString("\n")
	}

	if message.HTMLBody != "" {
		buf.WriteString("\n--- HTML ---\n")
		buf.WriteString(message.HTMLBody)
		buf.WriteString("\n")
	}

	if len(message.Attachments) > 0 {
		buf.WriteString("\nAttachments:\n")
		for _, att := range message.Attachments {
			buf.WriteString(fmt.Sprintf("  - %s (%s, %d bytes)\n", att.Name, att.MimeType, len(att.Content)))
		}
	}

	buf.WriteString("========================================\n\n")

	_, err := t.writer.Write([]byte(buf.String()))
	return err
}
