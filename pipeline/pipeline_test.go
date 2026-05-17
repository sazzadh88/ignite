package pipeline

import (
	"strings"
	"testing"
)

func TestSinglePipeTransformsData(t *testing.T) {
	result := Send("hello").
		Through(PipeFunc(func(passable any, next func(any) any) any {
			str := passable.(string)
			return next(strings.ToUpper(str))
		})).
		ThenReturn()

	if result != "HELLO" {
		t.Errorf("expected HELLO, got %v", result)
	}
}

func TestMultiplePipesChainInOrder(t *testing.T) {
	result := Send("hello").
		Through(
			PipeFunc(func(passable any, next func(any) any) any {
				str := passable.(string)
				return next(str + " world")
			}),
			PipeFunc(func(passable any, next func(any) any) any {
				str := passable.(string)
				return next(strings.ToUpper(str))
			}),
			PipeFunc(func(passable any, next func(any) any) any {
				str := passable.(string)
				return next("!" + str + "!")
			}),
		).
		ThenReturn()

	if result != "!HELLO WORLD!" {
		t.Errorf("expected !HELLO WORLD!, got %v", result)
	}
}

func TestPipeFuncWorksAsPipe(t *testing.T) {
	multiply := PipeFunc(func(passable any, next func(any) any) any {
		num := passable.(int)
		return next(num * 2)
	})

	result := Send(5).Through(multiply).ThenReturn()

	if result != 10 {
		t.Errorf("expected 10, got %v", result)
	}
}

func TestPipeCanShortCircuit(t *testing.T) {
	result := Send("hello").
		Through(
			PipeFunc(func(passable any, next func(any) any) any {
				// Short-circuit: don't call next
				return "short-circuited"
			}),
			PipeFunc(func(passable any, next func(any) any) any {
				// This should never be called
				t.Error("this pipe should not be executed")
				return next(passable)
			}),
		).
		ThenReturn()

	if result != "short-circuited" {
		t.Errorf("expected short-circuited, got %v", result)
	}
}

func TestThenReturnWorksWithoutFinalHandler(t *testing.T) {
	result := Send(42).
		Through(
			PipeFunc(func(passable any, next func(any) any) any {
				num := passable.(int)
				return next(num + 8)
			}),
		).
		ThenReturn()

	if result != 50 {
		t.Errorf("expected 50, got %v", result)
	}
}

func TestThenWithFinalHandler(t *testing.T) {
	result := Send(5).
		Through(
			PipeFunc(func(passable any, next func(any) any) any {
				num := passable.(int)
				return next(num * 2)
			}),
		).
		Then(func(passable any) any {
			num := passable.(int)
			return num + 100
		})

	if result != 110 {
		t.Errorf("expected 110, got %v", result)
	}
}

func TestEmptyPipelinePassesThroughUnchanged(t *testing.T) {
	result := Send("unchanged").ThenReturn()

	if result != "unchanged" {
		t.Errorf("expected unchanged, got %v", result)
	}
}

func TestPipeAlias(t *testing.T) {
	result := Send(10).
		Pipe(
			PipeFunc(func(passable any, next func(any) any) any {
				num := passable.(int)
				return next(num + 5)
			}),
		).
		ThenReturn()

	if result != 15 {
		t.Errorf("expected 15, got %v", result)
	}
}

func TestRealWorldMiddlewareExample(t *testing.T) {
	type Request struct {
		Path   string
		Logged bool
		Authed bool
	}

	logging := PipeFunc(func(passable any, next func(any) any) any {
		req := passable.(*Request)
		req.Logged = true
		return next(req)
	})

	auth := PipeFunc(func(passable any, next func(any) any) any {
		req := passable.(*Request)
		if req.Path == "/admin" {
			req.Authed = true
		}
		return next(req)
	})

	result := Send(&Request{Path: "/admin"}).
		Through(logging, auth).
		ThenReturn()

	req := result.(*Request)
	if !req.Logged {
		t.Error("expected request to be logged")
	}
	if !req.Authed {
		t.Error("expected request to be authenticated")
	}
}

type CustomPipe struct {
	multiplier int
}

func (c *CustomPipe) Handle(passable any, next func(any) any) any {
	num := passable.(int)
	return next(num * c.multiplier)
}

func TestCustomPipeInterface(t *testing.T) {
	pipe := &CustomPipe{multiplier: 3}

	result := Send(7).Through(pipe).ThenReturn()

	if result != 21 {
		t.Errorf("expected 21, got %v", result)
	}
}

func TestNewPipeline(t *testing.T) {
	p := New()
	if p == nil {
		t.Error("expected non-nil pipeline")
	}
	if p.method != "Handle" {
		t.Errorf("expected default method Handle, got %s", p.method)
	}
}

func TestViaMethodName(t *testing.T) {
	// Via sets custom method name (though we don't use reflection in this implementation)
	p := New().Via("CustomMethod")
	if p.method != "CustomMethod" {
		t.Errorf("expected CustomMethod, got %s", p.method)
	}
}
