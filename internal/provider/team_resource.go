package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/popsink/terraform-provider-popsink/internal/client"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &teamResource{}
	_ resource.ResourceWithConfigure   = &teamResource{}
	_ resource.ResourceWithImportState = &teamResource{}
)

// NewTeamResource creates a new team resource
func NewTeamResource() resource.Resource {
	return &teamResource{}
}

// teamResource defines the resource implementation
type teamResource struct {
	client *client.Client
}

// teamResourceModel describes the resource data model
type teamResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	EnvID       types.String `tfsdk:"env_id"`
}

// Metadata returns the resource type name
func (r *teamResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

// Schema defines the resource schema
func (r *teamResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Popsink team resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the team.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the team.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Short description of the team.",
				Required:    true,
			},
			"env_id": schema.StringAttribute{
				Description: "Optional environment ID the team is associated with.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *teamResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource
func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create team
	createReq := &client.TeamCreate{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	if !plan.EnvID.IsNull() && !plan.EnvID.IsUnknown() {
		envID := plan.EnvID.ValueString()
		createReq.EnvID = &envID
	}

	team, err := r.client.CreateTeam(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Team",
			fmt.Sprintf("Could not create team: %s", err.Error()),
		)
		return
	}

	// Update state with created team
	plan.ID = types.StringValue(team.ID)
	plan.Name = types.StringValue(team.Name)
	plan.Description = types.StringValue(team.Description)
	if team.EnvID != nil {
		plan.EnvID = types.StringValue(*team.EnvID)
	} else {
		plan.EnvID = types.StringNull()
	}

	tflog.Info(ctx, "Created team", map[string]any{"id": team.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the resource state
func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := r.client.GetTeam(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Team",
			fmt.Sprintf("Could not read team %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	// If team not found, remove from state
	if team == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state
	state.Name = types.StringValue(team.Name)
	state.Description = types.StringValue(team.Description)
	if team.EnvID != nil {
		state.EnvID = types.StringValue(*team.EnvID)
	} else {
		state.EnvID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource
func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan teamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build update request
	updateReq := &client.TeamUpdate{}

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}

	if !plan.Description.Equal(state.Description) {
		description := plan.Description.ValueString()
		updateReq.Description = &description
	}

	if !plan.EnvID.Equal(state.EnvID) {
		if !plan.EnvID.IsNull() && !plan.EnvID.IsUnknown() {
			envID := plan.EnvID.ValueString()
			updateReq.EnvID = &envID
		}
	}

	// Update team
	team, err := r.client.UpdateTeam(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Team",
			fmt.Sprintf("Could not update team %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	// Update state
	plan.ID = types.StringValue(team.ID)
	plan.Name = types.StringValue(team.Name)
	plan.Description = types.StringValue(team.Description)
	if team.EnvID != nil {
		plan.EnvID = types.StringValue(*team.EnvID)
	} else {
		plan.EnvID = types.StringNull()
	}

	tflog.Info(ctx, "Updated team", map[string]any{"id": team.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource
func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTeam(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Team",
			fmt.Sprintf("Could not delete team %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	tflog.Info(ctx, "Deleted team", map[string]any{"id": state.ID.ValueString()})
}

// ImportState imports the resource state
func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
