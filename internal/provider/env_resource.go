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

var (
	_ resource.Resource                = &envResource{}
	_ resource.ResourceWithConfigure   = &envResource{}
	_ resource.ResourceWithImportState = &envResource{}
)

// NewEnvResource creates a new environment resource.
func NewEnvResource() resource.Resource {
	return &envResource{}
}

type envResource struct {
	client *client.Client
}

type envResourceModel struct {
	ID                     types.String       `tfsdk:"id"`
	Name                   types.String       `tfsdk:"name"`
	RetentionConfiguration *brokerConfigModel `tfsdk:"retention_configuration"`
}

type brokerConfigModel struct {
	BootstrapServer  types.String `tfsdk:"bootstrap_server"`
	SecurityProtocol types.String `tfsdk:"security_protocol"`
	SaslMechanism    types.String `tfsdk:"sasl_mechanism"`
	SaslUsername     types.String `tfsdk:"sasl_username"`
	SaslPassword     types.String `tfsdk:"sasl_password"`
	CaCert           types.String `tfsdk:"ca_cert"`
	Cert             types.String `tfsdk:"cert"`
	Key              types.String `tfsdk:"key"`
	GroupID          types.String `tfsdk:"group_id"`
}

func (r *envResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_env"
}

func (r *envResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Popsink environment. An environment is the foundational namespace " +
			"of the data-plane: teams (and, through them, connectors and pipelines) are scoped " +
			"to an environment. Every environment carries a required Kafka broker retention " +
			"configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the environment.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the environment.",
				Required:    true,
			},
			"retention_configuration": schema.SingleNestedAttribute{
				Description: "Kafka broker retention configuration (required). Credential fields " +
					"are write-only from the API's perspective: they are accepted on create/update " +
					"but never returned on read, so the provider keeps the configured values in state.",
				Required: true,
				Attributes: map[string]schema.Attribute{
					"bootstrap_server": schema.StringAttribute{
						Description: "Kafka bootstrap server (host:port).",
						Required:    true,
					},
					"security_protocol": schema.StringAttribute{
						Description: "Security protocol. Defaults to PLAINTEXT server-side when omitted.",
						Optional:    true,
						Computed:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("PLAINTEXT", "SASL_PLAINTEXT", "SASL_SSL", "SSL"),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"sasl_mechanism": schema.StringAttribute{
						Description: "SASL mechanism. Defaults to PLAIN server-side when omitted.",
						Optional:    true,
						Computed:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("OAUTHBEARER", "PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512", "GSSAPI"),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"sasl_username": schema.StringAttribute{
						Description: "SASL username. Not returned on read.",
						Optional:    true,
					},
					"sasl_password": schema.StringAttribute{
						Description: "SASL password. Sensitive; not returned on read.",
						Optional:    true,
						Sensitive:   true,
					},
					"ca_cert": schema.StringAttribute{
						Description: "CA certificate (PEM). Sensitive; not returned on read.",
						Optional:    true,
						Sensitive:   true,
					},
					"cert": schema.StringAttribute{
						Description: "Client certificate (PEM). Sensitive; not returned on read.",
						Optional:    true,
						Sensitive:   true,
					},
					"key": schema.StringAttribute{
						Description: "Client key (PEM). Sensitive; not returned on read.",
						Optional:    true,
						Sensitive:   true,
					},
					"group_id": schema.StringAttribute{
						Description: "Consumer group ID.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
		},
	}
}

func (r *envResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *envResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan envResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &client.EnvCreate{
		Name:                   plan.Name.ValueString(),
		RetentionConfiguration: brokerConfigToClient(plan.RetentionConfiguration),
	}

	env, err := r.client.CreateEnv(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Environment", fmt.Sprintf("Could not create environment: %s", err))
		return
	}

	// Refresh public fields from the response, preserving the plan's credential
	// values (the API does not return them).
	plan.ID = types.StringValue(env.ID)
	plan.Name = types.StringValue(env.Name)
	applyBrokerPublicFields(plan.RetentionConfiguration, &env.RetentionConfiguration)

	tflog.Info(ctx, "Created environment", map[string]any{"id": env.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *envResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state envResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, err := r.client.GetEnv(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Environment", fmt.Sprintf("Could not read environment %s: %s", state.ID.ValueString(), err))
		return
	}

	if env == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(env.Name)
	// Preserve credential fields already in state; the API omits them on read.
	applyBrokerPublicFields(state.RetentionConfiguration, &env.RetentionConfiguration)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *envResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan envResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state envResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &client.EnvUpdate{}
	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}

	// The API's broker read shape omits credentials, so the plan and prior state
	// are never byte-equal on the credential fields; send the whole broker block
	// on any change to a broker attribute.
	if !brokerConfigsEqual(plan.RetentionConfiguration, state.RetentionConfiguration) {
		cfg := brokerConfigToClient(plan.RetentionConfiguration)
		updateReq.RetentionConfiguration = &cfg
	}

	env, err := r.client.UpdateEnv(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Environment", fmt.Sprintf("Could not update environment %s: %s", state.ID.ValueString(), err))
		return
	}

	plan.ID = types.StringValue(env.ID)
	plan.Name = types.StringValue(env.Name)
	applyBrokerPublicFields(plan.RetentionConfiguration, &env.RetentionConfiguration)

	tflog.Info(ctx, "Updated environment", map[string]any{"id": env.ID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *envResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state envResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteEnv(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Environment", fmt.Sprintf("Could not delete environment %s: %s", state.ID.ValueString(), err))
		return
	}

	tflog.Info(ctx, "Deleted environment", map[string]any{"id": state.ID.ValueString()})
}

func (r *envResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// brokerConfigToClient converts the nested TF model into the client broker
// configuration, sending optional string fields only when set.
func brokerConfigToClient(m *brokerConfigModel) client.BrokerConfiguration {
	cfg := client.BrokerConfiguration{
		BootstrapServer:  m.BootstrapServer.ValueString(),
		SecurityProtocol: stringOrEmpty(m.SecurityProtocol),
		SaslMechanism:    stringOrEmpty(m.SaslMechanism),
		SaslUsername:     stringPtrOrNil(m.SaslUsername),
		SaslPassword:     stringPtrOrNil(m.SaslPassword),
		CaCert:           stringOrEmpty(m.CaCert),
		Cert:             stringOrEmpty(m.Cert),
		Key:              stringOrEmpty(m.Key),
		GroupID:          stringPtrOrNil(m.GroupID),
	}
	return cfg
}

// applyBrokerPublicFields refreshes only the fields the API returns on read
// (bootstrap_server, security_protocol, sasl_mechanism, group_id), leaving the
// write-only credential fields in the model untouched.
func applyBrokerPublicFields(m *brokerConfigModel, read *client.BrokerConfiguration) {
	if m == nil {
		return
	}
	m.BootstrapServer = types.StringValue(read.BootstrapServer)
	m.SecurityProtocol = types.StringValue(read.SecurityProtocol)
	m.SaslMechanism = types.StringValue(read.SaslMechanism)
	if read.GroupID != nil {
		m.GroupID = types.StringValue(*read.GroupID)
	} else {
		m.GroupID = types.StringNull()
	}
}

func brokerConfigsEqual(a, b *brokerConfigModel) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.BootstrapServer.Equal(b.BootstrapServer) &&
		a.SecurityProtocol.Equal(b.SecurityProtocol) &&
		a.SaslMechanism.Equal(b.SaslMechanism) &&
		a.SaslUsername.Equal(b.SaslUsername) &&
		a.SaslPassword.Equal(b.SaslPassword) &&
		a.CaCert.Equal(b.CaCert) &&
		a.Cert.Equal(b.Cert) &&
		a.Key.Equal(b.Key) &&
		a.GroupID.Equal(b.GroupID)
}

func stringOrEmpty(s types.String) string {
	if s.IsNull() || s.IsUnknown() {
		return ""
	}
	return s.ValueString()
}

func stringPtrOrNil(s types.String) *string {
	if s.IsNull() || s.IsUnknown() {
		return nil
	}
	v := s.ValueString()
	return &v
}
