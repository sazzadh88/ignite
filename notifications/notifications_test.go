package notifications

import (
	"testing"
	"time"

	"github.com/sazzadh88/ignite/mail"
)

// testUser implements Notifiable for testing
type testUser struct {
	email string
	id    string
}

func (u *testUser) NotifyVia() []string {
	return []string{"mail", "database"}
}

func (u *testUser) RouteNotificationFor(channel string) string {
	switch channel {
	case "mail":
		return u.email
	case "database":
		return u.id
	default:
		return ""
	}
}

// testNotification implements Notification, ToMail, and ToDatabase
type testNotification struct {
	subject string
	message string
}

func (n *testNotification) Via(notifiable any) []string {
	if user, ok := notifiable.(*testUser); ok {
		return user.NotifyVia()
	}
	return []string{"mail"}
}

func (n *testNotification) ToMail(notifiable any) *mail.Message {
	msg := mail.NewMessage()
	msg.SetFrom("noreply@example.com")
	msg.SetSubject(n.subject)
	msg.SetBody(n.message)
	return msg
}

func (n *testNotification) ToDatabase(notifiable any) map[string]any {
	return map[string]any{
		"subject": n.subject,
		"message": n.message,
	}
}

func TestMailChannel(t *testing.T) {
	transport := mail.NewArrayTransport()
	mailer := mail.NewMailer(mail.MailConfig{
		FromAddress: "system@example.com",
		Transport:   transport,
	})

	channel := NewMailChannel(mailer)
	user := &testUser{email: "user@example.com", id: "1"}
	notification := &testNotification{
		subject: "Test Subject",
		message: "Test message",
	}

	err := channel.Send(user, notification)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	sent := transport.Sent()
	if len(sent) != 1 {
		t.Fatalf("expected 1 message, got %d", len(sent))
	}

	msg := sent[0]
	if msg.Subject != "Test Subject" {
		t.Errorf("expected subject %q, got %q", "Test Subject", msg.Subject)
	}

	if msg.Body != "Test message" {
		t.Errorf("expected body %q, got %q", "Test message", msg.Body)
	}

	if len(msg.To) == 0 || msg.To[0].Address != "user@example.com" {
		t.Error("expected message to be sent to user@example.com")
	}
}

func TestDatabaseChannel(t *testing.T) {
	var stored *DatabaseNotification

	storeFunc := func(n *DatabaseNotification) error {
		stored = n
		return nil
	}

	channel := NewDatabaseChannel(storeFunc)
	user := &testUser{email: "user@example.com", id: "1"}
	notification := &testNotification{
		subject: "Test Subject",
		message: "Test message",
	}

	err := channel.Send(user, notification)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if stored == nil {
		t.Fatal("expected notification to be stored")
	}

	if stored.NotifiableID != "1" {
		t.Errorf("expected notifiable ID %q, got %q", "1", stored.NotifiableID)
	}

	if stored.Data["subject"] != "Test Subject" {
		t.Errorf("expected subject in data, got %v", stored.Data)
	}
}

func TestSenderMultipleChannels(t *testing.T) {
	// Setup mail channel
	transport := mail.NewArrayTransport()
	mailer := mail.NewMailer(mail.MailConfig{
		FromAddress: "system@example.com",
		Transport:   transport,
	})
	mailChannel := NewMailChannel(mailer)

	// Setup database channel
	var stored *DatabaseNotification
	storeFunc := func(n *DatabaseNotification) error {
		stored = n
		return nil
	}
	dbChannel := NewDatabaseChannel(storeFunc)

	// Setup sender
	sender := NewSender()
	sender.Via("mail", mailChannel)
	sender.Via("database", dbChannel)

	// Send notification
	user := &testUser{email: "user@example.com", id: "1"}
	notification := &testNotification{
		subject: "Test Subject",
		message: "Test message",
	}

	err := sender.SendNow(user, notification)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify mail was sent
	sent := transport.Sent()
	if len(sent) != 1 {
		t.Errorf("expected 1 email, got %d", len(sent))
	}

	// Verify database was updated
	if stored == nil {
		t.Error("expected notification to be stored in database")
	}
}

