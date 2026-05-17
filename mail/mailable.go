package mail

// Mailable represents anything that can be converted to an email message.
type Mailable interface {
	// Build constructs and returns the message to be sent.
	Build() *Message
}
