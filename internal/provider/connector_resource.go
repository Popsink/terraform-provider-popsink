package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/popsink/terraform-provider-popsink/internal/client"
)

var (
	_ resource.Resource                = &connectorResource{}
	_ resource.ResourceWithConfigure   = &connectorResource{}
	_ resource.ResourceWithImportState = &connectorResource{}
)

func NewConnectorResource() resource.Resource {
	return &connectorResource{}
}

type connectorResource struct {
	client *client.Client
}

type connectorResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	ConnectorType     types.String `tfsdk:"connector_type"`
	JsonConfiguration types.String `tfsdk:"json_configuration"`
	TeamID            types.String `tfsdk:"team_id"`
	ItemsCount        types.Int64  `tfsdk:"items_count"`
	Status            types.String `tfsdk:"status"`
}

func (r *connectorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector"
}

func (r *connectorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Popsink connector (source or target).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the connector.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the connector.",
				Required:    true,
			},
			"connector_type": schema.StringAttribute{
				Description: "The type of connector. Valid values: KAFKA_SOURCE, IBMI_SOURCE, POSTGRES_SOURCE, ORACLE_TARGET, KAFKA_TARGET, ICEBERG_TARGET, SNOWFLAKE_TARGET.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"KAFKA_SOURCE",
						"IBMI_SOURCE",
						"POSTGRES_SOURCE",
						"ORACLE_TARGET",
						"KAFKA_TARGET",
						"ICEBERG_TARGET",
						"SNOWFLAKE_TARGET",
					),
				},
			},
			"json_configuration": schema.StringAttribute{
				Description: "The connector configuration as a JSON string.",
				Required:    true,
			},
			"team_id": schema.StringAttribute{
				Description: "The ID of the team that owns this connector.",
				Required:    true,
			},
			"items_count": schema.Int64Attribute{
				Description: "The number of items associated with this connector.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Description: "The current status of the connector.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *connectorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *connectorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan connectorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var jsonConfig map[string]any
	if err := json.Unmarshal([]byte(plan.JsonConfiguration.ValueString()), &jsonConfig); err != nil {
		resp.Diagnostics.AddError("Invalid JSON Configuration", fmt.Sprintf("Could not parse json_configuration: %s", err))
		return
	}

	createReq := &client.ConnectorCreate{
		Name:              plan.Name.ValueString(),
		ConnectorType:     plan.ConnectorType.ValueString(),
		JsonConfiguration: jsonConfig,
		TeamID:            plan.TeamID.ValueString(),
	}

	connector, err := r.client.CreateConnector(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Connector", fmt.Sprintf("Could not create connector: %s", err))
		return
	}

	r.mapToState(&plan, connector)
	// Preserve the original json_configuration from the plan to avoid diff noise
	tflog.Info(ctx, "Created connector", map[string]any{"id": connector.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state connectorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connector, err := r.client.GetConnector(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Connector", fmt.Sprintf("Could not read connector %s: %s", state.ID.ValueString(), err))
		return
	}

	if connector == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Keep the plan's json_configuration to avoid unnecessary diffs from key ordering
	planJSON := state.JsonConfiguration
	r.mapToState(&state, connector)
	state.JsonConfiguration = planJSON

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *connectorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan connectorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state connectorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &client.ConnectorUpdate{}

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}

	if !plan.ConnectorType.Equal(state.ConnectorType) {
		ct := plan.ConnectorType.ValueString()
		updateReq.ConnectorType = &ct
	}

	if !plan.JsonConfiguration.Equal(state.JsonConfiguration) {
		var jsonConfig map[string]any
		if err := json.Unmarshal([]byte(plan.JsonConfiguration.ValueString()), &jsonConfig); err != nil {
			resp.Diagnostics.AddError("Invalid JSON Configuration", fmt.Sprintf("Could not parse json_configuration: %s", err))
			return
		}
		updateReq.JsonConfiguration = &jsonConfig
	}

	if !plan.TeamID.Equal(state.TeamID) {
		teamID := plan.TeamID.ValueString()
		updateReq.TeamID = &teamID
	}

	connector, err := r.client.UpdateConnector(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Connector", fmt.Sprintf("Could not update connector %s: %s", state.ID.ValueString(), err))
		return
	}

	r.mapToState(&plan, connector)
	// Preserve original json_configuration from the plan
	tflog.Info(ctx, "Updated connector", map[string]any{"id": connector.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state connectorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteConnector(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Connector", fmt.Sprintf("Could not delete connector %s: %s", state.ID.ValueString(), err))
		return
	}

	tflog.Info(ctx, "Deleted connector", map[string]any{"id": state.ID.ValueString()})
}

func (r *connectorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapToState maps a ConnectorRead to the terraform state model.
// Note: JsonConfiguration is NOT set here — callers should preserve the plan value to avoid diff noise.
func (r *connectorResource) mapToState(model *connectorResourceModel, connector *client.ConnectorRead) {
	model.ID = types.StringValue(connector.ID)
	model.Name = types.StringValue(connector.Name)
	model.ConnectorType = types.StringValue(connector.ConnectorType)
	model.TeamID = types.StringValue(connector.TeamID)
	model.ItemsCount = types.Int64Value(int64(connector.ItemsCount))
	model.Status = types.StringValue(connector.Status)
}
