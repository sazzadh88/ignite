package queue

import (
	"fmt"
	"time"
)

// SyncDriver executes jobs immediately without queueing.
// This is useful for synchronous execution or testing.
type SyncDriver struct{}

// Push executes the job immediately.
func (d *SyncDriver) Push(job Job, queue string) error {
	return job.Handle()
}

// Pop always returns an error as sync driver doesn't queue jobs.
func (d *SyncDriver) Pop(queue string) (Job, error) {
	return nil, fmt.Errorf("sync driver does not support pop")
}

// Later executes the job immediately, ignoring the delay.
func (d *SyncDriver) Later(delay time.Duration, job Job, queue string) error {
	return job.Handle()
}

// Size always returns 0 as sync driver doesn't queue jobs.
func (d *SyncDriver) Size(queue string) int {
	return 0
}

// Flush is a no-op for sync driver.
func (d *SyncDriver) Flush(queue string) error {
	return nil
}
