package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateConnector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/connectors/" {
			t.Errorf("expected path /connectors/, got %s", r.URL.Path)
		}

		var body ConnectorCreate
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if body.Name != "test-connector" {
			t.Errorf("expected Name test-connector, got %s", body.Name)
		}

		if body.ConnectorType != "KAFKA_SOURCE" {
			t.Errorf("expected ConnectorType KAFKA_SOURCE, got %s", body.ConnectorType)
		}

		response := ConnectorRead{
			ID:                "conn-123",
			Name:              body.Name,
			ConnectorType:     body.ConnectorType,
			JsonConfiguration: body.JsonConfiguration,
			TeamID:            body.TeamID,
			ItemsCount:        0,
			Status:            "paused",
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	connector := &ConnectorCreate{
		Name:              "test-connector",
		ConnectorType:     "KAFKA_SOURCE",
		JsonConfiguration: map[string]any{"bootstrap_servers": "kafka:9092"},
		TeamID:            "team-123",
	}

	result, err := client.CreateConnector(context.Background(), connector)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "conn-123" {
		t.Errorf("expected ID conn-123, got %s", result.ID)
	}

	if result.Name != "test-connector" {
		t.Errorf("expected Name test-connector, got %s", result.Name)
	}

	if result.ConnectorType != "KAFKA_SOURCE" {
		t.Errorf("expected ConnectorType KAFKA_SOURCE, got %s", result.ConnectorType)
	}

	if result.Status != "paused" {
		t.Errorf("expected Status paused, got %s", result.Status)
	}
}

func TestGetConnector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}

		response := ConnectorRead{
			ID:                "conn-123",
			Name:              "test-connector",
			ConnectorType:     "ORACLE_TARGET",
			JsonConfiguration: map[string]any{"host": "oracle-host"},
			TeamID:            "team-123",
			ItemsCount:        5,
			Status:            "live",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	result, err := client.GetConnector(context.Background(), "conn-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.ID != "conn-123" {
		t.Errorf("expected ID conn-123, got %s", result.ID)
	}

	if result.ItemsCount != 5 {
		t.Errorf("expected ItemsCount 5, got %d", result.ItemsCount)
	}
}

func TestGetConnector_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	result, err := client.GetConnector(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil result for not found, got %v", result)
	}
}

func TestUpdateConnector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH request, got %s", r.Method)
		}

		response := ConnectorRead{
			ID:                "conn-123",
			Name:              "updated-connector",
			ConnectorType:     "KAFKA_SOURCE",
			JsonConfiguration: map[string]any{"bootstrap_servers": "kafka:9092"},
			TeamID:            "team-123",
			ItemsCount:        0,
			Status:            "paused",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	newName := "updated-connector"
	update := &ConnectorUpdate{
		Name: &newName,
	}

	result, err := client.UpdateConnector(context.Background(), "conn-123", update)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "updated-connector" {
		t.Errorf("expected Name updated-connector, got %s", result.Name)
	}
}

func TestDeleteConnector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE request, got %s", r.Method)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	err := client.DeleteConnector(context.Background(), "conn-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteConnector_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	err := client.DeleteConnector(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateConnector_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"detail":"invalid connector type"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", false)
	connector := &ConnectorCreate{
		Name:              "bad-connector",
		ConnectorType:     "INVALID",
		JsonConfiguration: map[string]any{},
		TeamID:            "team-123",
	}

	_, err := client.CreateConnector(context.Background(), connector)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
