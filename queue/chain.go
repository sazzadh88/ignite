package queue

// ChainedJob represents a sequence of jobs to execute in order.
// If any job fails, the chain stops.
type ChainedJob struct {
	Jobs []Job
}

// Dispatch executes all jobs in the chain sequentially.
// Returns an error if any job fails.
func (c *ChainedJob) Dispatch() error {
	for _, job := range c.Jobs {
		if err := job.Handle(); err != nil {
			return err
		}
	}
	return nil
}

// DispatchAsync pushes the chain to the queue.
func (c *ChainedJob) DispatchAsync() error {
	// Wrap the chain as a single job
	wrapper := &chainJobWrapper{chain: c}
	return Dispatch(wrapper)
}

// chainJobWrapper wraps a ChainedJob to implement the Job interface.
type chainJobWrapper struct {
	BaseJob
	chain *ChainedJob
}

func (w *chainJobWrapper) Handle() error {
	return w.chain.Dispatch()
}
