package notifications

import "time"

// DatabaseNotification represents a notification stored in the database.
type DatabaseNotification struct {
	ID           string
	Type         string
	NotifiableID string
	Data         map[string]any
	ReadAt       *time.Time
	CreatedAt    time.Time
}

// MarkAsRead marks the notification as read.
func (n *DatabaseNotification) MarkAsRead() {
	now := time.Now()
	n.ReadAt = &now
}

// IsRead returns true if the notification has been read.
func (n *DatabaseNotification) IsRead() bool {
	return n.ReadAt != nil
}

// IsUnread returns true if the notification has not been read.
func (n *DatabaseNotification) IsUnread() bool {
	return n.ReadAt == nil
}
