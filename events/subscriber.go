package events

// Subscriber defines the interface for event subscribers.
// Subscribers can register multiple event listeners at once.
type Subscriber interface {
	// Subscribe registers the subscriber's event listeners with the dispatcher.
	Subscribe(dispatcher *Dispatcher)
}
