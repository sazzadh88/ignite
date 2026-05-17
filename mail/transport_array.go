package mail

import "sync"

// ArrayTransport captures sent messages in memory for testing.
type ArrayTransport struct {
	mu       sync.RWMutex
	messages []*Message
}

// NewArrayTransport creates a new array transport that captures sent messages.
func NewArrayTransport() *ArrayTransport {
	return &ArrayTransport{
		messages: make([]*Message, 0),
	}
}

// Send captures the message in memory.
func (t *ArrayTransport) Send(message *Message) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Deep copy the message to avoid mutations
	msg := &Message{
		From:        append([]Address{}, message.From...),
		To:          append([]Address{}, message.To...),
		Cc:          append([]Address{}, message.Cc...),
		Bcc:         append([]Address{}, message.Bcc...),
		ReplyTo:     append([]Address{}, message.ReplyTo...),
		Subject:     message.Subject,
		Body:        message.Body,
		HTMLBody:    message.HTMLBody,
		Attachments: append([]Attachment{}, message.Attachments...),
		Headers:     make(map[string]string),
		Tags:        append([]string{}, message.Tags...),
		Metadata:    make(map[string]string),
	}

	for k, v := range message.Headers {
		msg.Headers[k] = v
	}

	for k, v := range message.Metadata {
		msg.Metadata[k] = v
	}

	t.messages = append(t.messages, msg)
	return nil
}

// Sent returns all captured messages.
func (t *ArrayTransport) Sent() []*Message {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]*Message, len(t.messages))
	copy(result, t.messages)
	return result
}

// Clear removes all captured messages.
func (t *ArrayTransport) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.messages = make([]*Message, 0)
}
