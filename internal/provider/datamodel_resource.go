package provider

import (
	"context"
	"fmt"

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
	datamodelStateRunning = "running"
	datamodelStateStopped = "stopped"
)

var (
	_ resource.Resource                = &datamodelResource{}
	_ resource.ResourceWithConfigure   = &datamodelResource{}
	_ resource.ResourceWithImportState = &datamodelResource{}
)

// NewDatamodelResource creates the popsink_datamodel adoption resource.
func NewDatamodelResource() resource.Resource {
	return &datamodelResource{}
}

type datamodelResource struct {
	client *client.Client
}

type datamodelResourceModel struct {
	ID           types.String `tfsdk:"id"`
	DatamodelID  types.String `tfsdk:"datamodel_id"`
	DesiredState types.String `tfsdk:"desired_state"`
	Name         types.String `tfsdk:"name"`
	State        types.String `tfsdk:"state"`
}

func (r *datamodelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datamodel"
}

func (r *datamodelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Adopts an existing Popsink datamodel and manages its lifecycle. Datamodels have no " +
			"create endpoint — they are derived from pipeline/connector configuration — so this resource " +
			"references one by ID (like `aws_default_vpc`): it does not create or delete the underlying " +
			"datamodel, only manages its desired lifecycle state. Destroying the resource removes it from " +
			"Terraform state and leaves the datamodel untouched.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The datamodel identifier (mirrors datamodel_id).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"datamodel_id": schema.StringAttribute{
				Description: "The ID of the existing datamodel to adopt. Changing this adopts a different datamodel.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			attrDesiredState: schema.StringAttribute{
				Description: "Desired lifecycle state: \"running\" (enabled) or \"stopped\" (disabled). " +
					"Managed via the datamodel start/stop endpoints; defaults to the server state.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf(datamodelStateRunning, datamodelStateStopped),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The datamodel name.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				Description: "The datamodel's current worker state.",
				Computed:    true,
			},
		},
	}
}

func (r *datamodelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *datamodelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan datamodelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.DatamodelID.ValueString()
	dm, err := r.client.GetDataModel(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Datamodel", fmt.Sprintf("Could not read datamodel %s: %s", id, err))
		return
	}
	if dm == nil {
		resp.Diagnostics.AddError(
			"Datamodel Not Found",
			fmt.Sprintf("No datamodel with ID %s exists. Datamodels are derived from pipeline/connector configuration and cannot be created by Terraform; create the source first, then adopt the datamodel.", id),
		)
		return
	}

	dm, err = r.reconcileDatamodelState(ctx, id, plan.DesiredState, dm)
	if err != nil {
		resp.Diagnostics.AddError("Error Setting Datamodel State", fmt.Sprintf("Could not set desired_state on datamodel %s: %s", id, err))
		return
	}

	r.mapDatamodelToState(&plan, dm)
	tflog.Info(ctx, "Adopted datamodel", map[string]any{"id": id})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *datamodelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state datamodelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dm, err := r.client.GetDataModel(ctx, state.DatamodelID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Datamodel", fmt.Sprintf("Could not read datamodel %s: %s", state.DatamodelID.ValueString(), err))
		return
	}
	if dm == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.mapDatamodelToState(&state, dm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *datamodelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan datamodelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state datamodelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.DatamodelID.ValueString()
	dm, err := r.client.GetDataModel(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Datamodel", fmt.Sprintf("Could not read datamodel %s: %s", id, err))
		return
	}
	if dm == nil {
		resp.Diagnostics.AddError("Datamodel Not Found", fmt.Sprintf("Datamodel %s no longer exists.", id))
		return
	}

	if !plan.DesiredState.Equal(state.DesiredState) {
		dm, err = r.reconcileDatamodelState(ctx, id, plan.DesiredState, dm)
		if err != nil {
			resp.Diagnostics.AddError("Error Setting Datamodel State", fmt.Sprintf("Could not set desired_state on datamodel %s: %s", id, err))
			return
		}
	}

	r.mapDatamodelToState(&plan, dm)
	tflog.Info(ctx, "Updated datamodel", map[string]any{"id": id})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes the resource from Terraform state only. The underlying
// datamodel is derived and is intentionally left untouched.
func (r *datamodelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state datamodelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "Releasing adopted datamodel from state (not deleting it)", map[string]any{"id": state.DatamodelID.ValueString()})
}

func (r *datamodelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("datamodel_id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// reconcileDatamodelState applies start/stop to match desired and returns the
// refreshed datamodel. When desired is unset or already satisfied, it returns
// the datamodel unchanged.
func (r *datamodelResource) reconcileDatamodelState(ctx context.Context, id string, desired types.String, current *client.DataModelRead) (*client.DataModelRead, error) {
	if desired.IsNull() || desired.IsUnknown() {
		return current, nil
	}
	switch desired.ValueString() {
	case datamodelStateRunning:
		if current.Enabled {
			return current, nil
		}
		return r.client.StartDataModel(ctx, id)
	case datamodelStateStopped:
		if !current.Enabled {
			return current, nil
		}
		return r.client.StopDataModel(ctx, id)
	default:
		return current, nil
	}
}

func (r *datamodelResource) mapDatamodelToState(model *datamodelResourceModel, dm *client.DataModelRead) {
	model.ID = types.StringValue(dm.ID)
	model.DatamodelID = types.StringValue(dm.ID)
	model.Name = types.StringValue(dm.Name)
	model.State = types.StringValue(dm.State)
	if dm.Enabled {
		model.DesiredState = types.StringValue(datamodelStateRunning)
	} else {
		model.DesiredState = types.StringValue(datamodelStateStopped)
	}
}
