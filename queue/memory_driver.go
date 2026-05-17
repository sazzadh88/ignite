package queue

import (
	"fmt"
	"sync"
	"time"
)

type delayedJob struct {
	job   Job
	ready time.Time
}

// MemoryDriver implements an in-memory queue using channels and slices.
// Suitable for development and testing.
type MemoryDriver struct {
	mu       sync.RWMutex
	queues   map[string][]Job
	delayed  map[string][]delayedJob
	stopCh   chan struct{}
	stopOnce sync.Once
}

// NewMemoryDriver creates a new in-memory queue driver.
func NewMemoryDriver() *MemoryDriver {
	d := &MemoryDriver{
		queues:  make(map[string][]Job),
		delayed: make(map[string][]delayedJob),
		stopCh:  make(chan struct{}),
	}
	go d.processDelayed()
	return d
}

// Push adds a job to the queue.
func (d *MemoryDriver) Push(job Job, queue string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.queues[queue] == nil {
		d.queues[queue] = make([]Job, 0)
	}
	d.queues[queue] = append(d.queues[queue], job)
	return nil
}

// Pop retrieves the next job from the queue.
func (d *MemoryDriver) Pop(queue string) (Job, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	jobs := d.queues[queue]
	if len(jobs) == 0 {
		return nil, fmt.Errorf("queue is empty")
	}

	job := jobs[0]
	d.queues[queue] = jobs[1:]
	return job, nil
}

// Later schedules a job to be available after a delay.
func (d *MemoryDriver) Later(delay time.Duration, job Job, queue string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.delayed[queue] == nil {
		d.delayed[queue] = make([]delayedJob, 0)
	}

	d.delayed[queue] = append(d.delayed[queue], delayedJob{
		job:   job,
		ready: time.Now().Add(delay),
	})
	return nil
}

// Size returns the number of immediately available jobs in the queue.
func (d *MemoryDriver) Size(queue string) int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.queues[queue])
}

// Flush removes all jobs from the queue.
func (d *MemoryDriver) Flush(queue string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.queues, queue)
	delete(d.delayed, queue)
	return nil
}

// Stop stops the background delayed job processor.
func (d *MemoryDriver) Stop() {
	d.stopOnce.Do(func() {
		close(d.stopCh)
	})
}

// processDelayed moves delayed jobs to the main queue when ready.
func (d *MemoryDriver) processDelayed() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.checkDelayed()
		}
	}
}

func (d *MemoryDriver) checkDelayed() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	for queue, jobs := range d.delayed {
		ready := make([]Job, 0)
		remaining := make([]delayedJob, 0)

		for _, dj := range jobs {
			if now.After(dj.ready) || now.Equal(dj.ready) {
				ready = append(ready, dj.job)
			} else {
				remaining = append(remaining, dj)
			}
		}

		if len(ready) > 0 {
			if d.queues[queue] == nil {
				d.queues[queue] = make([]Job, 0)
			}
			d.queues[queue] = append(d.queues[queue], ready...)
		}

		if len(remaining) > 0 {
			d.delayed[queue] = remaining
		} else {
			delete(d.delayed, queue)
		}
	}
}
