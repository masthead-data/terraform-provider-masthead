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
var _ datasource.DataSource = &DataProductDataSource{}

func NewDataProductDataSource() datasource.DataSource {
	return &DataProductDataSource{}
}

// DataProductDataSource defines the data source implementation.
type DataProductDataSource struct {
	client *masthead.Client
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
	var config DataProductResourceModel
	var state DataProductResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the data product from Masthead API
	productResponse, err := d.client.GetDataProduct(config.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read data product, got error: %s", err))
		return
	}

	// Map response body to model
	state.Name = types.StringValue(productResponse.Name)
	if productResponse.Description == "" {
		state.Description = types.StringNull()
	} else {
		state.Description = types.StringValue(productResponse.Description)
	}
	if productResponse.DataDomain != nil {
		state.DataDomainUUID = types.StringValue(productResponse.DataDomain.UUID)
	} else {
		state.DataDomainUUID = types.StringNull()
	}

	// Map data assets
	if len(productResponse.DataAssets) > 0 {
		dataAssets := make([]DataProductAssetResourceModel, 0, len(productResponse.DataAssets))
		for _, asset := range productResponse.DataAssets {
			var mappedAsset DataProductAssetResourceModel
			mappedAsset.Type = asset.Type
			mappedAsset.UUID = types.StringValue(asset.UUID)
			mappedAsset.Project = types.StringValue(asset.Project)
			mappedAsset.Dataset = types.StringValue(asset.Dataset)
			mappedAsset.Table = types.StringValue(asset.Table)
			if asset.Table == "" {
				mappedAsset.Table = types.StringNull()
			} else {
				mappedAsset.Table = types.StringValue(asset.Table)
			}
			mappedAsset.AlertType = types.StringValue(string(asset.AlertType))

			// Add the mapped asset to the list
			dataAssets = append(dataAssets, mappedAsset)
		}
		state.DataAssets = dataAssets
	} else {
		state.DataAssets = []DataProductAssetResourceModel{}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
