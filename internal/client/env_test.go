package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateEnv(t *testing.T) {
	var received EnvCreate
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/envs/" {
			t.Errorf("expected path /envs/, got %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)

		// Read shape strips credentials.
		response := EnvRead{
			ID:   "env-123",
			Name: "production",
			RetentionConfiguration: BrokerConfiguration{
				BootstrapServer:  "kafka.example.com:9092",
				SecurityProtocol: "SASL_SSL",
				SaslMechanism:    "SCRAM-SHA-256",
			},
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token", false)
	user, pass := "admin", "secret"
	result, err := c.CreateEnv(context.Background(), &EnvCreate{
		Name: "production",
		RetentionConfiguration: BrokerConfiguration{
			BootstrapServer:  "kafka.example.com:9092",
			SecurityProtocol: "SASL_SSL",
			SaslMechanism:    "SCRAM-SHA-256",
			SaslUsername:     &user,
			SaslPassword:     &pass,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "env-123" {
		t.Errorf("expected ID env-123, got %s", result.ID)
	}
	// The credential fields must be sent to the server on create.
	if received.RetentionConfiguration.SaslUsername == nil || *received.RetentionConfiguration.SaslUsername != "admin" {
		t.Errorf("expected sasl_username to be sent, got %v", received.RetentionConfiguration.SaslUsername)
	}
	if received.RetentionConfiguration.SaslPassword == nil || *received.RetentionConfiguration.SaslPassword != "secret" {
		t.Errorf("expected sasl_password to be sent, got %v", received.RetentionConfiguration.SaslPassword)
	}
}

func TestGetEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/envs/env-123" {
			t.Errorf("expected path /envs/env-123, got %s", r.URL.Path)
		}
		response := EnvRead{
			ID:   "env-123",
			Name: "production",
			RetentionConfiguration: BrokerConfiguration{
				BootstrapServer:  "kafka.example.com:9092",
				SecurityProtocol: "PLAINTEXT",
				SaslMechanism:    "PLAIN",
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token", false)
	result, err := c.GetEnv(context.Background(), "env-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected env, got nil")
	}
	if result.RetentionConfiguration.BootstrapServer != "kafka.example.com:9092" {
		t.Errorf("unexpected bootstrap server: %s", result.RetentionConfiguration.BootstrapServer)
	}
}

func TestGetEnvNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token", false)
	result, err := c.GetEnv(context.Background(), "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for 404, got %+v", result)
	}
}

func TestUpdateEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH request, got %s", r.Method)
		}
		response := EnvRead{
			ID:   "env-123",
			Name: "staging",
			RetentionConfiguration: BrokerConfiguration{
				BootstrapServer:  "kafka.example.com:9092",
				SecurityProtocol: "PLAINTEXT",
				SaslMechanism:    "PLAIN",
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token", false)
	name := "staging"
	result, err := c.UpdateEnv(context.Background(), "env-123", &EnvUpdate{Name: &name})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "staging" {
		t.Errorf("expected name staging, got %s", result.Name)
	}
}

func TestDeleteEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token", false)
	if err := c.DeleteEnv(context.Background(), "env-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteEnvNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token", false)
	if err := c.DeleteEnv(context.Background(), "missing"); err != nil {
		t.Fatalf("expected nil error on 404 delete, got %v", err)
	}
}
