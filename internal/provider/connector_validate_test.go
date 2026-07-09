package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/popsink/terraform-provider-popsink/internal/client"
)

func newConnectorResourceWithServer(url string) *connectorResource {
	return &connectorResource{client: client.NewClient(url, "tok", false)}
}

func TestValidateCredentials(t *testing.T) {
	t.Run("success emits no diagnostics", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(client.CredentialsCheck{IsSuccess: true, Message: "ok"})
		}))
		defer server.Close()

		var diags diag.Diagnostics
		newConnectorResourceWithServer(server.URL).validateCredentials(context.Background(), "POSTGRES_SOURCE", map[string]any{}, &diags)
		if diags.HasError() {
			t.Errorf("expected no diagnostics, got %v", diags)
		}
	})

	t.Run("failure surfaces the API message", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(client.CredentialsCheck{IsSuccess: false, Message: "bad password"})
		}))
		defer server.Close()

		var diags diag.Diagnostics
		newConnectorResourceWithServer(server.URL).validateCredentials(context.Background(), "POSTGRES_SOURCE", map[string]any{}, &diags)
		if !diags.HasError() {
			t.Fatal("expected an error diagnostic")
		}
		if got := diags.Errors()[0].Detail(); got != "bad password" {
			t.Errorf("expected detail 'bad password', got %q", got)
		}
	})

	t.Run("unsupported type is a clear error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		var diags diag.Diagnostics
		newConnectorResourceWithServer(server.URL).validateCredentials(context.Background(), "KAFKA_TARGET", map[string]any{}, &diags)
		if !diags.HasError() {
			t.Fatal("expected an error diagnostic")
		}
		if got := diags.Errors()[0].Summary(); got != "Credential Validation Unsupported" {
			t.Errorf("unexpected summary: %q", got)
		}
	})
}
