package pocketbase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anirudh-777/pb-agent/internal/config"
)

func TestClientSendsTokenAndDecodesHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "secret-token" {
			t.Fatalf("authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":200,"message":"API is healthy.","data":{"canBackup":true}}`))
	}))
	defer server.Close()

	client := New(config.Connection{URL: server.URL, Environment: "test"}, "secret-token")
	result, err := client.Health(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result["message"] != "API is healthy." {
		t.Fatalf("unexpected health: %#v", result)
	}
}

func TestClientReturnsSafeAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"Failed to create record.","data":{"name":{"code":"required"}}}`))
	}))
	defer server.Close()

	client := New(config.Connection{URL: server.URL, Environment: "test"}, "")
	_, err := client.Request(context.Background(), http.MethodPost, "/api/example", map[string]any{})
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Status != http.StatusBadRequest {
		t.Fatalf("unexpected error: %#v", err)
	}
}
