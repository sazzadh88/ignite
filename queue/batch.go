package queue

import (
	"fmt"
	"sync"
	"time"
)

// Batch represents a collection of jobs being processed together.
type Batch struct {
	// ID is a unique identifier for the batch.
	ID string
	// TotalJobs is the total number of jobs in the batch.
	TotalJobs int
	// ProcessedJobs is the number of jobs that have completed (success or failure).
	ProcessedJobs int
	// FailedJobs is the number of jobs that failed.
	FailedJobs int
	// Finished indicates whether all jobs have been processed.
	Finished bool
	// Cancelled indicates whether the batch was cancelled.
	Cancelled bool

	mu             sync.Mutex
	jobs           []Job
	thenCallback   func(*Batch)
	catchCallback  func(*Batch, error)
	finalCallback  func(*Batch)
	allowFailures  bool
	firstError     error
	createdAt      time.Time
}

// BatchBuilder constructs a batch with configuration.
type BatchBuilder struct {
	jobs          []Job
	thenCallback  func(*Batch)
	catchCallback func(*Batch, error)
	finalCallback func(*Batch)
	allowFailures bool
}

// Then sets a callback to run when all jobs succeed.
func (b *BatchBuilder) Then(fn func(*Batch)) *BatchBuilder {
	b.thenCallback = fn
	return b
}

// Catch sets a callback to run when any job fails.
func (b *BatchBuilder) Catch(fn func(*Batch, error)) *BatchBuilder {
	b.catchCallback = fn
	return b
}

// Finally sets a callback to run after all jobs complete, regardless of outcome.
func (b *BatchBuilder) Finally(fn func(*Batch)) *BatchBuilder {
	b.finalCallback = fn
	return b
}

// AllowFailures configures the batch to continue even if some jobs fail.
func (b *BatchBuilder) AllowFailures() *BatchBuilder {
	b.allowFailures = true
	return b
}

// Dispatch creates and dispatches the batch.
func (b *BatchBuilder) Dispatch() (*Batch, error) {
	batch := &Batch{
		ID:             fmt.Sprintf("batch_%d", time.Now().UnixNano()),
		TotalJobs:      len(b.jobs),
		jobs:           b.jobs,
		thenCallback:   b.thenCallback,
		catchCallback:  b.catchCallback,
		finalCallback:  b.finalCallback,
		allowFailures:  b.allowFailures,
		createdAt:      time.Now(),
	}

	// Process jobs
	go batch.process()

	return batch, nil
}

// process executes all jobs in the batch.
func (b *Batch) process() {
	var wg sync.WaitGroup
	errorCh := make(chan error, len(b.jobs))

	for _, job := range b.jobs {
		wg.Add(1)
		go func(j Job) {
			defer wg.Done()

			err := j.Handle()

			b.mu.Lock()
			b.ProcessedJobs++
			if err != nil {
				b.FailedJobs++
				if b.firstError == nil {
					b.firstError = err
				}
				errorCh <- err

				// Cancel remaining jobs if not allowing failures
				if !b.allowFailures {
					b.Cancelled = true
				}
			}
			b.mu.Unlock()
		}(job)
	}

	wg.Wait()
	close(errorCh)

	// Mark as finished
	b.mu.Lock()
	b.Finished = true
	finished := b.Finished
	failed := b.FailedJobs > 0
	firstErr := b.firstError
	b.mu.Unlock()

	// Run callbacks
	if finished {
		if failed {
			if b.catchCallback != nil {
				b.catchCallback(b, firstErr)
			}
		} else {
			if b.thenCallback != nil {
				b.thenCallback(b)
			}
		}

		if b.finalCallback != nil {
			b.finalCallback(b)
		}
	}
}

// Pending returns the number of jobs not yet processed.
func (b *Batch) Pending() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.TotalJobs - b.ProcessedJobs
}

// Progress returns the completion percentage (0-100).
func (b *Batch) Progress() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.TotalJobs == 0 {
		return 100
	}
	return (b.ProcessedJobs * 100) / b.TotalJobs
}

// Cancel cancels the batch.
func (b *Batch) Cancel() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Cancelled = true
}
