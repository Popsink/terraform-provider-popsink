package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/popsink/terraform-provider-popsink/internal/client"
)

func TestDeriveDesiredState(t *testing.T) {
	cases := map[string]string{
		"paused":   connectorStateStopped,
		"live":     connectorStateRunning,
		"building": connectorStateRunning,
		"error":    connectorStateRunning,
	}
	for status, want := range cases {
		if got := deriveDesiredState(status); got != want {
			t.Errorf("deriveDesiredState(%q) = %q, want %q", status, got, want)
		}
	}
}

func TestResolveStateTimeout(t *testing.T) {
	if d, err := resolveStateTimeout(types.StringNull()); err != nil || d != defaultConnectorStateTimeout {
		t.Errorf("null: got (%v,%v), want default", d, err)
	}
	if d, err := resolveStateTimeout(types.StringValue("90s")); err != nil || d != 90*time.Second {
		t.Errorf("90s: got (%v,%v)", d, err)
	}
	if _, err := resolveStateTimeout(types.StringValue("nonsense")); err == nil {
		t.Error("expected error for invalid duration")
	}
}

func TestNormalizeStateTimeout(t *testing.T) {
	if v := normalizeStateTimeout(types.StringNull()); v.ValueString() != "5m" {
		t.Errorf("null -> %q, want 5m", v.ValueString())
	}
	if v := normalizeStateTimeout(types.StringValue("300s")); v.ValueString() != "300s" {
		t.Errorf("expected literal preserved, got %q", v.ValueString())
	}
}

// TestReconcileConnectorState_StartsAndConverges drives a fake data-plane that
// reports "paused" until /start is called, then "building", then "live".
func TestReconcileConnectorState_StartsAndConverges(t *testing.T) {
	var started bool
	polls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/connectors/c1/start":
			started = true
			w.WriteHeader(http.StatusAccepted)
		case r.Method == http.MethodGet && r.URL.Path == "/connectors/c1":
			status := "paused"
			if started {
				polls++
				if polls >= 2 {
					status = "live"
				} else {
					status = "building"
				}
			}
			_ = json.NewEncoder(w).Encode(client.ConnectorRead{ID: "c1", Status: status})
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	r := &connectorResource{client: client.NewClient(server.URL, "tok", false), pollInterval: time.Millisecond}
	final, err := r.reconcileConnectorState(context.Background(), "c1", connectorStateRunning, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if final != "live" {
		t.Errorf("expected final status live, got %q", final)
	}
	if !started {
		t.Error("expected start to have been called")
	}
}

// TestReconcileConnectorState_NoOpWhenAlreadyStopped verifies a stop is not
// issued when the worker is already paused.
func TestReconcileConnectorState_NoOpWhenAlreadyStopped(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			t.Errorf("unexpected lifecycle call: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(client.ConnectorRead{ID: "c1", Status: "paused"})
	}))
	defer server.Close()

	r := &connectorResource{client: client.NewClient(server.URL, "tok", false), pollInterval: time.Millisecond}
	final, err := r.reconcileConnectorState(context.Background(), "c1", connectorStateStopped, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if final != "paused" {
		t.Errorf("expected paused, got %q", final)
	}
}

// TestReconcileConnectorState_Timeout ensures a non-converging worker surfaces a
// timeout error.
func TestReconcileConnectorState_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		// Never converges to paused.
		_ = json.NewEncoder(w).Encode(client.ConnectorRead{ID: "c1", Status: "live"})
	}))
	defer server.Close()

	r := &connectorResource{client: client.NewClient(server.URL, "tok", false), pollInterval: time.Millisecond}
	_, err := r.reconcileConnectorState(context.Background(), "c1", connectorStateStopped, 20*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}
