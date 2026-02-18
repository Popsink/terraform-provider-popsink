package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTeam(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		response := TeamRead{
			ID:          "team-123",
			Name:        "test-team",
			Description: "Test Description",
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	team := &TeamCreate{
		Name:        "test-team",
		Description: "Test Description",
	}

	result, err := client.CreateTeam(context.Background(), team)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "team-123" {
		t.Errorf("expected ID team-123, got %s", result.ID)
	}

	if result.Name != "test-team" {
		t.Errorf("expected Name test-team, got %s", result.Name)
	}
}

func TestDeleteTeam(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE request, got %s", r.Method)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	err := client.DeleteTeam(context.Background(), "team-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
