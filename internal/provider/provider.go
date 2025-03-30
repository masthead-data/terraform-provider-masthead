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

	"github.com/masthead-data/terraform-provider-masthead/internal/client"
)

// Ensure mastheadProvider satisfies various provider interfaces.
var _ provider.Provider = &mastheadProvider{}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &mastheadProvider{
			version: version,
		}
	}
}

// mastheadProvider defines the provider implementation.
type mastheadProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
	ApiKey  types.String `tfsdk:"api_token"`
}

// mastheadProviderModel maps provider schema data to a Go type.
type mastheadProviderModel struct {
	Token types.String `tfsdk:"api_token"`
}

func (p *mastheadProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "masthead"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *mastheadProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				MarkdownDescription: "Masthead API Token. This token is used to authenticate with the Masthead API. " +
					"To obtain a token, log in to your Masthead account and navigate to the **Settings / API Tokens** page. " +
					"Create a new token and copy it here. " +
					"Alternatively, you can set the `MASTHEAD_API_TOKEN` environment variable to use the token from there.",
				Required:  false,
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *mastheadProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config mastheadProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Unknown Masthead API Token",
			"The provider cannot create the Masthead API client as there is an unknown configuration value for the Masthead API token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MASTHEAD_API_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	api_token := os.Getenv("MASTHEAD_API_TOKEN")

	if !config.Token.IsNull() {
		api_token = config.Token.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if api_token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Missing Masthead API Token",
			"The provider cannot create the Masthead API client as there is a missing or empty value for the Masthead API token. "+
				"Set the token value in the configuration or use the MASTHEAD_API_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new Masthead client using the configuration values
	client, err := masthead.NewClient(&api_token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Masthead API Client",
			"An unexpected error occurred when creating the Masthead API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Masthead Client Error: "+err.Error(),
		)
		return
	}

	// Make the Masthead client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *mastheadProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUserResource,
		NewDataDomainResource,
		NewDataProductResource,
	}
}

func (p *mastheadProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUserDataSource,
		NewDataDomainDataSource,
		NewDataProductDataSource,
	}
}
