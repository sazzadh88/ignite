package queue

import "time"

// BaseJob provides default implementations of Job interface methods.
// Embed this in your custom jobs to inherit default behavior.
type BaseJob struct {
	// QueueName is the queue this job should be pushed to.
	QueueName string
	// MaxTries is the maximum number of times to attempt this job.
	MaxTries int
	// BackoffDurations defines delays between retry attempts.
	BackoffDurations []time.Duration
	// TimeoutDuration is the maximum time allowed for this job.
	TimeoutDuration time.Duration
	// Delay is how long to wait before making the job available.
	Delay time.Duration
}

// Queue returns the queue name, defaulting to "default".
func (b *BaseJob) Queue() string {
	if b.QueueName == "" {
		return "default"
	}
	return b.QueueName
}

// Tries returns the maximum attempts, defaulting to 1.
func (b *BaseJob) Tries() int {
	if b.MaxTries <= 0 {
		return 1
	}
	return b.MaxTries
}

// Timeout returns the job timeout, defaulting to 60 seconds.
func (b *BaseJob) Timeout() time.Duration {
	if b.TimeoutDuration <= 0 {
		return 60 * time.Second
	}
	return b.TimeoutDuration
}

// GetBackoff returns the backoff duration for the given attempt number.
func (b *BaseJob) GetBackoff(attempt int) time.Duration {
	if attempt > 0 && attempt <= len(b.BackoffDurations) {
		return b.BackoffDurations[attempt-1]
	}
	// Default exponential backoff: 2^attempt seconds
	backoff := time.Duration(1<<uint(attempt)) * time.Second
	if backoff > 5*time.Minute {
		backoff = 5 * time.Minute
	}
	return backoff
}
