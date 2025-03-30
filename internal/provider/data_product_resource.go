package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	masthead "github.com/masthead-data/terraform-provider-masthead/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &DataProductResource{}
var _ resource.ResourceWithImportState = &DataProductResource{}

func NewDataProductResource() resource.Resource {
	return &DataProductResource{}
}

// DataProductResource defines the resource implementation.
type DataProductResource struct {
	client *masthead.Client
}

// DataProductAssetResourceModel describes a data asset in the resource model
type DataProductAssetResourceModel struct {
	Type types.String `tfsdk:"type"`
	UUID types.String `tfsdk:"uuid"`
}

// DataProductResourceModel describes the resource data model.
type DataProductResourceModel struct {
	UUID           types.String                    `tfsdk:"uuid"`
	Name           types.String                    `tfsdk:"name"`
	DataDomainUUID types.String                    `tfsdk:"data_domain_uuid"`
	Description    types.String                    `tfsdk:"description"`
	DataAssets     []DataProductAssetResourceModel `tfsdk:"data_assets"`
}

func (r *DataProductResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_product"
}

func (r *DataProductResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Masthead data product",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				MarkdownDescription: "UUID of the data product",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the data product",
				Required:            true,
			},
			"data_domain_uuid": schema.StringAttribute{
				MarkdownDescription: "UUID of the data domain this product belongs to",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the data product",
				Optional:            true,
			},
			"data_assets": schema.ListNestedAttribute{
				MarkdownDescription: "List of data assets associated with this data product",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of the data asset (DATASET, TABLE)",
							Required:            true,
						},
						"uuid": schema.StringAttribute{
							MarkdownDescription: "UUID of the data asset",
							Required:            true,
						},
					},
				},
			},
		},
	}
}

func (r *DataProductResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*masthead.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *masthead.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *DataProductResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DataProductResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new data product
	product := masthead.DataProduct{
		Name:           data.Name.ValueString(),
		DataDomainUUID: data.DataDomainUUID.ValueString(),
		Description:    data.Description.ValueString(),
	}

	// Add data assets if specified
	if len(data.DataAssets) > 0 {
		product.DataAssets = make([]masthead.DataProductAsset, 0, len(data.DataAssets))
		for _, asset := range data.DataAssets {
			product.DataAssets = append(product.DataAssets, masthead.DataProductAsset{
				Type: masthead.DataProductAssetType(asset.Type.ValueString()),
				UUID: asset.UUID.ValueString(),
			})
		}
	}

	createdProduct, err := r.client.CreateDataProduct(product)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create data product, got error: %s", err))
		return
	}

	// Map response to model
	data.UUID = types.StringValue(createdProduct.UUID)
	data.Name = types.StringValue(createdProduct.Name)
	data.DataDomainUUID = types.StringValue(createdProduct.DataDomainUUID)
	data.Description = types.StringValue(createdProduct.Description)

	// Map data assets
	if len(createdProduct.DataAssets) > 0 {
		dataAssets := make([]DataProductAssetResourceModel, 0, len(createdProduct.DataAssets))
		for _, asset := range createdProduct.DataAssets {
			dataAssets = append(dataAssets, DataProductAssetResourceModel{
				Type: types.StringValue(string(asset.Type)),
				UUID: types.StringValue(asset.UUID),
			})
		}
		data.DataAssets = dataAssets
	}

	tflog.Trace(ctx, "created a data product resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DataProductResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DataProductResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get data product by UUID
	product, err := r.client.GetDataProduct(data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read data product, got error: %s", err))
		return
	}

	// Map response to model
	data.Name = types.StringValue(product.Name)
	data.DataDomainUUID = types.StringValue(product.DataDomainUUID)
	data.Description = types.StringValue(product.Description)

	// Map data assets
	if len(product.DataAssets) > 0 {
		dataAssets := make([]DataProductAssetResourceModel, 0, len(product.DataAssets))
		for _, asset := range product.DataAssets {
			dataAssets = append(dataAssets, DataProductAssetResourceModel{
				Type: types.StringValue(string(asset.Type)),
				UUID: types.StringValue(asset.UUID),
			})
		}
		data.DataAssets = dataAssets
	} else {
		data.DataAssets = []DataProductAssetResourceModel{}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DataProductResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DataProductResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing data product
	product := masthead.DataProduct{
		UUID:           data.UUID.ValueString(),
		Name:           data.Name.ValueString(),
		DataDomainUUID: data.DataDomainUUID.ValueString(),
		Description:    data.Description.ValueString(),
	}

	// Add data assets if specified
	if len(data.DataAssets) > 0 {
		product.DataAssets = make([]masthead.DataProductAsset, 0, len(data.DataAssets))
		for _, asset := range data.DataAssets {
			product.DataAssets = append(product.DataAssets, masthead.DataProductAsset{
				Type: masthead.DataProductAssetType(asset.Type.ValueString()),
				UUID: asset.UUID.ValueString(),
			})
		}
	}

	updatedProduct, err := r.client.UpdateDataProduct(product)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update data product, got error: %s", err))
		return
	}

	// Map response to model
	data.Name = types.StringValue(updatedProduct.Name)
	data.DataDomainUUID = types.StringValue(updatedProduct.DataDomainUUID)
	data.Description = types.StringValue(updatedProduct.Description)

	// Map data assets
	if len(updatedProduct.DataAssets) > 0 {
		dataAssets := make([]DataProductAssetResourceModel, 0, len(updatedProduct.DataAssets))
		for _, asset := range updatedProduct.DataAssets {
			dataAssets = append(dataAssets, DataProductAssetResourceModel{
				Type: types.StringValue(string(asset.Type)),
				UUID: types.StringValue(asset.UUID),
			})
		}
		data.DataAssets = dataAssets
	} else {
		data.DataAssets = []DataProductAssetResourceModel{}
	}

	tflog.Trace(ctx, "updated a data product resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DataProductResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DataProductResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete data product
	err := r.client.DeleteDataProduct(data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete data product, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted a data product resource")
}

func (r *DataProductResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
