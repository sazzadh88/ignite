# Queue Package

Laravel-inspired queue system for GoFrame with zero external dependencies.

## Features

- **Multiple Drivers**: Sync (immediate), Memory (in-memory for dev/test)
- **Delayed Jobs**: Schedule jobs to run after a delay
- **Job Chaining**: Execute jobs sequentially, stop on failure
- **Job Batching**: Execute jobs concurrently with progress tracking
- **Automatic Retries**: Configurable retry attempts with exponential backoff
- **Failed Job Tracking**: Store and inspect failed jobs
- **Worker Management**: Process queues with graceful shutdown

## Quick Start

### 1. Define a Job

```go
type EmailJob struct {
    queue.BaseJob
    To      string
    Subject string
    Body    string
}

func (j *EmailJob) Handle() error {
    // Send email logic here
    fmt.Printf("Sending email to %s\n", j.To)
    return nil
}
```

### 2. Configure Queue

```go
// Register memory driver (for development)
driver := queue.NewMemoryDriver()
queue.Register("memory", driver)
queue.SetDefault("memory")
```

### 3. Dispatch Jobs

```go
job := &EmailJob{
    BaseJob: queue.BaseJob{QueueName: "emails"},
    To:      "user@example.com",
    Subject: "Welcome",
    Body:    "Welcome message",
}

// Dispatch asynchronously
queue.Dispatch(job)

// Or execute synchronously
queue.DispatchSync(job)

// Or delay execution
queue.Later(5*time.Minute, job)
```

### 4. Process Queues

```go
driver := queue.Connection("memory")
worker := queue.NewWorker(driver, queue.WorkerConfig{
    Sleep:         1 * time.Second,
    StopWhenEmpty: false,
})

// Start processing (blocking)
worker.Work([]string{"default", "emails"})
```

## Advanced Features

### Retry Configuration

```go
job := &MyJob{
    BaseJob: queue.BaseJob{
        MaxTries:         3,
        BackoffDurations: []time.Duration{1*time.Second, 5*time.Second, 10*time.Second},
        TimeoutDuration:  30 * time.Second,
    },
}
```

### Job Chaining

```go
chain := queue.Chain([]queue.Job{job1, job2, job3})
if err := chain.Dispatch(); err != nil {
    log.Printf("Chain failed: %v", err)
}
```

### Job Batching

```go
batch, _ := queue.NewBatch([]queue.Job{job1, job2, job3}).
    Then(func(b *queue.Batch) {
        fmt.Printf("All %d jobs completed!\n", b.TotalJobs)
    }).
    Catch(func(b *queue.Batch, err error) {
        fmt.Printf("%d jobs failed\n", b.FailedJobs)
    }).
    Finally(func(b *queue.Batch) {
        fmt.Println("Batch finished")
    }).
    Dispatch()

// Monitor progress
fmt.Printf("Progress: %d%%\n", batch.Progress())
```

## Testing

Run all tests:

```bash
go test ./queue/...
```

Run tests with coverage:

```bash
go test ./queue/... -cover
```

## Architecture

- **queue.go** - Core interfaces and facade
- **job_base.go** - Base job implementation with defaults
- **sync_driver.go** - Immediate execution driver
- **memory_driver.go** - In-memory queue with delay support
- **worker.go** - Queue worker with retry logic
- **chain.go** - Sequential job execution
- **batch.go** - Concurrent job execution with callbacks
- **failed_job.go** - Failed job tracking
- **queue_test.go** - Comprehensive test suite

## License

Part of the GoFrame framework.
