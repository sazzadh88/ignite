package notifications

import (
	"fmt"
	"sync"
)

// Sender manages notification sending across multiple channels.
type Sender struct {
	mu       sync.RWMutex
	channels map[string]Channel
}

// NewSender creates a new notification sender.
func NewSender() *Sender {
	return &Sender{
		channels: make(map[string]Channel),
	}
}

// Via registers a channel with the sender.
func (s *Sender) Via(name string, channel Channel) *Sender {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.channels[name] = channel
	return s
}

// Send sends a notification to the given notifiable through the appropriate channels.
func (s *Sender) Send(notifiable Notifiable, notification Notification) error {
	// TODO: Queue for later sending
	return s.SendNow(notifiable, notification)
}

// SendNow sends a notification immediately through all applicable channels.
func (s *Sender) SendNow(notifiable Notifiable, notification Notification) error {
	// Get the channels this notification should use
	channels := notification.Via(notifiable)

	if len(channels) == 0 {
		return fmt.Errorf("no channels specified for notification")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Send through each channel
	var lastErr error
	for _, channelName := range channels {
		channel, ok := s.channels[channelName]
		if !ok {
			lastErr = fmt.Errorf("channel %q not registered", channelName)
			continue
		}

		if err := channel.Send(notifiable, notification); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// defaultSender is the global sender instance.
var defaultSender *Sender

// SetDefaultSender sets the global sender instance.
func SetDefaultSender(sender *Sender) {
	defaultSender = sender
}

// Notify sends a notification using the default sender.
func Notify(notifiable Notifiable, notification Notification) error {
	if defaultSender == nil {
		return fmt.Errorf("no default sender configured")
	}
	return defaultSender.Send(notifiable, notification)
}
