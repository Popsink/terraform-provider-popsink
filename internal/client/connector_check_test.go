package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckConnectorCredentialsSuccess(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_ = json.NewEncoder(w).Encode(CredentialsCheck{IsSuccess: true, Message: "ok"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "tok", false)
	res, err := c.CheckConnectorCredentials(context.Background(), "POSTGRES_SOURCE", map[string]any{"host": "db"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsSuccess {
		t.Errorf("expected success, got %+v", res)
	}
	if gotPath != "/postgres-source/check-credentials" {
		t.Errorf("expected path /postgres-source/check-credentials, got %s", gotPath)
	}
	if gotBody["host"] != "db" {
		t.Errorf("expected config forwarded as body, got %v", gotBody)
	}
}

func TestCheckConnectorCredentialsFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(CredentialsCheck{IsSuccess: false, Message: "authentication failed"})
	}))
	defer server.Close()

	res, err := NewClient(server.URL, "tok", false).CheckConnectorCredentials(context.Background(), "ORACLE_TARGET", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsSuccess || res.Message != "authentication failed" {
		t.Errorf("expected failure with message, got %+v", res)
	}
}

func TestCheckConnectorCredentialsUnsupported(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := NewClient(server.URL, "tok", false).CheckConnectorCredentials(context.Background(), "KAFKA_TARGET", map[string]any{})
	if !errors.Is(err, ErrCheckUnsupportedType) {
		t.Fatalf("expected ErrCheckUnsupportedType, got %v", err)
	}
}

func TestConnectorTypePath(t *testing.T) {
	cases := map[string]string{
		"POSTGRES_SOURCE":      "postgres-source",
		"UNITY_CATALOG_TARGET": "unity-catalog-target",
		"HUBSPOT_SOURCE":       "hubspot-source",
	}
	for in, want := range cases {
		if got := connectorTypePath(in); got != want {
			t.Errorf("connectorTypePath(%q) = %q, want %q", in, got, want)
		}
	}
}
