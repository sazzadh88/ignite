package queue

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// testJob is a simple job for testing.
type testJob struct {
	BaseJob
	handler func() error
	counter *int32
}

func (j *testJob) Handle() error {
	if j.counter != nil {
		atomic.AddInt32(j.counter, 1)
	}
	if j.handler != nil {
		return j.handler()
	}
	return nil
}

// TestSyncDriverExecutesImmediately tests that sync driver runs jobs immediately.
func TestSyncDriverExecutesImmediately(t *testing.T) {
	driver := &SyncDriver{}
	executed := false

	job := &testJob{
		handler: func() error {
			executed = true
			return nil
		},
	}

	err := driver.Push(job, "default")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if !executed {
		t.Error("expected job to execute immediately")
	}
}

// TestMemoryDriverPushPop tests memory driver push and pop operations.
func TestMemoryDriverPushPop(t *testing.T) {
	driver := NewMemoryDriver()
	defer driver.Stop()

	job1 := &testJob{BaseJob: BaseJob{QueueName: "test"}}
	job2 := &testJob{BaseJob: BaseJob{QueueName: "test"}}

	// Push jobs
	if err := driver.Push(job1, "test"); err != nil {
		t.Fatalf("push failed: %v", err)
	}
	if err := driver.Push(job2, "test"); err != nil {
		t.Fatalf("push failed: %v", err)
	}

	// Check size
	if size := driver.Size("test"); size != 2 {
		t.Errorf("expected size 2, got %d", size)
	}

	// Pop jobs
	popped, err := driver.Pop("test")
	if err != nil {
		t.Fatalf("pop failed: %v", err)
	}
	if popped != job1 {
		t.Error("expected first job")
	}

	popped, err = driver.Pop("test")
	if err != nil {
		t.Fatalf("pop failed: %v", err)
	}
	if popped != job2 {
		t.Error("expected second job")
	}

	// Queue should be empty
	if size := driver.Size("test"); size != 0 {
		t.Errorf("expected empty queue, got size %d", size)
	}
}

// TestMemoryDriverLater tests delayed job execution.
func TestMemoryDriverLater(t *testing.T) {
	driver := NewMemoryDriver()
	defer driver.Stop()

	job := &testJob{BaseJob: BaseJob{QueueName: "test"}}

	// Schedule job with 100ms delay
	start := time.Now()
	if err := driver.Later(100*time.Millisecond, job, "test"); err != nil {
		t.Fatalf("later failed: %v", err)
	}

	// Queue should initially be empty
	if size := driver.Size("test"); size != 0 {
		t.Errorf("expected empty queue initially, got size %d", size)
	}

	// Wait for job to become available
	time.Sleep(150 * time.Millisecond)

	// Job should now be available
	if size := driver.Size("test"); size != 1 {
		t.Errorf("expected size 1 after delay, got %d", size)
	}

	elapsed := time.Since(start)
	if elapsed < 100*time.Millisecond {
		t.Errorf("job became available too early: %v", elapsed)
	}
}

// TestWorkerProcessesJobs tests that worker processes jobs from queue.
func TestWorkerProcessesJobs(t *testing.T) {
	driver := NewMemoryDriver()
	defer driver.Stop()

	var counter int32
	job1 := &testJob{
		BaseJob: BaseJob{QueueName: "default"},
		counter: &counter,
	}
	job2 := &testJob{
		BaseJob: BaseJob{QueueName: "default"},
		counter: &counter,
	}

	driver.Push(job1, "default")
	driver.Push(job2, "default")

	config := WorkerConfig{
		MaxJobs:       2,
		Sleep:         10 * time.Millisecond,
		StopWhenEmpty: false,
	}

	worker := NewWorker(driver, config)
	worker.Work([]string{"default"})

	if atomic.LoadInt32(&counter) != 2 {
		t.Errorf("expected 2 jobs processed, got %d", counter)
	}
}

