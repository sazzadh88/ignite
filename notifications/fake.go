package notifications

import (
	"fmt"
	"sync"
)

// sentNotification represents a captured notification.
type sentNotification struct {
	notifiable   Notifiable
	notification Notification
}

// FakeSender provides a testing double for the Sender.
type FakeSender struct {
	mu    sync.RWMutex
	sent  []sentNotification
	queue []sentNotification
}

// Fake creates a new fake sender for testing.
func Fake() *FakeSender {
	return &FakeSender{
		sent:  make([]sentNotification, 0),
		queue: make([]sentNotification, 0),
	}
}

// Send captures the notification instead of sending it.
func (f *FakeSender) Send(notifiable Notifiable, notification Notification) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.sent = append(f.sent, sentNotification{
		notifiable:   notifiable,
		notification: notification,
	})
	return nil
}

// SendNow is an alias for Send in the fake.
func (f *FakeSender) SendNow(notifiable Notifiable, notification Notification) error {
	return f.Send(notifiable, notification)
}

// AssertSentTo checks if a notification of the given type was sent to the notifiable.
func (f *FakeSender) AssertSentTo(notifiable Notifiable, notificationType any) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	expectedType := fmt.Sprintf("%T", notificationType)

	for _, sent := range f.sent {
		if sent.notifiable == notifiable {
			actualType := fmt.Sprintf("%T", sent.notification)
			if actualType == expectedType {
				return true
			}
		}
	}
	return false
}

// AssertNotSentTo checks if a notification of the given type was not sent to the notifiable.
func (f *FakeSender) AssertNotSentTo(notifiable Notifiable, notificationType any) bool {
	return !f.AssertSentTo(notifiable, notificationType)
}

// AssertCount checks if the exact number of notifications were sent.
func (f *FakeSender) AssertCount(count int) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return len(f.sent) == count
}

// AssertNothingSent checks if no notifications were sent.
func (f *FakeSender) AssertNothingSent() bool {
	return f.AssertCount(0)
}

// Clear removes all captured notifications.
func (f *FakeSender) Clear() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.sent = make([]sentNotification, 0)
	f.queue = make([]sentNotification, 0)
}

// Sent returns all captured notifications.
func (f *FakeSender) Sent() []sentNotification {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make([]sentNotification, len(f.sent))
	copy(result, f.sent)
	return result
}
