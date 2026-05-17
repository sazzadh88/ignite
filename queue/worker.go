package queue

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WorkerConfig configures a queue worker.
type WorkerConfig struct {
	// MaxJobs is the maximum number of jobs to process before stopping (0 = unlimited).
	MaxJobs int
	// MaxTime is the maximum duration to run before stopping (0 = unlimited).
	MaxTime time.Duration
	// Sleep is the duration to wait when no jobs are available.
	Sleep time.Duration
	// StopWhenEmpty stops the worker when the queue is empty.
	StopWhenEmpty bool
	// Timeout is the default timeout for jobs (if Job.Timeout() returns 0).
	Timeout time.Duration
}

// Worker processes jobs from queues.
type Worker struct {
	driver Driver
	config WorkerConfig
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	store  FailedJobStore
}

// NewWorker creates a new queue worker.
func NewWorker(driver Driver, config WorkerConfig) *Worker {
	if config.Sleep <= 0 {
		config.Sleep = 1 * time.Second
	}
	if config.Timeout <= 0 {
		config.Timeout = 60 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		driver: driver,
		config: config,
		ctx:    ctx,
		cancel: cancel,
		store:  NewMemoryFailedStore(),
	}
}

// SetFailedStore sets the failed job store.
func (w *Worker) SetFailedStore(store FailedJobStore) {
	w.store = store
}

// Work starts processing jobs from the given queues.
func (w *Worker) Work(queues []string) {
	if len(queues) == 0 {
		queues = []string{"default"}
	}

	startTime := time.Now()
	processedCount := 0

	for {
		// Check stop conditions
		if w.ctx.Err() != nil {
			break
		}
		if w.config.MaxJobs > 0 && processedCount >= w.config.MaxJobs {
			break
		}
		if w.config.MaxTime > 0 && time.Since(startTime) >= w.config.MaxTime {
			break
		}

		// Try to get a job from any queue
		job, queue := w.popJob(queues)
		if job == nil {
			if w.config.StopWhenEmpty {
				break
			}
			time.Sleep(w.config.Sleep)
			continue
		}

		// Process the job
		w.processJob(job, queue)
		processedCount++
	}
}

// popJob tries to pop a job from any of the queues.
func (w *Worker) popJob(queues []string) (Job, string) {
	for _, queue := range queues {
		job, err := w.driver.Pop(queue)
		if err == nil && job != nil {
			return job, queue
		}
	}
	return nil, ""
}

// processJob executes a job with retry logic.
func (w *Worker) processJob(job Job, queue string) {
	attempts := 0
	maxTries := job.Tries()
	if maxTries <= 0 {
		maxTries = 1
	}

	for attempts < maxTries {
		attempts++

		err := w.runJob(job)
		if err == nil {
			return // Success
		}

		// Job failed
		if attempts >= maxTries {
			// Max tries reached, store as failed
			w.handleFailedJob(job, queue, err)
			return
		}

		// Retry with backoff
		backoff := w.getBackoff(job, attempts)
		time.Sleep(backoff)
	}
}

// runJob executes a job with timeout.
func (w *Worker) runJob(job Job) error {
	timeout := job.Timeout()
	if timeout <= 0 {
		timeout = w.config.Timeout
	}

	ctx, cancel := context.WithTimeout(w.ctx, timeout)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- job.Handle()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return fmt.Errorf("job timed out after %v", timeout)
	}
}

// BackoffProvider is an optional interface for jobs that provide custom backoff logic.
type BackoffProvider interface {
	GetBackoff(attempt int) time.Duration
}

// getBackoff determines the backoff duration for a retry.
func (w *Worker) getBackoff(job Job, attempt int) time.Duration {
	if provider, ok := job.(BackoffProvider); ok {
		return provider.GetBackoff(attempt)
	}
	// Default exponential backoff
	backoff := time.Duration(1<<uint(attempt)) * time.Second
	if backoff > 5*time.Minute {
		backoff = 5 * time.Minute
	}
	return backoff
}

// handleFailedJob stores a failed job.
func (w *Worker) handleFailedJob(job Job, queue string, err error) {
	if w.store == nil {
		return
	}

	failed := FailedJob{
		ID:        fmt.Sprintf("%p", job),
		Queue:     queue,
		Payload:   fmt.Sprintf("%+v", job),
		Exception: err,
		FailedAt:  time.Now(),
	}

	_ = w.store.Store(failed)
}

// Stop gracefully stops the worker.
func (w *Worker) Stop() {
	w.cancel()
	w.wg.Wait()
}
