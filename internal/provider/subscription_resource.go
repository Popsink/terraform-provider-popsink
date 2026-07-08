package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/popsink/terraform-provider-popsink/internal/client"
)

const (
	desiredStateRunning = "running"
	desiredStatePaused  = "paused"
)

var (
	_ resource.Resource                = &subscriptionResource{}
	_ resource.ResourceWithConfigure   = &subscriptionResource{}
	_ resource.ResourceWithImportState = &subscriptionResource{}
)

// NewSubscriptionResource creates a new subscription resource.
func NewSubscriptionResource() resource.Resource {
	return &subscriptionResource{}
}

type subscriptionResource struct {
	client *client.Client
}

type subscriptionResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	DatamodelID        types.String `tfsdk:"datamodel_id"`
	TargetConnectorID  types.String `tfsdk:"target_connector_id"`
	SmtConfig          types.String `tfsdk:"smt_config"`
	ConsumerID         types.String `tfsdk:"consumer_id"`
	ErrorTableEnabled  types.Bool   `tfsdk:"error_table_enabled"`
	ErrorTableName     types.String `tfsdk:"error_table_name"`
	ErrorTableTargetID types.String `tfsdk:"error_table_target_id"`
	TargetTableName    types.String `tfsdk:"target_table_name"`
	Backfill           types.Bool   `tfsdk:"backfill"`
	DesiredState       types.String `tfsdk:"desired_state"`
	Status             types.String `tfsdk:"status"`
	ConfigHash         types.String `tfsdk:"config_hash"`
}

func (r *subscriptionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription"
}

func (r *subscriptionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Popsink subscription. A subscription maps a datamodel to a target " +
			"connector (optionally through SMT transformations) and is managed independently of the " +
			"pipeline composite.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the subscription.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the subscription.",
				Required:    true,
			},
			"datamodel_id": schema.StringAttribute{
				Description: "The ID of the datamodel this subscription reads from. Changing this forces a new subscription.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_connector_id": schema.StringAttribute{
				Description: "The ID of the target connector this subscription delivers to. Changing this forces a new subscription.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"smt_config": schema.StringAttribute{
				Description: "SMT (single-message transform) chain configuration as a JSON array string. " +
					"Treated opaquely in v1. Sent to the API's SMT/mapper configuration; the provider keeps " +
					"the configured value in state and uses config_hash for drift detection.",
				Optional: true,
			},
			"consumer_id": schema.StringAttribute{
				Description: "Optional consumer ID for the subscription.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"error_table_enabled": schema.BoolAttribute{
				Description: "Whether an error table is enabled for this subscription.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"error_table_name": schema.StringAttribute{
				Description: "Name of the error table (when enabled).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"error_table_target_id": schema.StringAttribute{
				Description: "Target connector ID that receives error-table rows.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"target_table_name": schema.StringAttribute{
				Description: "Target table name. Defaults to \"root\".",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"backfill": schema.BoolAttribute{
				Description: "Whether the subscription backfills existing data.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"desired_state": schema.StringAttribute{
				Description: "Desired lifecycle state: \"running\" (enabled) or \"paused\" (disabled). " +
					"Defaults to the server state after creation.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf(desiredStateRunning, desiredStatePaused),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Description: "The current runtime status of the subscription.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"config_hash": schema.StringAttribute{
				Description: "Server-computed stable hash of the subscription configuration, for drift detection.",
				Computed:    true,
			},
		},
	}
}

