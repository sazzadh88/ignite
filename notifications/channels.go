package notifications

import (
	"fmt"
	"time"

	"github.com/sazzadh88/ignite/mail"
)

// Channel defines the interface for notification channels.
type Channel interface {
	// Send sends the notification to the notifiable entity.
	Send(notifiable Notifiable, notification Notification) error
}

// MailChannel sends notifications via email.
type MailChannel struct {
	mailer *mail.Mailer
}

// NewMailChannel creates a new mail channel with the given mailer.
func NewMailChannel(mailer *mail.Mailer) *MailChannel {
	return &MailChannel{
		mailer: mailer,
	}
}

// Send sends the notification via email.
func (c *MailChannel) Send(notifiable Notifiable, notification Notification) error {
	// Check if notification implements ToMail
	mailNotification, ok := notification.(ToMail)
	if !ok {
		return fmt.Errorf("notification does not implement ToMail interface")
	}

	// Get email address from notifiable
	email := notifiable.RouteNotificationFor("mail")
	if email == "" {
		return fmt.Errorf("notifiable does not have an email address")
	}

	// Build the message
	message := mailNotification.ToMail(notifiable)

	// Add recipient if not already set
	if len(message.To) == 0 {
		message.AddTo(email)
	}

	// Send the message
	return c.mailer.Send(message)
}

// DatabaseChannel stores notifications in the database.
type DatabaseChannel struct {
	storeFunc func(*DatabaseNotification) error
}

// NewDatabaseChannel creates a new database channel with the given store function.
func NewDatabaseChannel(storeFunc func(*DatabaseNotification) error) *DatabaseChannel {
	return &DatabaseChannel{
		storeFunc: storeFunc,
	}
}

// Send stores the notification in the database.
func (c *DatabaseChannel) Send(notifiable Notifiable, notification Notification) error {
	// Check if notification implements ToDatabase
	dbNotification, ok := notification.(ToDatabase)
	if !ok {
		return fmt.Errorf("notification does not implement ToDatabase interface")
	}

	// Get notifiable ID
	notifiableID := notifiable.RouteNotificationFor("database")
	if notifiableID == "" {
		return fmt.Errorf("notifiable does not have a database ID")
	}

	// Build the database notification
	data := dbNotification.ToDatabase(notifiable)

	dbNotif := &DatabaseNotification{
		Type:         fmt.Sprintf("%T", notification),
		NotifiableID: notifiableID,
		Data:         data,
		CreatedAt:    time.Now(),
	}

	// Store it
	return c.storeFunc(dbNotif)
}

// SlackChannel sends notifications via Slack webhooks.
type SlackChannel struct {
	// TODO: Implement Slack webhook integration when HTTP client is available
}

// NewSlackChannel creates a new Slack channel.
func NewSlackChannel() *SlackChannel {
	return &SlackChannel{}
}

// Send sends the notification via Slack (placeholder).
func (c *SlackChannel) Send(notifiable Notifiable, notification Notification) error {
	// Check if notification implements ToSlack
	slackNotification, ok := notification.(ToSlack)
	if !ok {
		return fmt.Errorf("notification does not implement ToSlack interface")
	}

	// Get webhook URL from notifiable
	webhook := notifiable.RouteNotificationFor("slack")
	if webhook == "" {
		return fmt.Errorf("notifiable does not have a Slack webhook URL")
	}

	// Build the message
	_ = slackNotification.ToSlack(notifiable)

	// TODO: Send to Slack webhook using HTTP client
	return fmt.Errorf("slack channel not yet implemented")
}
