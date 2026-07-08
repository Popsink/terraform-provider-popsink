package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// TestConnectorResourceJSONConfigurationIsSensitive locks in that credentials
// carried by json_configuration are redacted from plan output and logs (#26).
func TestConnectorResourceJSONConfigurationIsSensitive(t *testing.T) {
	r := NewConnectorResource()

	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", resp.Diagnostics)
	}

	attr, ok := resp.Schema.Attributes["json_configuration"]
	if !ok {
		t.Fatal("expected json_configuration attribute in connector schema")
	}

	if !attr.IsSensitive() {
		t.Error("expected json_configuration to be marked Sensitive")
	}
}

// TestConnectorTypesCoversDataPlaneEnum locks in the full set of connector
// types the provider accepts (#24). Keep this in sync with the data-plane
// ConnectorType enum; the count guard makes accidental drift visible.
func TestConnectorTypesCoversDataPlaneEnum(t *testing.T) {
	const want = 28
	if got := len(connectorTypes); got != want {
		t.Fatalf("expected %d connector types, got %d", want, got)
	}

	seen := make(map[string]bool, len(connectorTypes))
	for _, ct := range connectorTypes {
		if seen[ct] {
			t.Errorf("duplicate connector type: %s", ct)
		}
		seen[ct] = true
	}

	// Spot-check types that were previously missing (incl. SALESFORCE_SOURCE,
	// which the tracking issue itself omitted).
	for _, ct := range []string{
		"ORACLE_SOURCE", "MSSQL_SOURCE", "MYSQL_SOURCE", "SALESFORCE_SOURCE",
		"BIGQUERY_SOURCE", "SNOWFLAKE_SOURCE", "UNITY_CATALOG_TARGET",
		"POSTGRES_TARGET", "ELASTICSEARCH_TARGET", "BIGQUERY_TARGET", "WEBHOOK_TARGET",
	} {
		if !seen[ct] {
			t.Errorf("expected connector type %s to be supported", ct)
		}
	}
}

// TestConnectorResourceSchemaBuilds ensures the schema (which references
// connectorTypes in its validator) is well-formed.
func TestConnectorResourceSchemaBuilds(t *testing.T) {
	var resp resource.SchemaResponse
	NewConnectorResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", resp.Diagnostics)
	}
	if _, ok := resp.Schema.Attributes["connector_type"]; !ok {
		t.Fatal("expected connector_type attribute in connector schema")
	}
}
