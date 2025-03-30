package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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

// DataDomainDataSourceModel describes the data source data model.
type DataDomainDataSourceModel struct {
	UUID             types.String `tfsdk:"uuid"`
	Name             types.String `tfsdk:"name"`
	Email            types.String `tfsdk:"email"`
	SlackChannelName types.String `tfsdk:"slack_channel_name"`
	SlackChannelID   types.String `tfsdk:"slack_channel_id"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
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
				Required:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email associated with the data domain",
				Required:            true,
			},
			"slack_channel_name": schema.StringAttribute{
				MarkdownDescription: "Name of the Slack channel associated with the data domain",
				Computed:            true,
			},
			"slack_channel_id": schema.StringAttribute{
				MarkdownDescription: "ID of the Slack channel associated with the data domain",
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
	var data DataDomainDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the data domain from Masthead API
	domain, err := d.client.GetDomain(data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read data domain, got error: %s", err))
		return
	}

	// Map response body to model
	data.Name = types.StringValue(domain.Name)
	data.Email = types.StringValue(domain.Email)
	if domain.SlackChannel.Name != "" {
		data.SlackChannelName = types.StringValue(domain.SlackChannel.Name)
	}
	if domain.SlackChannel.ID != "" {
		data.SlackChannelID = types.StringValue(domain.SlackChannel.ID)
	}

	tflog.Debug(ctx, "Read data domain data source", map[string]interface{}{
		"uuid": data.UUID.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
