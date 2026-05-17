// Package pipeline provides a fluent interface for passing data through
// a series of pipes, similar to Laravel's Pipeline pattern.
package pipeline

// Pipe represents a single stage in the pipeline that can process data
// and optionally pass it to the next pipe.
type Pipe interface {
	Handle(passable any, next func(any) any) any
}

// PipeFunc is a function adapter that implements the Pipe interface.
type PipeFunc func(passable any, next func(any) any) any

// Handle implements the Pipe interface for PipeFunc.
func (f PipeFunc) Handle(passable any, next func(any) any) any {
	return f(passable, next)
}

// Pipeline provides a fluent interface for passing data through a series of pipes.
type Pipeline struct {
	passable any
	pipes    []any
	method   string
}

// New creates a new Pipeline instance.
func New() *Pipeline {
	return &Pipeline{
		method: "Handle",
	}
}

// Send sets the data that will be passed through the pipeline.
func Send(passable any) *Pipeline {
	return New().Send(passable)
}

// Send sets the data that will be passed through the pipeline.
func (p *Pipeline) Send(passable any) *Pipeline {
	p.passable = passable
	return p
}

// Through sets the pipes that the data will pass through.
// Each pipe can be a Pipe interface or a PipeFunc.
func (p *Pipeline) Through(pipes ...any) *Pipeline {
	p.pipes = pipes
	return p
}

// Pipe is an alias for Through.
func (p *Pipeline) Pipe(pipes ...any) *Pipeline {
	return p.Through(pipes...)
}

// Via sets a custom method name to call on pipe objects.
// Default is "Handle".
func (p *Pipeline) Via(method string) *Pipeline {
	p.method = method
	return p
}

// Then executes the pipeline with a final handler.
func (p *Pipeline) Then(fn func(any) any) any {
	pipeline := p.buildPipeline(fn)
	return pipeline(p.passable)
}

// ThenReturn executes the pipeline and returns the result without a final handler.
func (p *Pipeline) ThenReturn() any {
	return p.Then(func(passable any) any {
		return passable
	})
}

// buildPipeline creates a nested function chain from the pipes.
func (p *Pipeline) buildPipeline(final func(any) any) func(any) any {
	pipeline := final

	// Build pipeline in reverse order
	for i := len(p.pipes) - 1; i >= 0; i-- {
		pipe := p.pipes[i]
		next := pipeline

		pipeline = func(passable any) any {
			return p.executePipe(pipe, passable, next)
		}
	}

	return pipeline
}

// executePipe executes a single pipe.
func (p *Pipeline) executePipe(pipe any, passable any, next func(any) any) any {
	switch v := pipe.(type) {
	case Pipe:
		return v.Handle(passable, next)
	case func(any, func(any) any) any:
		return v(passable, next)
	default:
		// If it's not a recognized type, just pass through
		return next(passable)
	}
}
