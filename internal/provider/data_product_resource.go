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
	Type      masthead.DataProductAssetType `tfsdk:"type"`
	UUID      types.String                  `tfsdk:"uuid"`
	Project   types.String                  `tfsdk:"project"`
	Dataset   types.String                  `tfsdk:"dataset"`
	Table     types.String                  `tfsdk:"table"`
	AlertType types.String                  `tfsdk:"alert_type"`
}

// DataProductResourceModel describes the resource data model.
type DataProductResourceModel struct {
	UUID           types.String                    `tfsdk:"uuid"`
	Name           types.String                    `tfsdk:"name"`
	Description    types.String                    `tfsdk:"description"`
	DataDomainUUID types.String                    `tfsdk:"data_domain_uuid"`
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
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the data product",
				Optional:            true,
			},
			"data_assets": schema.ListNestedAttribute{
				MarkdownDescription: "List of data assets associated with this data product",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of the data asset (DATASET, TABLE)",
							Required:            true,
						},
						"uuid": schema.StringAttribute{
							MarkdownDescription: "UUID of the data asset",
							Computed:            true,
						},
						"project": schema.StringAttribute{
							MarkdownDescription: "Project associated with the data asset",
							Required:            true,
						},
						"dataset": schema.StringAttribute{
							MarkdownDescription: "Dataset associated with the data asset",
							Required:            true,
						},
						"table": schema.StringAttribute{
							MarkdownDescription: "Table associated with the data asset",
							Optional:            true,
						},
						"alert_type": schema.StringAttribute{
							MarkdownDescription: "Alert type associated with the data asset",
							Computed:            true,
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
	var plan DataProductResourceModel
	var state DataProductResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new data product
	productRequest := masthead.DataProduct{
		Name:           plan.Name.ValueString(),
		Description:    plan.Description.ValueString(),
		DataDomainUUID: plan.DataDomainUUID.ValueString(),
	}

	// Add data assets if specified
	if len(plan.DataAssets) > 0 {
		productRequest.DataAssets = make([]masthead.DataProductAsset, 0, len(plan.DataAssets))
		for _, asset := range plan.DataAssets {
			productRequest.DataAssets = append(productRequest.DataAssets, masthead.DataProductAsset{
				Type:    asset.Type,
				UUID:    asset.UUID.ValueString(),
				Project: asset.Project.ValueString(),
				Dataset: asset.Dataset.ValueString(),
				Table:   asset.Table.ValueString(),
			})
		}
	}

	productResponse, err := r.client.CreateDataProduct(productRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create data product, got error: %s", err))
		return
	}

	// Map response to model
	state.UUID = types.StringValue(productResponse.UUID)
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

func (r *DataProductResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var plan DataProductResourceModel
	var state DataProductResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get data product by UUID
	productResponse, err := r.client.GetDataProduct(plan.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read data product, got error: %s", err))
		return
	}

	// Map response to model
	state.UUID = types.StringValue(productResponse.UUID)
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
			dataAssets = append(dataAssets, DataProductAssetResourceModel{
				Type:      asset.Type,
				UUID:      types.StringValue(asset.UUID),
				Project:   types.StringValue(asset.Project),
				Dataset:   types.StringValue(asset.Dataset),
				Table:     types.StringValue(asset.Table),
				AlertType: types.StringValue(string(asset.AlertType)),
			})
		}
		state.DataAssets = dataAssets
	} else {
		state.DataAssets = []DataProductAssetResourceModel{}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DataProductResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DataProductResourceModel
	var state DataProductResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing data product
	productRequest := masthead.DataProduct{
		UUID:           plan.UUID.ValueString(),
		Name:           plan.Name.ValueString(),
		DataDomainUUID: plan.DataDomainUUID.ValueString(),
		Description:    plan.Description.ValueString(),
	}

	// Add data assets if specified
	if len(plan.DataAssets) > 0 {
		productRequest.DataAssets = make([]masthead.DataProductAsset, 0, len(plan.DataAssets))
		for _, asset := range plan.DataAssets {
			productRequest.DataAssets = append(productRequest.DataAssets, masthead.DataProductAsset{
				Type:      asset.Type,
				UUID:      asset.UUID.ValueString(),
				Project:   asset.Project.ValueString(),
				Dataset:   asset.Dataset.ValueString(),
				Table:     asset.Table.ValueString(),
				AlertType: masthead.AlertType(asset.AlertType.ValueString()),
			})
		}
	}

	productResponse, err := r.client.UpdateDataProduct(productRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update data product, got error: %s", err))
		return
	}

	// Map response to model
	state.UUID = types.StringValue(productResponse.UUID)
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

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DataProductResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DataProductResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete data product
	err := r.client.DeleteDataProduct(state.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete data product, got error: %s", err))
		return
	}
}

func (r *DataProductResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
