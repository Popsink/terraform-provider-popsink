package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateSubscription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/subscriptions/" {
			t.Errorf("expected POST /subscriptions/, got %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "sub-123"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token", false)
	id, err := c.CreateSubscription(context.Background(), &SubscriptionCreate{
		Name:              "orders-to-snowflake",
		DatamodelID:       "dm-1",
		TargetConnectorID: "conn-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "sub-123" {
		t.Errorf("expected id sub-123, got %s", id)
	}
}

func TestGetSubscription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/subscriptions/sub-123" {
			t.Errorf("expected path /subscriptions/sub-123, got %s", r.URL.Path)
		}
		dm := "dm-1"
		target := "conn-1"
		_ = json.NewEncoder(w).Encode(SubscriptionRead{
			ID:              "sub-123",
			Name:            "orders",
			Status:          "live",
			DatamodelID:     &dm,
			TargetID:        &target,
			TargetTableName: "root",
			Enabled:         true,
			ConfigHash:      "abc123",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token", false)
	sub, err := c.GetSubscription(context.Background(), "sub-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub == nil || sub.ID != "sub-123" {
		t.Fatalf("expected sub-123, got %+v", sub)
	}
	if sub.TargetID == nil || *sub.TargetID != "conn-1" {
		t.Errorf("expected targetId conn-1, got %v", sub.TargetID)
	}
	if !sub.Enabled {
		t.Error("expected enabled=true")
	}
}

func TestGetSubscriptionNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token", false)
	sub, err := c.GetSubscription(context.Background(), "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub != nil {
		t.Errorf("expected nil on 404, got %+v", sub)
	}
}

func TestUpdateSubscription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "sub-123"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token", false)
	name := "renamed"
	if err := c.UpdateSubscription(context.Background(), "sub-123", &SubscriptionUpdate{Name: &name}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteSubscription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token", false)
	if err := c.DeleteSubscription(context.Background(), "sub-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubscriptionLifecycle(t *testing.T) {
	for _, tc := range []struct {
		action  string
		enabled bool
		call    func(*Client) (*SubscriptionRead, error)
	}{
		{"start", true, func(c *Client) (*SubscriptionRead, error) {
			return c.StartSubscription(context.Background(), "sub-123")
		}},
		{"pause", false, func(c *Client) (*SubscriptionRead, error) {
			return c.PauseSubscription(context.Background(), "sub-123")
		}},
	} {
		t.Run(tc.action, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				want := "/subscriptions/sub-123/" + tc.action
				if r.URL.Path != want {
					t.Errorf("expected path %s, got %s", want, r.URL.Path)
				}
				w.WriteHeader(http.StatusAccepted)
				_ = json.NewEncoder(w).Encode(SubscriptionRead{ID: "sub-123", Enabled: tc.enabled})
			}))
			defer server.Close()

			c := NewClient(server.URL, "test-token", false)
			sub, err := tc.call(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sub.Enabled != tc.enabled {
				t.Errorf("expected enabled=%v, got %v", tc.enabled, sub.Enabled)
			}
		})
	}
}
