package container

import (
	"testing"
)

func TestBind(t *testing.T) {
	c := New()
	c.Bind("greeting", "hello")
	result := c.Make("greeting")
	if result != "hello" {
		t.Errorf("expected 'hello', got '%v'", result)
	}
}

func TestBindFactory(t *testing.T) {
	c := New()
	counter := 0
	c.Bind("counter", func(c *Container) any {
		counter++
		return counter
	})

	r1 := c.Make("counter")
	r2 := c.Make("counter")
	if r1 != 1 || r2 != 2 {
		t.Errorf("factory should create new instance each time, got %v, %v", r1, r2)
	}
}

func TestSingleton(t *testing.T) {
	c := New()
	counter := 0
	c.Singleton("service", func(c *Container) any {
		counter++
		return counter
	})

	r1 := c.Make("service")
	r2 := c.Make("service")
	if r1 != r2 {
		t.Errorf("singleton should return same instance, got %v, %v", r1, r2)
	}
}

func TestInstance(t *testing.T) {
	c := New()
	c.Instance("app", "myapp")
	if c.Make("app") != "myapp" {
		t.Error("instance not stored")
	}
}

func TestBound(t *testing.T) {
	c := New()
	if c.Bound("x") {
		t.Error("should not be bound")
	}
	c.Bind("x", "val")
	if !c.Bound("x") {
		t.Error("should be bound")
	}
}

func TestResolved(t *testing.T) {
	c := New()
	c.Bind("x", "val")
	if c.Resolved("x") {
		t.Error("should not be resolved yet")
	}
	c.Make("x")
	if !c.Resolved("x") {
		t.Error("should be resolved")
	}
}

func TestAlias(t *testing.T) {
	c := New()
	c.Bind("database.connection", "mysql")
	c.Alias("database.connection", "db")
	if c.Make("db") != "mysql" {
		t.Error("alias not working")
	}
}

func TestTag(t *testing.T) {
	c := New()
	c.Bind("report.html", "html-reporter")
	c.Bind("report.csv", "csv-reporter")
	c.Tag([]string{"report.html", "report.csv"}, "reporters")

	tagged := c.Tagged("reporters")
	if len(tagged) != 2 {
		t.Errorf("expected 2 tagged, got %d", len(tagged))
	}
}

func TestFlush(t *testing.T) {
	c := New()
	c.Bind("x", "val")
	c.Instance("y", "val2")
	c.Flush()
	if c.Bound("x") || c.Bound("y") {
		t.Error("flush should clear all bindings")
	}
}

func TestExtend(t *testing.T) {
	c := New()
	c.Bind("greeting", "hello")
	c.Extend("greeting", func(original any, c *Container) any {
		return original.(string) + " world"
	})
	if c.Make("greeting") != "hello world" {
		t.Error("extend not working")
	}
}

func TestResolving(t *testing.T) {
	c := New()
	c.Bind("x", "val")
	called := false
	c.Resolving("x", func(v any) {
		called = true
	})
	c.Make("x")
	if !called {
		t.Error("resolving callback not called")
	}
}

func TestMakePanicsOnUnbound(t *testing.T) {
	c := New()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on unbound make")
		}
	}()
	c.Make("nonexistent")
}

func TestCall(t *testing.T) {
	c := New()
	c.Instance("config.Name", "Ignite")

	fn := func(name string) string {
		return "Hello " + name
	}

	results := c.Call(fn, "World")
	if results[0] != "Hello World" {
		t.Errorf("expected 'Hello World', got '%v'", results[0])
	}
}

func TestMakeInto(t *testing.T) {
	c := New()
	c.Bind("port", 8080)
	port := MakeInto[int](c, "port")
	if port != 8080 {
		t.Errorf("expected 8080, got %d", port)
	}
}

func TestContextualBinding(t *testing.T) {
	c := New()
	c.When("UserController").Needs("Logger").Give("file-logger")
	if len(c.contextual) != 1 {
		t.Error("contextual binding not registered")
	}
}
