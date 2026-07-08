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
