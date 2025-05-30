package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	masthead "github.com/masthead-data/terraform-provider-masthead/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &DataDomainDataSource{}

func NewDataDomainDataSource() datasource.DataSource {
	return &DataDomainDataSource{}
}

// DataDomainDataSource defines the data source implementation.
type DataDomainDataSource struct {
	client *masthead.Client
}

func (d *DataDomainDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_domain"
}

func (d *DataDomainDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch information about a Masthead data domain",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				MarkdownDescription: "UUID of the data domain",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the data domain",
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email associated with the data domain",
				Computed:            true,
			},
			"slack_channel_name": schema.StringAttribute{
				MarkdownDescription: "Slack channel name associated with the data domain",
				Computed:            true,
			},
		},
	}
}

func (d *DataDomainDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*masthead.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *masthead.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *DataDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DataDomainResourceModel
	var state DataDomainResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the data domain from Masthead API
	domainResponse, err := d.client.GetDomain(config.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read data domain, got error: %s", err))
		return
	}

	// Map response body to model
	state.Name = types.StringValue(domainResponse.Name)
	state.Email = types.StringValue(domainResponse.Email)
	if domainResponse.SlackChannel != (masthead.SlackChannel{}) {
		state.SlackChannelName = types.StringValue(domainResponse.SlackChannel.Name)
	} else {
		state.SlackChannelName = types.StringNull()
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
