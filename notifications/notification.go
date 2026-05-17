// Package notifications provides a multi-channel notification system.
package notifications

import "github.com/sazzadh88/ignite/mail"

// Notification defines the interface for all notifications.
type Notification interface {
	// Via returns the channels this notification should be sent through for the given notifiable.
	Via(notifiable any) []string
}

// ToMail defines the interface for notifications that can be sent via email.
type ToMail interface {
	// ToMail builds the email message for the given notifiable.
	ToMail(notifiable any) *mail.Message
}

// ToDatabase defines the interface for notifications that can be stored in the database.
type ToDatabase interface {
	// ToDatabase returns the data to be stored in the database for the given notifiable.
	ToDatabase(notifiable any) map[string]any
}

// ToSlack defines the interface for notifications that can be sent via Slack.
type ToSlack interface {
	// ToSlack builds the Slack message for the given notifiable.
	ToSlack(notifiable any) *SlackMessage
}

// Notifiable defines the interface for entities that can receive notifications.
type Notifiable interface {
	// NotifyVia returns the preferred notification channels.
	NotifyVia() []string

	// RouteNotificationFor returns the routing information for the given channel.
	// For example, for "mail" it would return an email address,
	// for "slack" it would return a webhook URL, etc.
	RouteNotificationFor(channel string) string
}
