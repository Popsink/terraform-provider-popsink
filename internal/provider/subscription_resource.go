package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/popsink/terraform-provider-popsink/internal/client"
)

var (
	_ resource.Resource                = &subscriptionResource{}
	_ resource.ResourceWithConfigure   = &subscriptionResource{}
	_ resource.ResourceWithImportState = &subscriptionResource{}
)

func NewSubscriptionResource() resource.Resource {
	return &subscriptionResource{}
}

type subscriptionResource struct {
	client *client.Client
}

type subscriptionResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	DatamodelID       types.String `tfsdk:"datamodel_id"`
	TargetConnectorID types.String `tfsdk:"target_connector_id"`
	TargetTableName   types.String `tfsdk:"target_table_name"`
	Backfill          types.Bool   `tfsdk:"backfill"`
	ErrorTableEnabled types.Bool   `tfsdk:"error_table_enabled"`
	ErrorTableName    types.String `tfsdk:"error_table_name"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	Status            types.String `tfsdk:"status"`
}

func (r *subscriptionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription"
}

func (r *subscriptionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Popsink subscription (data model to target connector).",
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
				Description: "The ID of the source data model.",
				Required:    true,
			},
			"target_connector_id": schema.StringAttribute{
				Description: "The ID of the target connector.",
				Required:    true,
			},
			"target_table_name": schema.StringAttribute{
				Description: "The target table name.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("root"),
			},
			"backfill": schema.BoolAttribute{
				Description: "Whether to backfill historical data.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"error_table_enabled": schema.BoolAttribute{
				Description: "Whether to enable error table for failed records.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"error_table_name": schema.StringAttribute{
				Description: "The name of the error table.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the subscription is currently enabled.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The current status of the subscription.",
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

	createReq := &client.SubscriptionCreate{
		Name:              plan.Name.ValueString(),
		DatamodelID:       plan.DatamodelID.ValueString(),
		TargetConnectorID: plan.TargetConnectorID.ValueString(),
		TargetTableName:   plan.TargetTableName.ValueString(),
		Backfill:          plan.Backfill.ValueBool(),
		ErrorTableEnabled: plan.ErrorTableEnabled.ValueBool(),
		ErrorTableName:    plan.ErrorTableName.ValueString(),
	}

	sub, err := r.client.CreateSubscription(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Subscription", fmt.Sprintf("Could not create subscription: %s", err))
		return
	}

	r.mapToState(&plan, sub)
	tflog.Info(ctx, "Created subscription", map[string]any{"id": sub.ID})
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

	updateReq := &client.SubscriptionUpdate{}

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}

	if !plan.TargetTableName.Equal(state.TargetTableName) {
		ttn := plan.TargetTableName.ValueString()
		updateReq.TargetTableName = &ttn
	}

	if !plan.Backfill.Equal(state.Backfill) {
		bf := plan.Backfill.ValueBool()
		updateReq.Backfill = &bf
	}

	if !plan.ErrorTableEnabled.Equal(state.ErrorTableEnabled) {
		ete := plan.ErrorTableEnabled.ValueBool()
		updateReq.ErrorTableEnabled = &ete
	}

	if !plan.ErrorTableName.Equal(state.ErrorTableName) {
		etn := plan.ErrorTableName.ValueString()
		updateReq.ErrorTableName = &etn
	}

	sub, err := r.client.UpdateSubscription(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Subscription", fmt.Sprintf("Could not update subscription %s: %s", state.ID.ValueString(), err))
		return
	}

	r.mapToState(&plan, sub)
	tflog.Info(ctx, "Updated subscription", map[string]any{"id": sub.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subscriptionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSubscription(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Subscription", fmt.Sprintf("Could not delete subscription %s: %s", state.ID.ValueString(), err))
		return
	}

	tflog.Info(ctx, "Deleted subscription", map[string]any{"id": state.ID.ValueString()})
}

func (r *subscriptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *subscriptionResource) mapToState(model *subscriptionResourceModel, sub *client.SubscriptionRead) {
	model.ID = types.StringValue(sub.ID)
	model.Name = types.StringValue(sub.Name)
	model.DatamodelID = types.StringValue(sub.DatamodelID)
	model.TargetConnectorID = types.StringValue(sub.TargetConnectorID)
	model.TargetTableName = types.StringValue(sub.TargetTableName)
	model.Backfill = types.BoolValue(sub.Backfill)
	model.ErrorTableEnabled = types.BoolValue(sub.ErrorTableEnabled)
	model.ErrorTableName = types.StringValue(sub.ErrorTableName)
	model.Enabled = types.BoolValue(sub.Enabled)
	model.Status = types.StringValue(sub.Status)
}
