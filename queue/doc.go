// Package queue provides a Laravel-inspired queue system for Ignite.
//
// The queue package allows you to defer the processing of time-consuming tasks,
// such as sending emails or processing uploads, until a later time. This can
// dramatically speed up web requests to your application.
//
// # Basic Usage
//
// Define a job by implementing the Job interface:
//
//	type EmailJob struct {
//	    queue.BaseJob
//	    To      string
//	    Subject string
//	    Body    string
//	}
//
//	func (j *EmailJob) Handle() error {
//	    // Send email logic
//	    return nil
//	}
//
// Dispatch the job:
//
//	job := &EmailJob{
//	    BaseJob: queue.BaseJob{QueueName: "emails"},
//	    To:      "user@example.com",
//	    Subject: "Welcome",
//	    Body:    "Welcome to our service",
//	}
//	queue.Dispatch(job)
//
// # Drivers
//
// The package includes several drivers:
//
//   - SyncDriver: Executes jobs immediately (useful for testing)
//   - MemoryDriver: In-memory queue using channels (for development/testing)
//
// Register and configure a driver:
//
//	driver := queue.NewMemoryDriver()
//	queue.Register("memory", driver)
//	queue.SetDefault("memory")
//
// # Queue Workers
//
// Start a worker to process jobs:
//
//	driver := queue.Connection("memory")
//	worker := queue.NewWorker(driver, queue.WorkerConfig{
//	    Sleep:         1 * time.Second,
//	    MaxJobs:       100,
//	    StopWhenEmpty: false,
//	})
//	worker.Work([]string{"default", "emails"})
//
// # Delayed Jobs
//
// Schedule a job to run after a delay:
//
//	queue.Later(5*time.Minute, job)
//
// # Job Chaining
//
// Execute jobs in sequence, stopping on first failure:
//
//	chain := queue.Chain([]queue.Job{job1, job2, job3})
//	chain.Dispatch()
//
// # Job Batching
//
// Execute multiple jobs concurrently with callbacks:
//
//	batch, _ := queue.NewBatch([]queue.Job{job1, job2, job3}).
//	    Then(func(b *queue.Batch) {
//	        // All jobs succeeded
//	    }).
//	    Catch(func(b *queue.Batch, err error) {
//	        // At least one job failed
//	    }).
//	    Finally(func(b *queue.Batch) {
//	        // Always called
//	    }).
//	    Dispatch()
//
// # Retry Logic
//
// Jobs are automatically retried based on their Tries() configuration:
//
//	job := &MyJob{
//	    BaseJob: queue.BaseJob{
//	        MaxTries:         3,
//	        BackoffDurations: []time.Duration{1*time.Second, 5*time.Second},
//	    },
//	}
//
// # Failed Jobs
//
// Failed jobs are tracked and can be inspected:
//
//	store := queue.NewMemoryFailedStore()
//	worker.SetFailedStore(store)
//	// ... after some failures
//	failed := store.All()
package queue