// TestWorkerRetry tests that worker retries failed jobs.
func TestWorkerRetry(t *testing.T) {
	driver := NewMemoryDriver()
	defer driver.Stop()

	attempts := int32(0)
	job := &testJob{
		BaseJob: BaseJob{
			QueueName:        "default",
			MaxTries:         3,
			BackoffDurations: []time.Duration{1 * time.Millisecond, 1 * time.Millisecond},
		},
		handler: func() error {
			atomic.AddInt32(&attempts, 1)
			return errors.New("job failed")
		},
	}

	driver.Push(job, "default")

	config := WorkerConfig{
		MaxJobs:       1,
		Sleep:         10 * time.Millisecond,
		StopWhenEmpty: false,
	}

	store := NewMemoryFailedStore()
	worker := NewWorker(driver, config)
	worker.SetFailedStore(store)
	worker.Work([]string{"default"})

	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}

	// Should be in failed job store
	failedJobs := store.All()
	if len(failedJobs) != 1 {
		t.Errorf("expected 1 failed job, got %d", len(failedJobs))
	}
}

// TestChain tests chained job execution.
func TestChain(t *testing.T) {
	var order []int
	var mu sync.Mutex

	job1 := &testJob{
		handler: func() error {
			mu.Lock()
			order = append(order, 1)
			mu.Unlock()
			return nil
		},
	}
	job2 := &testJob{
		handler: func() error {
			mu.Lock()
			order = append(order, 2)
			mu.Unlock()
			return nil
		},
	}
	job3 := &testJob{
		handler: func() error {
			mu.Lock()
			order = append(order, 3)
			mu.Unlock()
			return nil
		},
	}

	chain := Chain([]Job{job1, job2, job3})
	err := chain.Dispatch()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(order) != 3 || order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Errorf("expected order [1,2,3], got %v", order)
	}
}

// TestChainStopsOnFailure tests that chain stops when a job fails.
func TestChainStopsOnFailure(t *testing.T) {
	executed := make(map[int]bool)
	var mu sync.Mutex

	job1 := &testJob{
		handler: func() error {
			mu.Lock()
			executed[1] = true
			mu.Unlock()
			return nil
		},
	}
	job2 := &testJob{
		handler: func() error {
			mu.Lock()
			executed[2] = true
			mu.Unlock()
			return errors.New("job 2 failed")
		},
	}
	job3 := &testJob{
		handler: func() error {
			mu.Lock()
			executed[3] = true
			mu.Unlock()
			return nil
		},
	}

	chain := Chain([]Job{job1, job2, job3})
	err := chain.Dispatch()

	if err == nil {
		t.Error("expected error from failed job")
	}

	mu.Lock()
	defer mu.Unlock()
	if !executed[1] {
		t.Error("job 1 should have executed")
	}
	if !executed[2] {
		t.Error("job 2 should have executed")
	}
	if executed[3] {
		t.Error("job 3 should not have executed")
	}
}

// TestBatch tests batch job execution.
func TestBatch(t *testing.T) {
	var counter int32

	jobs := []Job{
		&testJob{counter: &counter},
		&testJob{counter: &counter},
		&testJob{counter: &counter},
	}

	thenCalled := false
	batch, err := NewBatch(jobs).Then(func(b *Batch) {
		thenCalled = true
	}).Dispatch()

	if err != nil {
		t.Fatalf("batch dispatch failed: %v", err)
	}

	// Wait for batch to complete
	time.Sleep(100 * time.Millisecond)

	if !batch.Finished {
		t.Error("batch should be finished")
	}

	if atomic.LoadInt32(&counter) != 3 {
		t.Errorf("expected 3 jobs executed, got %d", counter)
	}

	if !thenCalled {
		t.Error("then callback should have been called")
	}
}

// TestBatchCatch tests batch catch callback on failure.
func TestBatchCatch(t *testing.T) {
	jobs := []Job{
		&testJob{handler: func() error { return nil }},
		&testJob{handler: func() error { return errors.New("failed") }},
	}

	catchCalled := false
	finallyCalled := false

	batch, err := NewBatch(jobs).
		Catch(func(b *Batch, err error) {
			catchCalled = true
		}).
		Finally(func(b *Batch) {
			finallyCalled = true
		}).
		Dispatch()

	if err != nil {
		t.Fatalf("batch dispatch failed: %v", err)
	}

	// Wait for batch to complete
	time.Sleep(100 * time.Millisecond)

	if !batch.Finished {
		t.Error("batch should be finished")
	}

	if batch.FailedJobs != 1 {
		t.Errorf("expected 1 failed job, got %d", batch.FailedJobs)
	}

	if !catchCalled {
		t.Error("catch callback should have been called")
	}

	if !finallyCalled {
		t.Error("finally callback should have been called")
	}
}

