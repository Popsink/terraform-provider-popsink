package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/popsink/terraform-provider-popsink/internal/client"
)

// configureDataSourceClient extracts the shared API client from provider data.
func configureDataSourceClient(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) *client.Client {
	if req.ProviderData == nil {
		return nil
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return nil
	}
	return c
}

// nullableStr converts an optional client string into a types.String.
func nullableStr(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

// ---------------------------------------------------------------------------
// popsink_team
// ---------------------------------------------------------------------------

var (
	_ datasource.DataSource              = &teamDataSource{}
	_ datasource.DataSourceWithConfigure = &teamDataSource{}
)

// NewTeamDataSource creates the popsink_team data source.
func NewTeamDataSource() datasource.DataSource { return &teamDataSource{} }

type teamDataSource struct{ client *client.Client }

type teamDataSourceModel struct {
	Name        types.String `tfsdk:"name"`
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	EnvID       types.String `tfsdk:"env_id"`
}

func (d *teamDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (d *teamDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing Popsink team by name.",
		Attributes: map[string]schema.Attribute{
			"name":        schema.StringAttribute{Description: "The name of the team to look up.", Required: true},
			"id":          schema.StringAttribute{Description: "The unique identifier of the team.", Computed: true},
			"description": schema.StringAttribute{Description: "The team description.", Computed: true},
			"env_id":      schema.StringAttribute{Description: "The environment the team belongs to.", Computed: true},
		},
	}
}

func (d *teamDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSourceClient(req, resp)
}

func (d *teamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg teamDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := d.client.FindTeamByName(ctx, cfg.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Team", fmt.Sprintf("Could not look up team %q: %s", cfg.Name.ValueString(), err))
		return
	}
	if team == nil {
		resp.Diagnostics.AddError("Team Not Found", fmt.Sprintf("No team found with name %q.", cfg.Name.ValueString()))
		return
	}

	cfg.ID = types.StringValue(team.ID)
	cfg.Name = types.StringValue(team.Name)
	cfg.Description = types.StringValue(team.Description)
	cfg.EnvID = nullableStr(team.EnvID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

// ---------------------------------------------------------------------------
// popsink_connector
// ---------------------------------------------------------------------------

var (
	_ datasource.DataSource              = &connectorDataSource{}
	_ datasource.DataSourceWithConfigure = &connectorDataSource{}
)

// NewConnectorDataSource creates the popsink_connector data source.
func NewConnectorDataSource() datasource.DataSource { return &connectorDataSource{} }

type connectorDataSource struct{ client *client.Client }

type connectorDataSourceModel struct {
	Name          types.String `tfsdk:"name"`
	ID            types.String `tfsdk:"id"`
	ConnectorType types.String `tfsdk:"connector_type"`
	TeamID        types.String `tfsdk:"team_id"`
}

func (d *connectorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector"
}

func (d *connectorDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing Popsink connector by name.",
		Attributes: map[string]schema.Attribute{
			"name":           schema.StringAttribute{Description: "The name of the connector to look up.", Required: true},
			"id":             schema.StringAttribute{Description: "The unique identifier of the connector.", Computed: true},
			"connector_type": schema.StringAttribute{Description: "The connector type.", Computed: true},
			"team_id":        schema.StringAttribute{Description: "The team that owns the connector.", Computed: true},
		},
	}
}

func (d *connectorDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSourceClient(req, resp)
}

func (d *connectorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg connectorDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn, err := d.client.FindConnectorByName(ctx, cfg.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Connector", fmt.Sprintf("Could not look up connector %q: %s", cfg.Name.ValueString(), err))
		return
	}
	if conn == nil {
		resp.Diagnostics.AddError("Connector Not Found", fmt.Sprintf("No connector found with name %q.", cfg.Name.ValueString()))
		return
	}

	cfg.ID = types.StringValue(conn.ID)
	cfg.Name = types.StringValue(conn.Name)
	cfg.ConnectorType = types.StringValue(conn.ConnectorType)
	cfg.TeamID = types.StringValue(conn.TeamID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

// ---------------------------------------------------------------------------
// popsink_env
// ---------------------------------------------------------------------------

var (
	_ datasource.DataSource              = &envDataSource{}
	_ datasource.DataSourceWithConfigure = &envDataSource{}
)

// NewEnvDataSource creates the popsink_env data source.
func NewEnvDataSource() datasource.DataSource { return &envDataSource{} }

type envDataSource struct{ client *client.Client }

type envDataSourceModel struct {
	Name types.String `tfsdk:"name"`
	ID   types.String `tfsdk:"id"`
}

func (d *envDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env"
}

func (d *envDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing Popsink environment by name.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{Description: "The name of the environment to look up.", Required: true},
			"id":   schema.StringAttribute{Description: "The unique identifier of the environment.", Computed: true},
		},
	}
}

func (d *envDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSourceClient(req, resp)
}

func (d *envDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg envDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, err := d.client.FindEnvByName(ctx, cfg.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Environment", fmt.Sprintf("Could not look up environment %q: %s", cfg.Name.ValueString(), err))
		return
	}
	if env == nil {
		resp.Diagnostics.AddError("Environment Not Found", fmt.Sprintf("No environment found with name %q.", cfg.Name.ValueString()))
		return
	}

	cfg.ID = types.StringValue(env.ID)
	cfg.Name = types.StringValue(env.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

// ---------------------------------------------------------------------------
// popsink_pipeline
// ---------------------------------------------------------------------------

var (
	_ datasource.DataSource              = &pipelineDataSource{}
	_ datasource.DataSourceWithConfigure = &pipelineDataSource{}
)

// NewPipelineDataSource creates the popsink_pipeline data source.
func NewPipelineDataSource() datasource.DataSource { return &pipelineDataSource{} }

type pipelineDataSource struct{ client *client.Client }

type pipelineDataSourceModel struct {
	Name types.String `tfsdk:"name"`
	ID   types.String `tfsdk:"id"`
}

func (d *pipelineDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline"
}

func (d *pipelineDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing Popsink pipeline by name.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{Description: "The name of the pipeline to look up.", Required: true},
			"id":   schema.StringAttribute{Description: "The unique identifier of the pipeline.", Computed: true},
		},
	}
}

func (d *pipelineDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSourceClient(req, resp)
}

func (d *pipelineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg pipelineDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pipeline, err := d.client.FindPipelineByName(ctx, cfg.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Pipeline", fmt.Sprintf("Could not look up pipeline %q: %s", cfg.Name.ValueString(), err))
		return
	}
	if pipeline == nil {
		resp.Diagnostics.AddError("Pipeline Not Found", fmt.Sprintf("No pipeline found with name %q.", cfg.Name.ValueString()))
		return
	}

	cfg.ID = types.StringValue(pipeline.ID)
	cfg.Name = types.StringValue(pipeline.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}
