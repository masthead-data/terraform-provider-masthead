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
var _ datasource.DataSource = &DataProductDataSource{}

func NewDataProductDataSource() datasource.DataSource {
	return &DataProductDataSource{}
}

// DataProductDataSource defines the data source implementation.
type DataProductDataSource struct {
	client *masthead.Client
}

// DataProductAssetModel describes a data asset in the data source model
type DataProductAssetModel struct {
	Type      masthead.DataProductAssetType `tfsdk:"type"`
	UUID      types.String                  `tfsdk:"uuid"`
	Project   types.String                  `tfsdk:"project"`
	Dataset   types.String                  `tfsdk:"dataset"`
	Table     types.String                  `tfsdk:"table"`
	AlertType masthead.AlertType            `tfsdk:"alert_type"`
}

// DataProductDataSourceModel describes the data source data model.
type DataProductDataSourceModel struct {
	UUID           types.String            `tfsdk:"uuid"`
	Name           types.String            `tfsdk:"name"`
	Description    types.String            `tfsdk:"description"`
	DataDomainUUID types.String            `tfsdk:"data_domain_uuid"`
	Domain         DataDomainResourceModel `tfsdk:"domain"`
	DataAssets     []DataProductAssetModel `tfsdk:"data_assets"`
}

func (d *DataProductDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_product"
}

func (d *DataProductDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch information about a Masthead data product",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				MarkdownDescription: "UUID of the data product",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the data product",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the data product",
				Computed:            true,
			},
			"data_domain_uuid": schema.StringAttribute{
				MarkdownDescription: "UUID of the data domain this product belongs to",
				Computed:            true,
			},
			"domain": schema.SingleNestedAttribute{
				MarkdownDescription: "Data domain associated with this data product",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"uuid": schema.StringAttribute{
						MarkdownDescription: "UUID of the data domain",
						Computed:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "Name of the data domain",
						Computed:            true,
					},
					"email": schema.StringAttribute{
						MarkdownDescription: "Email associated with the data domain",
						Computed:            true,
					},
					"slack_channel": schema.SingleNestedAttribute{
						MarkdownDescription: "Slack channel associated with the data domain",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								MarkdownDescription: "Name of the Slack channel",
								Computed:            true,
							},
							"id": schema.StringAttribute{
								MarkdownDescription: "ID of the Slack channel",
								Computed:            true,
							},
						},
					},
				},
			},
			"data_assets": schema.ListNestedAttribute{
				MarkdownDescription: "List of data assets associated with this data product",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of the data asset (DATASET, TABLE)",
							Computed:            true,
						},
						"uuid": schema.StringAttribute{
							MarkdownDescription: "UUID of the data asset",
							Computed:            true,
						},
						"project": schema.StringAttribute{
							MarkdownDescription: "Project of the data asset",
							Computed:            true,
						},
						"dataset": schema.StringAttribute{
							MarkdownDescription: "Dataset of the data asset",
							Computed:            true,
						},
						"table": schema.StringAttribute{
							MarkdownDescription: "Table of the data asset",
							Computed:            true,
						},
						"alert_type": schema.StringAttribute{
							MarkdownDescription: "Alert type of the data asset (e.g., DATASET, TABLE)",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *DataProductDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataProductDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DataProductDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the data product from Masthead API
	product, err := d.client.GetDataProduct(data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read data product, got error: %s", err))
		return
	}

	// Map response body to model
	data.Name = types.StringValue(product.Name)
	data.Description = types.StringValue(product.Description)
	data.DataDomainUUID = types.StringValue(product.DataDomainUUID)

	// Map data assets
	dataAssets := make([]DataProductAssetModel, 0, len(product.DataAssets))
	for _, asset := range product.DataAssets {
		dataAssets = append(dataAssets, DataProductAssetModel{
			Type: asset.Type,
			UUID: types.StringValue(asset.UUID),
			Project: types.StringValue(asset.Project),
			Dataset: types.StringValue(asset.Dataset),
			Table: types.StringValue(asset.Table),
			AlertType: asset.AlertType,
		})
	}
	data.DataAssets = dataAssets


	data.Domain = DataDomainResourceModel{
		UUID: types.StringValue(product.Domain.UUID),
		Name: types.StringValue(product.Domain.Name),
		Email: types.StringValue(product.Domain.Email),
		SlackChannel: SlackChannelModel{
			Name: types.StringValue(product.Domain.SlackChannel.Name),
			ID: types.StringValue(product.Domain.SlackChannel.ID),
		},
	}

	tflog.Debug(ctx, "Read data product data source", map[string]interface{}{
		"uuid": data.UUID.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