// TestFailedJobStore tests failed job storage operations.
func TestFailedJobStore(t *testing.T) {
	store := NewMemoryFailedStore()

	job1 := FailedJob{
		ID:        "job1",
		Queue:     "default",
		Payload:   "test payload 1",
		Exception: errors.New("error 1"),
		FailedAt:  time.Now(),
	}
	job2 := FailedJob{
		ID:        "job2",
		Queue:     "default",
		Payload:   "test payload 2",
		Exception: errors.New("error 2"),
		FailedAt:  time.Now(),
	}

	// Store jobs
	if err := store.Store(job1); err != nil {
		t.Errorf("store failed: %v", err)
	}
	if err := store.Store(job2); err != nil {
		t.Errorf("store failed: %v", err)
	}

	// Test All
	all := store.All()
	if len(all) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(all))
	}

	// Test Find
	found := store.Find("job1")
	if found == nil {
		t.Error("expected to find job1")
	}
	if found != nil && found.ID != "job1" {
		t.Errorf("expected job1, got %s", found.ID)
	}

	// Test Forget
	if err := store.Forget("job1"); err != nil {
		t.Errorf("forget failed: %v", err)
	}
	if store.Find("job1") != nil {
		t.Error("job1 should have been forgotten")
	}

	// Test Flush
	if err := store.Flush(); err != nil {
		t.Errorf("flush failed: %v", err)
	}
	if len(store.All()) != 0 {
		t.Error("store should be empty after flush")
	}
}

// TestMemoryDriverFlush tests flushing a queue.
func TestMemoryDriverFlush(t *testing.T) {
	driver := NewMemoryDriver()
	defer driver.Stop()

	driver.Push(&testJob{}, "test")
	driver.Push(&testJob{}, "test")

	if size := driver.Size("test"); size != 2 {
		t.Errorf("expected size 2, got %d", size)
	}

	if err := driver.Flush("test"); err != nil {
		t.Errorf("flush failed: %v", err)
	}

	if size := driver.Size("test"); size != 0 {
		t.Errorf("expected empty queue, got size %d", size)
	}
}

// TestFacadeFunctions tests package-level facade functions.
func TestFacadeFunctions(t *testing.T) {
	// Register a memory driver as default
	driver := NewMemoryDriver()
	defer driver.Stop()
	Register("memory", driver)
	SetDefault("memory")

	var counter int32
	job := &testJob{
		BaseJob: BaseJob{QueueName: "default"},
		counter: &counter,
	}

	// Test Dispatch
	if err := Dispatch(job); err != nil {
		t.Errorf("Dispatch failed: %v", err)
	}

	if driver.Size("default") != 1 {
		t.Error("job should be in queue")
	}

	// Test DispatchSync
	if err := DispatchSync(job); err != nil {
		t.Errorf("DispatchSync failed: %v", err)
	}

	if atomic.LoadInt32(&counter) != 1 {
		t.Error("DispatchSync should execute immediately")
	}
}

// TestBaseJobDefaults tests BaseJob default implementations.
func TestBaseJobDefaults(t *testing.T) {
	job := &BaseJob{}

	if job.Queue() != "default" {
		t.Errorf("expected default queue, got %s", job.Queue())
	}

	if job.Tries() != 1 {
		t.Errorf("expected 1 try, got %d", job.Tries())
	}

	if job.Timeout() != 60*time.Second {
		t.Errorf("expected 60s timeout, got %v", job.Timeout())
	}
}

// TestWorkerStopWhenEmpty tests worker stops when queue is empty.
func TestWorkerStopWhenEmpty(t *testing.T) {
	driver := NewMemoryDriver()
	defer driver.Stop()

	var counter int32
	job := &testJob{
		BaseJob: BaseJob{QueueName: "default"},
		counter: &counter,
	}

	driver.Push(job, "default")

	config := WorkerConfig{
		Sleep:         10 * time.Millisecond,
		StopWhenEmpty: true,
	}

	start := time.Now()
	worker := NewWorker(driver, config)
	worker.Work([]string{"default"})
	elapsed := time.Since(start)

	if atomic.LoadInt32(&counter) != 1 {
		t.Errorf("expected 1 job processed, got %d", counter)
	}

	// Should stop quickly after queue is empty
	if elapsed > 500*time.Millisecond {
		t.Errorf("worker took too long to stop: %v", elapsed)
	}
}
