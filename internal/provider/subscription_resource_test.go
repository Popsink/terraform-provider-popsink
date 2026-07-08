package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/popsink/terraform-provider-popsink/internal/client"
)

func TestSubscriptionResourceSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewSubscriptionResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", resp.Diagnostics)
	}
	for _, attr := range []string{"datamodel_id", "target_connector_id", "desired_state", "smt_config", "config_hash"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected %s attribute in subscription schema", attr)
		}
	}
}

func TestParseSmtConfig(t *testing.T) {
	// Null / empty -> nil, no error.
	if out, err := parseSmtConfig(types.StringNull()); err != nil || out != nil {
		t.Errorf("expected (nil,nil) for null, got (%v,%v)", out, err)
	}
	if out, err := parseSmtConfig(types.StringValue("")); err != nil || out != nil {
		t.Errorf("expected (nil,nil) for empty, got (%v,%v)", out, err)
	}
	// Valid JSON array.
	out, err := parseSmtConfig(types.StringValue(`[{"function":"cast","column":"id"}]`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out[0]["function"] != "cast" {
		t.Errorf("unexpected parse result: %v", out)
	}
	// Invalid JSON -> error.
	if _, err := parseSmtConfig(types.StringValue("{not json")); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// TestMapToStateDerivesDesiredStateFromEnabled verifies desired_state tracks the
// server-side enabled flag (enabled -> running, disabled -> paused) and that
// smt_config is left untouched by the read mapping.
func TestMapToStateDerivesDesiredStateFromEnabled(t *testing.T) {
	r := &subscriptionResource{}

	dm, target := "dm-1", "conn-1"
	base := client.SubscriptionRead{
		ID:              "sub-123",
		Name:            "orders",
		Status:          "live",
		DatamodelID:     &dm,
		TargetID:        &target,
		TargetTableName: "root",
		ConfigHash:      "hash-1",
	}

	running := base
	running.Enabled = true
	model := &subscriptionResourceModel{SmtConfig: types.StringValue(`[{"a":1}]`)}
	r.mapToState(model, &running)
	if model.DesiredState.ValueString() != desiredStateRunning {
		t.Errorf("expected desired_state running, got %s", model.DesiredState.ValueString())
	}
	if model.TargetConnectorID.ValueString() != "conn-1" {
		t.Errorf("expected target_connector_id from targetId, got %s", model.TargetConnectorID.ValueString())
	}
	if model.SmtConfig.ValueString() != `[{"a":1}]` {
		t.Errorf("expected smt_config preserved, got %s", model.SmtConfig.ValueString())
	}

	paused := base
	paused.Enabled = false
	model2 := &subscriptionResourceModel{}
	r.mapToState(model2, &paused)
	if model2.DesiredState.ValueString() != desiredStatePaused {
		t.Errorf("expected desired_state paused, got %s", model2.DesiredState.ValueString())
	}
}
