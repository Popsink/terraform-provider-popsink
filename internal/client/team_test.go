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

func TestGetTeam(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}

		if r.URL.Path != "/teams/team-123" {
			t.Errorf("expected path /teams/team-123, got %s", r.URL.Path)
		}

		response := TeamRead{
			ID:          "team-123",
			Name:        "test-team",
			Description: "Test Description",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	result, err := client.GetTeam(context.Background(), "team-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.ID != "team-123" {
		t.Errorf("expected ID team-123, got %s", result.ID)
	}

	if result.Description != "Test Description" {
		t.Errorf("expected Description 'Test Description', got %s", result.Description)
	}
}

func TestGetTeam_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	result, err := client.GetTeam(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil result for not found, got %v", result)
	}
}

func TestUpdateTeam(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH request, got %s", r.Method)
		}

		if r.URL.Path != "/teams/team-123" {
			t.Errorf("expected path /teams/team-123, got %s", r.URL.Path)
		}

		var body TeamUpdate
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if body.Name == nil || *body.Name != "updated-team" {
			t.Errorf("expected Name updated-team, got %v", body.Name)
		}

		response := TeamRead{
			ID:          "team-123",
			Name:        *body.Name,
			Description: "Test Description",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	updatedName := "updated-team"
	update := &TeamUpdate{
		Name: &updatedName,
	}

	result, err := client.UpdateTeam(context.Background(), "team-123", update)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != updatedName {
		t.Errorf("expected Name %s, got %s", updatedName, result.Name)
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

func TestDeleteTeam_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	err := client.DeleteTeam(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
