package mail

import (
	"bytes"
	"strings"
	"testing"
)

// testMailable is a simple mailable for testing.
type testMailable struct {
	subject string
	body    string
}

func (m *testMailable) Build() *Message {
	msg := NewMessage()
	msg.SetFrom("sender@example.com", "Test Sender")
	msg.SetSubject(m.subject)
	msg.SetBody(m.body)
	return msg
}

func TestNewMessage(t *testing.T) {
	msg := NewMessage()

	if msg == nil {
		t.Fatal("expected message to be created")
	}

	if msg.Headers == nil {
		t.Error("expected headers map to be initialized")
	}

	if msg.Metadata == nil {
		t.Error("expected metadata map to be initialized")
	}
}

func TestMessageBuilder(t *testing.T) {
	msg := NewMessage()
	msg.SetFrom("sender@example.com", "Test Sender").
		AddTo("recipient@example.com", "Test Recipient").
		AddCc("cc@example.com").
		AddBcc("bcc@example.com").
		SetReplyTo("reply@example.com").
		SetSubject("Test Subject").
		SetBody("Test body").
		SetHTML("<p>Test HTML</p>").
		SetHeader("X-Custom", "value").
		AddTag("test").
		SetMetadata("key", "value")

	if len(msg.From) != 1 || msg.From[0].Address != "sender@example.com" {
		t.Error("expected from address to be set")
	}

	if msg.From[0].Name != "Test Sender" {
		t.Error("expected from name to be set")
	}

	if len(msg.To) != 1 || msg.To[0].Address != "recipient@example.com" {
		t.Error("expected to address to be set")
	}

	if len(msg.Cc) != 1 || msg.Cc[0].Address != "cc@example.com" {
		t.Error("expected cc address to be set")
	}

	if len(msg.Bcc) != 1 || msg.Bcc[0].Address != "bcc@example.com" {
		t.Error("expected bcc address to be set")
	}

	if len(msg.ReplyTo) != 1 || msg.ReplyTo[0].Address != "reply@example.com" {
		t.Error("expected reply-to address to be set")
	}

	if msg.Subject != "Test Subject" {
		t.Error("expected subject to be set")
	}

	if msg.Body != "Test body" {
		t.Error("expected body to be set")
	}

	if msg.HTMLBody != "<p>Test HTML</p>" {
		t.Error("expected HTML body to be set")
	}

	if msg.Headers["X-Custom"] != "value" {
		t.Error("expected custom header to be set")
	}

	if len(msg.Tags) != 1 || msg.Tags[0] != "test" {
		t.Error("expected tag to be set")
	}

	if msg.Metadata["key"] != "value" {
		t.Error("expected metadata to be set")
	}
}

func TestMessageAttachments(t *testing.T) {
	msg := NewMessage()
	content := []byte("test content")

	msg.Attach("test.txt", content, "text/plain")

	if len(msg.Attachments) != 1 {
		t.Fatal("expected one attachment")
	}

	att := msg.Attachments[0]
	if att.Name != "test.txt" {
		t.Error("expected attachment name to match")
	}

	if string(att.Content) != string(content) {
		t.Error("expected attachment content to match")
	}

	if att.MimeType != "text/plain" {
		t.Error("expected mime type to match")
	}
}

func TestAddressString(t *testing.T) {
	tests := []struct {
		name       string
		addr       Address
		wantPlain  string
		wantFormat string
	}{
		{
			name:       "address only",
			addr:       Address{Address: "test@example.com"},
			wantPlain:  "test@example.com",
			wantFormat: "",
		},
		{
			name:       "address with name",
			addr:       Address{Name: "Test User", Address: "test@example.com"},
			wantPlain:  "",
			wantFormat: "<%s>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.addr.String()
			if tt.wantPlain != "" && result != tt.wantPlain {
				t.Errorf("expected %q, got %q", tt.wantPlain, result)
			}
			if tt.wantFormat != "" && !strings.Contains(result, tt.addr.Address) {
				t.Errorf("expected result to contain address %q, got %q", tt.addr.Address, result)
			}
		})
	}
}

