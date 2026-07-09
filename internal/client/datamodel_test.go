package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetDataModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/datamodels/dm-1" {
			t.Errorf("expected /datamodels/dm-1, got %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(DataModelRead{ID: "dm-1", Name: "orders", State: "live", Enabled: true})
	}))
	defer server.Close()

	dm, err := NewClient(server.URL, "tok", false).GetDataModel(context.Background(), "dm-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dm == nil || dm.Name != "orders" || !dm.Enabled {
		t.Fatalf("unexpected datamodel: %+v", dm)
	}
}

func TestGetDataModelNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	dm, err := NewClient(server.URL, "tok", false).GetDataModel(context.Background(), "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dm != nil {
		t.Errorf("expected nil for 404, got %+v", dm)
	}
}

func TestDataModelStartStop(t *testing.T) {
	for _, tc := range []struct {
		action  string
		enabled bool
		call    func(*Client) (*DataModelRead, error)
	}{
		{"start", true, func(c *Client) (*DataModelRead, error) { return c.StartDataModel(context.Background(), "dm-1") }},
		{"stop", false, func(c *Client) (*DataModelRead, error) { return c.StopDataModel(context.Background(), "dm-1") }},
	} {
		t.Run(tc.action, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				want := "/datamodels/dm-1/" + tc.action
				if r.Method != http.MethodPost || r.URL.Path != want {
					t.Errorf("expected POST %s, got %s %s", want, r.Method, r.URL.Path)
				}
				w.WriteHeader(http.StatusAccepted)
				_ = json.NewEncoder(w).Encode(DataModelRead{ID: "dm-1", Enabled: tc.enabled})
			}))
			defer server.Close()

			dm, err := tc.call(NewClient(server.URL, "tok", false))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if dm.Enabled != tc.enabled {
				t.Errorf("expected enabled=%v, got %v", tc.enabled, dm.Enabled)
			}
		})
	}
}
