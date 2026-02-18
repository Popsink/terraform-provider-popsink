package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/popsink/terraform-provider-popsink/internal/client"
)

// Ensure the implementation satisfies the provider.Provider interface
var _ provider.Provider = &popsinkProvider{}

// popsinkProvider defines the provider implementation
type popsinkProvider struct {
	version string
}

// popsinkProviderModel describes the provider data model
type popsinkProviderModel struct {
	BaseURL  types.String `tfsdk:"base_url"`
	Token    types.String `tfsdk:"token"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

// New creates a new provider instance
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &popsinkProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name
func (p *popsinkProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "popsink"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data
func (p *popsinkProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Popsink API to manage connectors and subscriptions.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Description: "The base URL for the Popsink API. May also be provided via POPSINK_BASE_URL environment variable.",
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: "The API token for authenticating with the Popsink API. May also be provided via POPSINK_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"insecure": schema.BoolAttribute{
				Description: "Skip TLS certificate verification. May also be provided via POPSINK_INSECURE environment variable.",
				Optional:    true,
			},
		},
	}
}

// Configure prepares the provider with the user-provided configuration
func (p *popsinkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Popsink client")

	var config popsinkProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default values from environment variables
	baseURL := os.Getenv("POPSINK_BASE_URL")
	token := os.Getenv("POPSINK_TOKEN")

	// Override with provider configuration if set
	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	// Validate required fields
	if baseURL == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Missing Base URL",
			"The provider cannot create the Popsink API client as there is a missing or empty value for the base URL. "+
				"Set the base_url value in the configuration or use the POPSINK_BASE_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing API Token",
			"The provider cannot create the Popsink API client as there is a missing or empty value for the API token. "+
				"Set the token value in the configuration or use the POPSINK_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Determine insecure mode
	insecure := os.Getenv("POPSINK_INSECURE") == "true"
	if !config.Insecure.IsNull() {
		insecure = config.Insecure.ValueBool()
	}

	// Create and configure the client
	c := client.NewClient(baseURL, token, insecure)

	// Make the client available to resources and data sources
	resp.DataSourceData = c
	resp.ResourceData = c

	tflog.Info(ctx, "Configured Popsink client", map[string]any{"base_url": baseURL})
}

// Resources returns the provider's resources
func (p *popsinkProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewTeamResource,
		NewConnectorResource,
		NewSubscriptionResource,
	}
}

// DataSources returns the provider's data sources
func (p *popsinkProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// No data sources yet
	}
}
