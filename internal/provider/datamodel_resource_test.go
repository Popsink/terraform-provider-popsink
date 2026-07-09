package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/popsink/terraform-provider-popsink/internal/client"
)

func TestDatamodelResourceSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewDatamodelResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", resp.Diagnostics)
	}
	for _, attr := range []string{"id", "datamodel_id", "desired_state", "name", "state"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected %s attribute", attr)
		}
	}
}

func TestDatamodelMapToStateDerivesDesiredState(t *testing.T) {
	r := &datamodelResource{}

	running := &client.DataModelRead{ID: "dm-1", Name: "orders", State: "live", Enabled: true}
	var m datamodelResourceModel
	r.mapDatamodelToState(&m, running)
	if m.DesiredState.ValueString() != datamodelStateRunning {
		t.Errorf("enabled -> %q, want running", m.DesiredState.ValueString())
	}
	if m.ID.ValueString() != "dm-1" || m.DatamodelID.ValueString() != "dm-1" {
		t.Errorf("expected id and datamodel_id set to dm-1, got id=%s datamodel_id=%s", m.ID.ValueString(), m.DatamodelID.ValueString())
	}

	stopped := &client.DataModelRead{ID: "dm-1", Enabled: false}
	var m2 datamodelResourceModel
	r.mapDatamodelToState(&m2, stopped)
	if m2.DesiredState.ValueString() != datamodelStateStopped {
		t.Errorf("disabled -> %q, want stopped", m2.DesiredState.ValueString())
	}
}

// TestReconcileDatamodelState verifies start is issued only when the datamodel
// is not already in the desired state.
func TestReconcileDatamodelState(t *testing.T) {
	t.Run("no-op when already running", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("unexpected lifecycle call: %s", r.URL.Path)
		}))
		defer server.Close()

		r := &datamodelResource{client: client.NewClient(server.URL, "tok", false)}
		cur := &client.DataModelRead{ID: "dm-1", Enabled: true}
		got, err := r.reconcileDatamodelState(context.Background(), "dm-1", types.StringValue(datamodelStateRunning), cur)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got.Enabled {
			t.Error("expected unchanged running datamodel")
		}
	})

	t.Run("stops a running datamodel", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/datamodels/dm-1/stop" {
				t.Errorf("expected stop call, got %s", r.URL.Path)
			}
			_ = json.NewEncoder(w).Encode(client.DataModelRead{ID: "dm-1", Enabled: false})
		}))
		defer server.Close()

		r := &datamodelResource{client: client.NewClient(server.URL, "tok", false)}
		cur := &client.DataModelRead{ID: "dm-1", Enabled: true}
		got, err := r.reconcileDatamodelState(context.Background(), "dm-1", types.StringValue(datamodelStateStopped), cur)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Enabled {
			t.Error("expected datamodel stopped")
		}
	})

	t.Run("no-op when desired unset", func(t *testing.T) {
		r := &datamodelResource{}
		cur := &client.DataModelRead{ID: "dm-1", Enabled: true}
		got, err := r.reconcileDatamodelState(context.Background(), "dm-1", types.StringNull(), cur)
		if err != nil || got != cur {
			t.Errorf("expected unchanged datamodel and no error, got (%+v, %v)", got, err)
		}
	})
}
