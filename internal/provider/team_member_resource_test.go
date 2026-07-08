package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestTeamMemberResourceSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewTeamMemberResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", resp.Diagnostics)
	}
	for _, attr := range []string{"id", "team_id", "user_id", "role"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected %s attribute", attr)
		}
	}
}

func TestRoleFromAdmin(t *testing.T) {
	if got := roleFromAdmin(true); got != teamRoleOwner {
		t.Errorf("admin=true -> %q, want owner", got)
	}
	if got := roleFromAdmin(false); got != teamRoleMember {
		t.Errorf("admin=false -> %q, want member", got)
	}
}

func TestTeamMemberImportStateInvalid(t *testing.T) {
	r := NewTeamMemberResource().(*teamMemberResource)
	for _, id := range []string{"", "onlyone", "a/b/c", "/b", "a/"} {
		var resp resource.ImportStateResponse
		r.ImportState(context.Background(), resource.ImportStateRequest{ID: id}, &resp)
		if !resp.Diagnostics.HasError() {
			t.Errorf("expected error for import id %q", id)
		}
	}
}
