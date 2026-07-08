package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/popsink/terraform-provider-popsink/internal/client"
)

func TestEnvResourceSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewEnvResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", resp.Diagnostics)
	}
	if _, ok := resp.Schema.Attributes["retention_configuration"]; !ok {
		t.Fatal("expected retention_configuration attribute")
	}
}

// TestApplyBrokerPublicFieldsPreservesCredentials verifies the read-back mapping
// refreshes public fields while keeping write-only credential fields (which the
// API never returns) intact.
func TestApplyBrokerPublicFieldsPreservesCredentials(t *testing.T) {
	m := &brokerConfigModel{
		BootstrapServer:  types.StringValue("old:9092"),
		SecurityProtocol: types.StringValue("PLAINTEXT"),
		SaslMechanism:    types.StringValue("PLAIN"),
		SaslUsername:     types.StringValue("admin"),
		SaslPassword:     types.StringValue("secret"),
		CaCert:           types.StringValue("ca-pem"),
	}

	read := &client.BrokerConfiguration{
		BootstrapServer:  "new:9092",
		SecurityProtocol: "SASL_SSL",
		SaslMechanism:    "SCRAM-SHA-256",
		// credentials intentionally absent from read shape
	}

	applyBrokerPublicFields(m, read)

	if m.BootstrapServer.ValueString() != "new:9092" {
		t.Errorf("expected bootstrap refreshed to new:9092, got %s", m.BootstrapServer.ValueString())
	}
	if m.SecurityProtocol.ValueString() != "SASL_SSL" {
		t.Errorf("expected security_protocol refreshed, got %s", m.SecurityProtocol.ValueString())
	}
	if m.SaslUsername.ValueString() != "admin" {
		t.Errorf("expected sasl_username preserved, got %s", m.SaslUsername.ValueString())
	}
	if m.SaslPassword.ValueString() != "secret" {
		t.Errorf("expected sasl_password preserved, got %s", m.SaslPassword.ValueString())
	}
	if m.CaCert.ValueString() != "ca-pem" {
		t.Errorf("expected ca_cert preserved, got %s", m.CaCert.ValueString())
	}
}

func TestBrokerConfigToClientOmitsUnsetCredentials(t *testing.T) {
	m := &brokerConfigModel{
		BootstrapServer: types.StringValue("kafka:9092"),
		// everything else null
		SecurityProtocol: types.StringNull(),
		SaslMechanism:    types.StringNull(),
		SaslUsername:     types.StringNull(),
		SaslPassword:     types.StringNull(),
		CaCert:           types.StringNull(),
		Cert:             types.StringNull(),
		Key:              types.StringNull(),
		GroupID:          types.StringNull(),
	}

	cfg := brokerConfigToClient(m)
	if cfg.BootstrapServer != "kafka:9092" {
		t.Errorf("unexpected bootstrap server: %s", cfg.BootstrapServer)
	}
	if cfg.SaslUsername != nil {
		t.Errorf("expected nil sasl_username when unset, got %v", *cfg.SaslUsername)
	}
	if cfg.SaslPassword != nil {
		t.Errorf("expected nil sasl_password when unset, got %v", *cfg.SaslPassword)
	}
}
