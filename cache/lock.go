package cache

import (
	"fmt"
	"time"
)

// Lock provides atomic locking functionality using the cache.
type Lock struct {
	repo  *Repository
	key   string
	ttl   time.Duration
	owner string
}

// NewLock creates a new Lock for the given key.
func NewLock(repo *Repository, key string, ttl time.Duration) *Lock {
	return &Lock{
		repo:  repo,
		key:   fmt.Sprintf("lock:%s", key),
		ttl:   ttl,
		owner: fmt.Sprintf("%d", time.Now().UnixNano()),
	}
}

// Get attempts to acquire the lock and execute the callback.
// Returns true if the lock was acquired and the callback executed.
func (l *Lock) Get(fn func()) bool {
	if !l.acquire() {
		return false
	}
	defer l.Release()

	fn()
	return true
}

// Block waits for the lock to become available, then executes the callback.
// Returns false if the timeout is reached without acquiring the lock.
func (l *Lock) Block(timeout time.Duration, fn func()) bool {
	start := time.Now()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if l.acquire() {
			defer l.Release()
			fn()
			return true
		}

		select {
		case <-ticker.C:
			if time.Since(start) >= timeout {
				return false
			}
		}
	}
}

// Release releases the lock if owned by this instance.
func (l *Lock) Release() bool {
	owner := l.repo.GetString(l.key)
	if owner != l.owner {
		return false
	}

	return l.repo.Forget(l.key)
}

// ForceRelease releases the lock regardless of ownership.
func (l *Lock) ForceRelease() {
	l.repo.Forget(l.key)
}

// acquire attempts to acquire the lock.
func (l *Lock) acquire() bool {
	return l.repo.Add(l.key, l.owner, l.ttl)
}