func TestFakeSender(t *testing.T) {
	fake := Fake()
	user := &testUser{email: "user@example.com", id: "1"}
	notification := &testNotification{
		subject: "Test Subject",
		message: "Test message",
	}

	err := fake.Send(user, notification)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !fake.AssertCount(1) {
		t.Error("expected 1 notification to be sent")
	}

	if !fake.AssertSentTo(user, &testNotification{}) {
		t.Error("expected notification to be sent to user")
	}

	anotherUser := &testUser{email: "other@example.com", id: "2"}
	if !fake.AssertNotSentTo(anotherUser, &testNotification{}) {
		t.Error("expected notification not to be sent to another user")
	}

	fake.Clear()
	if !fake.AssertNothingSent() {
		t.Error("expected no notifications after clear")
	}
}

func TestSlackMessage(t *testing.T) {
	msg := NewSlackMessage()
	msg.SetContent("Test notification").
		AddAttachment(func(a *SlackAttachment) {
			a.SetColor("good").
				SetTitle("Test Title").
				SetText("Test text").
				SetFooter("Test footer").
				AddField("Field 1", "Value 1", true).
				AddField("Field 2", "Value 2", false)
		})

	if msg.Content != "Test notification" {
		t.Errorf("expected content %q, got %q", "Test notification", msg.Content)
	}

	if len(msg.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(msg.Attachments))
	}

	att := msg.Attachments[0]
	if att.Color != "good" {
		t.Errorf("expected color %q, got %q", "good", att.Color)
	}

	if att.Title != "Test Title" {
		t.Errorf("expected title %q, got %q", "Test Title", att.Title)
	}

	if att.Text != "Test text" {
		t.Errorf("expected text %q, got %q", "Test text", att.Text)
	}

	if len(att.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(att.Fields))
	}

	if att.Fields[0].Title != "Field 1" || att.Fields[0].Value != "Value 1" {
		t.Error("expected field 1 to match")
	}

	if !att.Fields[0].Short {
		t.Error("expected field 1 to be short")
	}

	if att.Fields[1].Short {
		t.Error("expected field 2 not to be short")
	}
}

func TestDatabaseNotification(t *testing.T) {
	notif := &DatabaseNotification{
		ID:           "1",
		Type:         "test",
		NotifiableID: "user1",
		Data:         map[string]any{"key": "value"},
		CreatedAt:    time.Now(),
	}

	if !notif.IsUnread() {
		t.Error("expected notification to be unread initially")
	}

	if notif.IsRead() {
		t.Error("expected notification not to be read initially")
	}

	notif.MarkAsRead()

	if !notif.IsRead() {
		t.Error("expected notification to be read after marking")
	}

	if notif.IsUnread() {
		t.Error("expected notification not to be unread after marking")
	}

	if notif.ReadAt == nil {
		t.Error("expected ReadAt to be set after marking as read")
	}
}

// mailOnlyNotification only uses the mail channel
type mailOnlyNotification struct {
	subject string
	message string
}

func (n *mailOnlyNotification) Via(notifiable any) []string {
	return []string{"mail"}
}

func (n *mailOnlyNotification) ToMail(notifiable any) *mail.Message {
	msg := mail.NewMessage()
	msg.SetFrom("noreply@example.com")
	msg.SetSubject(n.subject)
	msg.SetBody(n.message)
	return msg
}

func TestNotificationWithOnlyMail(t *testing.T) {
	notification := &mailOnlyNotification{
		subject: "Mail Only",
		message: "This only goes via email",
	}

	transport := mail.NewArrayTransport()
	mailer := mail.NewMailer(mail.MailConfig{
		FromAddress: "system@example.com",
		Transport:   transport,
	})
	mailChannel := NewMailChannel(mailer)

	sender := NewSender()
	sender.Via("mail", mailChannel)

	user := &testUser{email: "user@example.com", id: "1"}

	err := sender.SendNow(user, notification)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	sent := transport.Sent()
	if len(sent) != 1 {
		t.Errorf("expected 1 message, got %d", len(sent))
	}
}