func (r *subscriptionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *subscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan subscriptionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	smt, err := parseSmtConfig(plan.SmtConfig)
	if err != nil {
		resp.Diagnostics.AddError("Invalid smt_config", fmt.Sprintf("Could not parse smt_config as a JSON array: %s", err))
		return
	}

	createReq := &client.SubscriptionCreate{
		Name:              plan.Name.ValueString(),
		DatamodelID:       plan.DatamodelID.ValueString(),
		TargetConnectorID: plan.TargetConnectorID.ValueString(),
		SmtConfig:         smt,
		ConsumerID:        subStrPtr(plan.ConsumerID),
		ErrorTableEnabled: plan.ErrorTableEnabled.ValueBool(),
		ErrorTableName:    subStrOrEmpty(plan.ErrorTableName),
		TargetTableName:   subStrOrEmpty(plan.TargetTableName),
		Backfill:          plan.Backfill.ValueBool(),
	}

	id, err := r.client.CreateSubscription(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Subscription", fmt.Sprintf("Could not create subscription: %s", err))
		return
	}

	// error_table_target_id is not part of the create body; apply it via update.
	if !plan.ErrorTableTargetID.IsNull() && !plan.ErrorTableTargetID.IsUnknown() {
		etid := plan.ErrorTableTargetID.ValueString()
		if err := r.client.UpdateSubscription(ctx, id, &client.SubscriptionUpdate{ErrorTableTargetID: &etid}); err != nil {
			resp.Diagnostics.AddError("Error Creating Subscription", fmt.Sprintf("Could not set error_table_target_id: %s", err))
			return
		}
	}

	sub, err := r.reconcileDesiredState(ctx, id, plan.DesiredState)
	if err != nil {
		resp.Diagnostics.AddError("Error Setting Subscription State", fmt.Sprintf("Could not set desired_state on subscription %s: %s", id, err))
		return
	}
	if sub == nil {
		sub, err = r.client.GetSubscription(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError("Error Reading Subscription", fmt.Sprintf("Could not read subscription %s after create: %s", id, err))
			return
		}
	}

	r.mapToState(&plan, sub)
	tflog.Info(ctx, "Created subscription", map[string]any{"id": id})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state subscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sub, err := r.client.GetSubscription(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Subscription", fmt.Sprintf("Could not read subscription %s: %s", state.ID.ValueString(), err))
		return
	}
	if sub == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.mapToState(&state, sub)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *subscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan subscriptionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state subscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	updateReq := &client.SubscriptionUpdate{}
	changed := false

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		updateReq.Name = &v
		changed = true
	}
	if !plan.SmtConfig.Equal(state.SmtConfig) {
		smt, err := parseSmtConfig(plan.SmtConfig)
		if err != nil {
			resp.Diagnostics.AddError("Invalid smt_config", fmt.Sprintf("Could not parse smt_config as a JSON array: %s", err))
			return
		}
		updateReq.MapperConfig = smt
		changed = true
	}
	if !plan.ConsumerID.Equal(state.ConsumerID) {
		updateReq.ConsumerID = subStrPtr(plan.ConsumerID)
		changed = true
	}
	if !plan.ErrorTableEnabled.Equal(state.ErrorTableEnabled) {
		v := plan.ErrorTableEnabled.ValueBool()
		updateReq.ErrorTableEnabled = &v
		changed = true
	}
	if !plan.ErrorTableName.Equal(state.ErrorTableName) {
		v := plan.ErrorTableName.ValueString()
		updateReq.ErrorTableName = &v
		changed = true
	}
	if !plan.ErrorTableTargetID.Equal(state.ErrorTableTargetID) {
		updateReq.ErrorTableTargetID = subStrPtr(plan.ErrorTableTargetID)
		changed = true
	}
	if !plan.TargetTableName.Equal(state.TargetTableName) {
		v := plan.TargetTableName.ValueString()
		updateReq.TargetTableName = &v
		changed = true
	}
	if !plan.Backfill.Equal(state.Backfill) {
		v := plan.Backfill.ValueBool()
		updateReq.Backfill = &v
		changed = true
	}

	if changed {
		if err := r.client.UpdateSubscription(ctx, id, updateReq); err != nil {
			resp.Diagnostics.AddError("Error Updating Subscription", fmt.Sprintf("Could not update subscription %s: %s", id, err))
			return
		}
	}

	// Reconcile lifecycle if desired_state changed.
	var sub *client.SubscriptionRead
	if !plan.DesiredState.Equal(state.DesiredState) {
		var err error
		sub, err = r.reconcileDesiredState(ctx, id, plan.DesiredState)
		if err != nil {
			resp.Diagnostics.AddError("Error Setting Subscription State", fmt.Sprintf("Could not set desired_state on subscription %s: %s", id, err))
			return
		}
	}
	if sub == nil {
		var err error
		sub, err = r.client.GetSubscription(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError("Error Reading Subscription", fmt.Sprintf("Could not read subscription %s after update: %s", id, err))
			return
		}
	}

	r.mapToState(&plan, sub)
	tflog.Info(ctx, "Updated subscription", map[string]any{"id": id})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSubscription(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Subscription", fmt.Sprintf("Could not delete subscription %s: %s", state.ID.ValueString(), err))
		return
	}
	tflog.Info(ctx, "Deleted subscription", map[string]any{"id": state.ID.ValueString()})
}

func (r *subscriptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// reconcileDesiredState applies start/pause to match the desired state and
// returns the refreshed subscription. Returns (nil, nil) when no action was
// needed (desired_state unknown/null), so the caller performs its own read.
func (r *subscriptionResource) reconcileDesiredState(ctx context.Context, id string, desired types.String) (*client.SubscriptionRead, error) {
	if desired.IsNull() || desired.IsUnknown() {
		return nil, nil
	}
	switch desired.ValueString() {
	case desiredStatePaused:
		return r.client.PauseSubscription(ctx, id)
	case desiredStateRunning:
		return r.client.StartSubscription(ctx, id)
	default:
		return nil, nil
	}
}

// parseSmtConfig parses the opaque smt_config JSON array string into the client
// shape. An unset value yields nil.
func parseSmtConfig(v types.String) ([]map[string]any, error) {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return nil, nil
	}
	var out []map[string]any
	if err := json.Unmarshal([]byte(v.ValueString()), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// mapToState refreshes the model from a read response. smt_config is preserved
// from the model (the provider treats it opaquely; drift is tracked via
// config_hash). Required replace-triggering IDs are only overwritten when the
// response provides them (e.g. on import).
func (r *subscriptionResource) mapToState(model *subscriptionResourceModel, sub *client.SubscriptionRead) {
	model.ID = types.StringValue(sub.ID)
	model.Name = types.StringValue(sub.Name)
	model.Status = types.StringValue(sub.Status)
	model.ConfigHash = types.StringValue(sub.ConfigHash)

	if sub.DatamodelID != nil {
		model.DatamodelID = types.StringValue(*sub.DatamodelID)
	}
	if sub.TargetID != nil {
		model.TargetConnectorID = types.StringValue(*sub.TargetID)
	}

	model.TargetTableName = types.StringValue(sub.TargetTableName)
	model.ErrorTableEnabled = types.BoolValue(sub.ErrorTableEnabled)
	model.ErrorTableName = types.StringValue(sub.ErrorTableName)
	model.Backfill = types.BoolValue(sub.Backfill)

	model.ConsumerID = nullableString(sub.ConsumerID)
	model.ErrorTableTargetID = nullableString(sub.ErrorTableTargetID)

	if sub.Enabled {
		model.DesiredState = types.StringValue(desiredStateRunning)
	} else {
		model.DesiredState = types.StringValue(desiredStatePaused)
	}
}

func nullableString(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

func subStrOrEmpty(s types.String) string {
	if s.IsNull() || s.IsUnknown() {
		return ""
	}
	return s.ValueString()
}

func subStrPtr(s types.String) *string {
	if s.IsNull() || s.IsUnknown() {
		return nil
	}
	v := s.ValueString()
	return &v
}
