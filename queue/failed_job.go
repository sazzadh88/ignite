package queue

import (
	"sync"
	"time"
)

// FailedJob represents a job that has failed all retry attempts.
type FailedJob struct {
	// ID is a unique identifier for the failed job.
	ID string
	// Queue is the name of the queue the job was on.
	Queue string
	// Payload is a string representation of the job.
	Payload string
	// Exception is the error that caused the failure.
	Exception error
	// FailedAt is when the job failed.
	FailedAt time.Time
}

// FailedJobStore defines the interface for storing and retrieving failed jobs.
type FailedJobStore interface {
	// Store saves a failed job.
	Store(job FailedJob) error
	// All returns all failed jobs.
	All() []FailedJob
	// Find returns a failed job by ID, or nil if not found.
	Find(id string) *FailedJob
	// Flush removes all failed jobs.
	Flush() error
	// Forget removes a specific failed job by ID.
	Forget(id string) error
}

// MemoryFailedStore is an in-memory implementation of FailedJobStore.
type MemoryFailedStore struct {
	mu   sync.RWMutex
	jobs map[string]FailedJob
}

// NewMemoryFailedStore creates a new in-memory failed job store.
func NewMemoryFailedStore() *MemoryFailedStore {
	return &MemoryFailedStore{
		jobs: make(map[string]FailedJob),
	}
}

// Store saves a failed job.
func (s *MemoryFailedStore) Store(job FailedJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
	return nil
}

// All returns all failed jobs.
func (s *MemoryFailedStore) All() []FailedJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]FailedJob, 0, len(s.jobs))
	for _, job := range s.jobs {
		result = append(result, job)
	}
	return result
}

// Find returns a failed job by ID.
func (s *MemoryFailedStore) Find(id string) *FailedJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if job, ok := s.jobs[id]; ok {
		return &job
	}
	return nil
}

// Flush removes all failed jobs.
func (s *MemoryFailedStore) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs = make(map[string]FailedJob)
	return nil
}

// Forget removes a specific failed job.
func (s *MemoryFailedStore) Forget(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.jobs, id)
	return nil
}
