package broadcasting

// ShouldBroadcast defines an interface for events that should be broadcast.
type ShouldBroadcast interface {
	// BroadcastOn returns the channels the event should broadcast on.
	BroadcastOn() []Channel
	// BroadcastAs returns the event name for the broadcast.
	BroadcastAs() string
	// BroadcastWith returns the data to broadcast.
	BroadcastWith() map[string]any
}

// ShouldBroadcastNow marks an event that should be broadcast immediately,
// bypassing the queue system.
type ShouldBroadcastNow interface {
	ShouldBroadcast
}

// BroadcastQueueable defines an optional interface for events that can specify
// their queue connection and name.
type BroadcastQueueable interface {
	// BroadcastQueue returns the queue name for this broadcast.
	BroadcastQueue() string
}

// BroadcastAfterCommitable defines an optional interface for events that should
// be broadcast only after database transaction commits.
type BroadcastAfterCommitable interface {
	// BroadcastAfterCommit returns true if the broadcast should wait for commit.
	BroadcastAfterCommit() bool
}
