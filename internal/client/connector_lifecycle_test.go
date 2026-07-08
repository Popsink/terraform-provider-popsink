package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStartStopConnectorWorker(t *testing.T) {
	for _, tc := range []struct {
		name   string
		action string
		call   func(*Client) error
	}{
		{"start", "start", func(c *Client) error { return c.StartConnectorWorker(context.Background(), "conn-1") }},
		{"stop", "stop", func(c *Client) error { return c.StopConnectorWorker(context.Background(), "conn-1") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				want := "/connectors/conn-1/" + tc.action
				if r.Method != http.MethodPost || r.URL.Path != want {
					t.Errorf("expected POST %s, got %s %s", want, r.Method, r.URL.Path)
				}
				w.WriteHeader(http.StatusAccepted)
			}))
			defer server.Close()

			if err := tc.call(NewClient(server.URL, "test-token", false)); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestStartConnectorWorkerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"detail":"not a retention connector"}`))
	}))
	defer server.Close()

	if err := NewClient(server.URL, "test-token", false).StartConnectorWorker(context.Background(), "conn-1"); err == nil {
		t.Fatal("expected error, got nil")
	}
}
