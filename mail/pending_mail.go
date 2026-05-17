package mail

import "fmt"

// PendingMail represents a fluent email builder.
type PendingMail struct {
	mailer  *Mailer
	to      []Address
	cc      []Address
	bcc     []Address
	replyTo []Address
}

// To adds a recipient to the pending mail.
func (p *PendingMail) To(address string, name ...string) *PendingMail {
	addr := Address{Address: address}
	if len(name) > 0 && name[0] != "" {
		addr.Name = name[0]
	}
	p.to = append(p.to, addr)
	return p
}

// Cc adds a CC recipient to the pending mail.
func (p *PendingMail) Cc(address string, name ...string) *PendingMail {
	addr := Address{Address: address}
	if len(name) > 0 && name[0] != "" {
		addr.Name = name[0]
	}
	p.cc = append(p.cc, addr)
	return p
}

// Bcc adds a BCC recipient to the pending mail.
func (p *PendingMail) Bcc(address string, name ...string) *PendingMail {
	addr := Address{Address: address}
	if len(name) > 0 && name[0] != "" {
		addr.Name = name[0]
	}
	p.bcc = append(p.bcc, addr)
	return p
}

// ReplyTo sets the reply-to address for the pending mail.
func (p *PendingMail) ReplyTo(address string, name ...string) *PendingMail {
	addr := Address{Address: address}
	if len(name) > 0 && name[0] != "" {
		addr.Name = name[0]
	}
	p.replyTo = []Address{addr}
	return p
}

// Send builds the mailable and sends it immediately.
func (p *PendingMail) Send(mailable Mailable) error {
	if len(p.to) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	message := mailable.Build()

	// Apply pending mail recipients
	message.To = append(message.To, p.to...)
	message.Cc = append(message.Cc, p.cc...)
	message.Bcc = append(message.Bcc, p.bcc...)
	if len(p.replyTo) > 0 {
		message.ReplyTo = p.replyTo
	}

	return p.mailer.Send(message)
}

// Queue queues the mailable for later sending (placeholder).
func (p *PendingMail) Queue(mailable Mailable) error {
	// TODO: Implement queue integration when queue package is available
	return fmt.Errorf("queue functionality not yet implemented")
}
