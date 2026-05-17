package httpclient_test

import (
	"fmt"
	"github.com/sazzad/ignite/httpclient"
	"net/http"
	"net/http/httptest"
)

func ExampleClient_Get() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	resp, _ := httpclient.Http.Get(server.URL)
	fmt.Println(resp.Status())
	// Output: 200
}

func ExampleClient_WithToken() {
	client := httpclient.New().
		WithToken("my-secret-token").
		AcceptJSON()

	_ = client // Use client for authenticated requests
}

func ExampleClient_AsJSON() {
	client := httpclient.New().AsJSON()

	data := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	}

	_, _ = client.Post("https://api.example.com/users", data)
}

func ExampleClient_Retry() {
	client := httpclient.New().
		Retry(3, 100) // Retry 3 times with 100ms delay

	_, _ = client.Get("https://api.example.com/data")
}

func ExampleFake() {
	fake := httpclient.Fake(
		httpclient.FakeResponse(map[string]string{"status": "ok"}, 200),
	)

	resp, _ := fake.Get("https://api.example.com/test")
	data, _ := resp.JSON()

	fmt.Println(data["status"])
	// Output: ok
}
