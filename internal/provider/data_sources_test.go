package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestDataSourceSchemas(t *testing.T) {
	cases := map[string]datasource.DataSource{
		"team":      NewTeamDataSource(),
		"connector": NewConnectorDataSource(),
		"env":       NewEnvDataSource(),
		"pipeline":  NewPipelineDataSource(),
	}
	for name, ds := range cases {
		t.Run(name, func(t *testing.T) {
			var resp datasource.SchemaResponse
			ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected schema diagnostics: %v", resp.Diagnostics)
			}
			if _, ok := resp.Schema.Attributes["name"]; !ok {
				t.Error("expected a required name attribute")
			}
			if _, ok := resp.Schema.Attributes["id"]; !ok {
				t.Error("expected a computed id attribute")
			}
		})
	}
}

func TestDataSourceMetadata(t *testing.T) {
	cases := map[string]struct {
		ds   datasource.DataSource
		want string
	}{
		"team":      {NewTeamDataSource(), "popsink_team"},
		"connector": {NewConnectorDataSource(), "popsink_connector"},
		"env":       {NewEnvDataSource(), "popsink_env"},
		"pipeline":  {NewPipelineDataSource(), "popsink_pipeline"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var resp datasource.MetadataResponse
			tc.ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "popsink"}, &resp)
			if resp.TypeName != tc.want {
				t.Errorf("expected %s, got %s", tc.want, resp.TypeName)
			}
		})
	}
}
