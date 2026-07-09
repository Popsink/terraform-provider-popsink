package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/popsink/terraform-provider-popsink/internal/client"
)

const (
	teamRoleMember = "member"
	teamRoleOwner  = "owner"
)

var (
	_ resource.Resource                = &teamMemberResource{}
	_ resource.ResourceWithConfigure   = &teamMemberResource{}
	_ resource.ResourceWithImportState = &teamMemberResource{}
)

// NewTeamMemberResource creates a new team member resource.
func NewTeamMemberResource() resource.Resource {
	return &teamMemberResource{}
}

type teamMemberResource struct {
	client *client.Client
}

type teamMemberResourceModel struct {
	ID     types.String `tfsdk:"id"`
	TeamID types.String `tfsdk:"team_id"`
	UserID types.String `tfsdk:"user_id"`
	Role   types.String `tfsdk:"role"`
}

func (r *teamMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_member"
}

func (r *teamMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages membership of a user in a Popsink team. There is no update endpoint for a " +
			"membership, so changing the team, user, or role replaces the resource (remove + re-add).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The membership entry identifier (distinct from the user ID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrTeamID: schema.StringAttribute{
				Description: "The ID of the team. Changing this forces a new membership.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the user to add to the team. Changing this forces a new membership.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Description: "The member's role: \"member\" or \"owner\" (owners have admin privileges). " +
					"Defaults to \"member\". Changing this forces a new membership (no update endpoint).",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf(teamRoleMember, teamRoleOwner),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *teamMemberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *teamMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role := teamRoleMember
	if !plan.Role.IsNull() && !plan.Role.IsUnknown() {
		role = plan.Role.ValueString()
	}

	teamID := plan.TeamID.ValueString()
	userID := plan.UserID.ValueString()

	body := &client.TeamMemberBulkCreate{Owners: []string{}, Members: []string{}}
	if role == teamRoleOwner {
		body.Owners = []string{userID}
	} else {
		body.Members = []string{userID}
	}

	if err := r.client.BulkCreateTeamMembers(ctx, teamID, body); err != nil {
		resp.Diagnostics.AddError("Error Adding Team Member", fmt.Sprintf("Could not add user %s to team %s: %s", userID, teamID, err))
		return
	}

	member, err := r.client.FindTeamMemberByUser(ctx, teamID, userID)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Team Member", fmt.Sprintf("Could not read membership after create: %s", err))
		return
	}
	if member == nil {
		resp.Diagnostics.AddError(
			"Team Member Not Created",
			fmt.Sprintf("User %s was not found in team %s after the add request. Verify the user ID and your permissions.", userID, teamID),
		)
		return
	}

	plan.ID = types.StringValue(member.ID)
	plan.Role = types.StringValue(roleFromAdmin(member.Admin))
	tflog.Info(ctx, "Added team member", map[string]any{attrTeamID: teamID, "member_id": member.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	member, err := r.client.GetTeamMember(ctx, state.TeamID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Team Member", fmt.Sprintf("Could not read membership %s: %s", state.ID.ValueString(), err))
		return
	}
	if member == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.UserID = types.StringValue(member.UserID)
	state.Role = types.StringValue(roleFromAdmin(member.Admin))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update should never be called: every attribute is RequiresReplace. It is
// implemented to satisfy the interface and is a no-op passthrough.
func (r *teamMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan teamMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teamMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteTeamMember(ctx, state.TeamID.ValueString(), state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Removing Team Member", fmt.Sprintf("Could not remove membership %s from team %s: %s", state.ID.ValueString(), state.TeamID.ValueString(), err))
		return
	}

	tflog.Info(ctx, "Removed team member", map[string]any{attrTeamID: state.TeamID.ValueString(), "member_id": state.ID.ValueString()})
}

// ImportState accepts a composite ID of the form "team_id/member_id".
func (r *teamMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in the form \"team_id/member_id\", got %q.", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(attrTeamID), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func roleFromAdmin(admin bool) string {
	if admin {
		return teamRoleOwner
	}
	return teamRoleMember
}
