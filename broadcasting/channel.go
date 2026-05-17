package broadcasting

// ChannelType represents the type of broadcast channel.
type ChannelType int

const (
	// Public channel is accessible to all users.
	Public ChannelType = iota
	// Private channel requires authorization.
	Private
	// Presence channel requires authorization and tracks connected users.
	Presence
	// EncryptedPrivate channel requires authorization and encrypts data.
	EncryptedPrivate
)

// Channel represents a broadcast channel with a name and type.
type Channel struct {
	Name string
	Type ChannelType
}

// PublicChannel creates a public broadcast channel.
func PublicChannel(name string) Channel {
	return Channel{Name: name, Type: Public}
}

// PrivateChannel creates a private broadcast channel.
func PrivateChannel(name string) Channel {
	return Channel{Name: name, Type: Private}
}

// PresenceChannel creates a presence broadcast channel.
func PresenceChannel(name string) Channel {
	return Channel{Name: name, Type: Presence}
}

// EncryptedPrivateChannel creates an encrypted private broadcast channel.
func EncryptedPrivateChannel(name string) Channel {
	return Channel{Name: name, Type: EncryptedPrivate}
}
