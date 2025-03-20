package provider

import (
	"context"
	"os"
	"github.com/masthead-data/terraform-provider-masthead/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure mastheadProvider satisfies various provider interfaces.
var _ provider.Provider = &mastheadProvider{}

// mastheadProvider defines the provider implementation.
type mastheadProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *mastheadProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "masthead"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *mastheadProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "host": schema.StringAttribute{
                Optional: true,
            },
            "token": schema.StringAttribute{
                Optional: true,
            },
        },
    }
}

func (p *mastheadProvider) Resources(ctx context.Context) []func() resource.Resource {
	return nil
}


func (p *mastheadProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &mastheadProvider{
			version: version,
		}
	}
}

// mastheadProviderModel maps provider schema data to a Go type.
type mastheadProviderModel struct {
    Host     types.String `tfsdk:"host"`
    Token types.String `tfsdk:"token"`
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

    if config.Host.IsUnknown() {
        resp.Diagnostics.AddAttributeError(
            path.Root("host"),
            "Unknown Masthead API Host",
            "The provider cannot create the Masthead API client as there is an unknown configuration value for the Masthead API host. "+
                "Either target apply the source of the value first, set the value statically in the configuration, or use the MASTHEAD_HOST environment variable.",
        )
    }

    if config.Token.IsUnknown() {
        resp.Diagnostics.AddAttributeError(
            path.Root("token"),
            "Unknown Masthead API Token",
            "The provider cannot create the Masthead API client as there is an unknown configuration value for the Masthead API token. "+
                "Either target apply the source of the value first, set the value statically in the configuration, or use the MASTHEAD_TOKEN environment variable.",
        )
    }

    if resp.Diagnostics.HasError() {
        return
    }

    // Default values to environment variables, but override
    // with Terraform configuration value if set.

    host := os.Getenv("MASTHEAD_HOST")
    token := os.Getenv("MASTHEAD_TOKEN")


    if !config.Host.IsNull() {
        host = config.Host.ValueString()
    }

    if !config.Token.IsNull() {
        token = config.Token.ValueString()
    }

    // If any of the expected configurations are missing, return
    // errors with provider-specific guidance.

    if host == "" {
        resp.Diagnostics.AddAttributeError(
            path.Root("host"),
            "Missing Masthead API Host",
            "The provider cannot create the Masthead API client as there is a missing or empty value for the Masthead API host. "+
                "Set the host value in the configuration or use the Masthead_HOST environment variable. "+
                "If either is already set, ensure the value is not empty.",
        )
    }

    if token == "" {
        resp.Diagnostics.AddAttributeError(
            path.Root("username"),
            "Missing Masthead API Username",
            "The provider cannot create the Masthead API client as there is a missing or empty value for the Masthead API username. "+
                "Set the username value in the configuration or use the Masthead_USERNAME environment variable. "+
                "If either is already set, ensure the value is not empty.",
        )
    }

    if resp.Diagnostics.HasError() {
        return
    }

    // Create a new Masthead client using the configuration values
    client, err := masthead.NewClient(&host, &token)
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
