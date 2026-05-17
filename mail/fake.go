package mail

import "sync"

// FakeMailer provides a testing double for the Mailer.
type FakeMailer struct {
	mu       sync.RWMutex
	messages []*Message
}

// Fake creates a new fake mailer for testing.
func Fake() *FakeMailer {
	return &FakeMailer{
		messages: make([]*Message, 0),
	}
}

// Send captures the message instead of sending it.
func (f *FakeMailer) Send(message *Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Deep copy the message
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

	f.messages = append(f.messages, msg)
	return nil
}

// To creates a pending mail (for fluent API compatibility).
func (f *FakeMailer) To(address string, name ...string) *FakePendingMail {
	addr := Address{Address: address}
	if len(name) > 0 && name[0] != "" {
		addr.Name = name[0]
	}

	return &FakePendingMail{
		fake: f,
		to:   []Address{addr},
	}
}

// AssertSent checks if any message matches the given predicate.
func (f *FakeMailer) AssertSent(fn func(*Message) bool) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, msg := range f.messages {
		if fn(msg) {
			return true
		}
	}
	return false
}

// AssertNotSent checks if no messages match the given predicate.
func (f *FakeMailer) AssertNotSent(fn func(*Message) bool) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, msg := range f.messages {
		if fn(msg) {
			return false
		}
	}
	return true
}

// AssertSentCount checks if the exact number of messages were sent.
func (f *FakeMailer) AssertSentCount(count int) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return len(f.messages) == count
}

// AssertNothingSent checks if no messages were sent.
func (f *FakeMailer) AssertNothingSent() bool {
	return f.AssertSentCount(0)
}

// Clear removes all captured messages.
func (f *FakeMailer) Clear() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.messages = make([]*Message, 0)
}

// Sent returns all captured messages.
func (f *FakeMailer) Sent() []*Message {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make([]*Message, len(f.messages))
	copy(result, f.messages)
	return result
}

// FakePendingMail mimics PendingMail for testing.
type FakePendingMail struct {
	fake    *FakeMailer
	to      []Address
	cc      []Address
	bcc     []Address
	replyTo []Address
}

// To adds a recipient.
func (p *FakePendingMail) To(address string, name ...string) *FakePendingMail {
	addr := Address{Address: address}
	if len(name) > 0 && name[0] != "" {
		addr.Name = name[0]
	}
	p.to = append(p.to, addr)
	return p
}

// Cc adds a CC recipient.
func (p *FakePendingMail) Cc(address string, name ...string) *FakePendingMail {
	addr := Address{Address: address}
	if len(name) > 0 && name[0] != "" {
		addr.Name = name[0]
	}
	p.cc = append(p.cc, addr)
	return p
}

// Bcc adds a BCC recipient.
func (p *FakePendingMail) Bcc(address string, name ...string) *FakePendingMail {
	addr := Address{Address: address}
	if len(name) > 0 && name[0] != "" {
		addr.Name = name[0]
	}
	p.bcc = append(p.bcc, addr)
	return p
}

// ReplyTo sets the reply-to address.
func (p *FakePendingMail) ReplyTo(address string, name ...string) *FakePendingMail {
	addr := Address{Address: address}
	if len(name) > 0 && name[0] != "" {
		addr.Name = name[0]
	}
	p.replyTo = []Address{addr}
	return p
}

// Send captures the mailable.
func (p *FakePendingMail) Send(mailable Mailable) error {
	message := mailable.Build()

	// Apply pending mail recipients
	message.To = append(message.To, p.to...)
	message.Cc = append(message.Cc, p.cc...)
	message.Bcc = append(message.Bcc, p.bcc...)
	if len(p.replyTo) > 0 {
		message.ReplyTo = p.replyTo
	}

	return p.fake.Send(message)
}

// Queue captures the mailable (same as Send for fake).
func (p *FakePendingMail) Queue(mailable Mailable) error {
	return p.Send(mailable)
}
