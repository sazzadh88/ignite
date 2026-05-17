package mail

// Transport defines how messages are delivered.
type Transport interface {
	// Send delivers the message and returns an error if unsuccessful.
	Send(message *Message) error
}

// MailConfig holds mailer configuration.
type MailConfig struct {
	// Default from address
	FromAddress string
	FromName    string

	// Transport to use
	Transport Transport
}

// Mailer manages email sending with a configured transport.
type Mailer struct {
	config MailConfig
}

// NewMailer creates a new mailer with the given configuration.
func NewMailer(config MailConfig) *Mailer {
	return &Mailer{
		config: config,
	}
}

// To creates a pending mail to the given recipient.
func (m *Mailer) To(address string, name ...string) *PendingMail {
	pm := &PendingMail{
		mailer: m,
		to:     []Address{},
		cc:     []Address{},
		bcc:    []Address{},
	}

	addr := Address{Address: address}
	if len(name) > 0 && name[0] != "" {
		addr.Name = name[0]
	}
	pm.to = append(pm.to, addr)

	return pm
}

// Send sends a message directly using the configured transport.
func (m *Mailer) Send(message *Message) error {
	// Set default from address if not set
	if len(message.From) == 0 && m.config.FromAddress != "" {
		message.SetFrom(m.config.FromAddress, m.config.FromName)
	}

	return m.config.Transport.Send(message)
}

// Mail is the global mailer instance.
var Mail *Mailer