func TestPendingMail(t *testing.T) {
	transport := NewArrayTransport()
	mailer := NewMailer(MailConfig{
		FromAddress: "default@example.com",
		Transport:   transport,
	})

	mailable := &testMailable{
		subject: "Test",
		body:    "Test body",
	}

	err := mailer.To("recipient@example.com").
		Cc("cc@example.com").
		Bcc("bcc@example.com").
		Send(mailable)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	sent := transport.Sent()
	if len(sent) != 1 {
		t.Fatalf("expected 1 message, got %d", len(sent))
	}

	msg := sent[0]
	if len(msg.To) != 1 || msg.To[0].Address != "recipient@example.com" {
		t.Error("expected to address to be set")
	}

	if len(msg.Cc) != 1 || msg.Cc[0].Address != "cc@example.com" {
		t.Error("expected cc address to be set")
	}

	if len(msg.Bcc) != 1 || msg.Bcc[0].Address != "bcc@example.com" {
		t.Error("expected bcc address to be set")
	}
}

func TestLogTransport(t *testing.T) {
	var buf bytes.Buffer
	transport := NewLogTransport(&buf)

	msg := NewMessage()
	msg.SetFrom("sender@example.com").
		AddTo("recipient@example.com").
		SetSubject("Test Subject").
		SetBody("Test body")

	err := transport.Send(msg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Test Subject") {
		t.Error("expected output to contain subject")
	}

	if !strings.Contains(output, "Test body") {
		t.Error("expected output to contain body")
	}

	if !strings.Contains(output, "sender@example.com") {
		t.Error("expected output to contain from address")
	}

	if !strings.Contains(output, "recipient@example.com") {
		t.Error("expected output to contain to address")
	}
}

func TestArrayTransport(t *testing.T) {
	transport := NewArrayTransport()

	msg1 := NewMessage()
	msg1.SetFrom("sender@example.com").
		AddTo("recipient1@example.com").
		SetSubject("Message 1")

	msg2 := NewMessage()
	msg2.SetFrom("sender@example.com").
		AddTo("recipient2@example.com").
		SetSubject("Message 2")

	err := transport.Send(msg1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = transport.Send(msg2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	sent := transport.Sent()
	if len(sent) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(sent))
	}

	if sent[0].Subject != "Message 1" {
		t.Error("expected first message subject to match")
	}

	if sent[1].Subject != "Message 2" {
		t.Error("expected second message subject to match")
	}

	transport.Clear()
	if len(transport.Sent()) != 0 {
		t.Error("expected no messages after clear")
	}
}

func TestFakeMailer(t *testing.T) {
	fake := Fake()

	mailable := &testMailable{
		subject: "Test Subject",
		body:    "Test body",
	}

	err := fake.To("recipient@example.com").Send(mailable)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !fake.AssertSentCount(1) {
		t.Error("expected 1 message to be sent")
	}

	if !fake.AssertSent(func(msg *Message) bool {
		return msg.Subject == "Test Subject"
	}) {
		t.Error("expected message with subject 'Test Subject'")
	}

	if !fake.AssertSent(func(msg *Message) bool {
		return len(msg.To) > 0 && msg.To[0].Address == "recipient@example.com"
	}) {
		t.Error("expected message to recipient@example.com")
	}

	if !fake.AssertNotSent(func(msg *Message) bool {
		return msg.Subject == "Other Subject"
	}) {
		t.Error("expected no message with subject 'Other Subject'")
	}

	fake.Clear()
	if !fake.AssertNothingSent() {
		t.Error("expected no messages after clear")
	}
}

func TestMailerDefaultFromAddress(t *testing.T) {
	transport := NewArrayTransport()
	mailer := NewMailer(MailConfig{
		FromAddress: "default@example.com",
		FromName:    "Default Sender",
		Transport:   transport,
	})

	msg := NewMessage()
	msg.AddTo("recipient@example.com").
		SetSubject("Test").
		SetBody("Test body")

	err := mailer.Send(msg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	sent := transport.Sent()
	if len(sent) != 1 {
		t.Fatalf("expected 1 message, got %d", len(sent))
	}

	if len(sent[0].From) != 1 {
		t.Fatal("expected from address to be set")
	}

	if sent[0].From[0].Address != "default@example.com" {
		t.Error("expected default from address")
	}

	if sent[0].From[0].Name != "Default Sender" {
		t.Error("expected default from name")
	}
}
