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
var _ resource.Resource = &DataDomainResource{}
var _ resource.ResourceWithImportState = &DataDomainResource{}

func NewDataDomainResource() resource.Resource {
	return &DataDomainResource{}
}

// DataDomainResource defines the resource implementation.
type DataDomainResource struct {
	client *masthead.Client
}

type SlackChannelModel struct {
	Name types.String `tfsdk:"name"`
	ID   types.String `tfsdk:"id"`
}

type DataDomainResourceModel struct {
	UUID         types.String      `tfsdk:"uuid"`
	Name         types.String      `tfsdk:"name"`
	Email        types.String      `tfsdk:"email"`
	SlackChannel SlackChannelModel `tfsdk:"slack_channel"`
}

func (r *DataDomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_domain"
}

func (r *DataDomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Masthead data domain",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				MarkdownDescription: "UUID of the data domain",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				Optional:            true,
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
	}
}

func (r *DataDomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DataDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DataDomainResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new domain
	domain := masthead.DataDomain{
		Name:             data.Name.ValueString(),
		Email:            data.Email.ValueString(),
		SlackChannelName: data.SlackChannel.Name.ValueString(),
	}

	createdDomain, err := r.client.CreateDomain(domain)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create data domain, got error: %s", err))
		return
	}

	// Map response to model
	data.UUID = types.StringValue(createdDomain.UUID)
	data.Name = types.StringValue(createdDomain.Name)
	data.Email = types.StringValue(createdDomain.Email)
	data.SlackChannel.Name = types.StringValue(createdDomain.SlackChannel.Name)
	data.SlackChannel.ID = types.StringValue(createdDomain.SlackChannel.ID)

	tflog.Trace(ctx, "created a data domain resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DataDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DataDomainResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get domain by UUID
	domain, err := r.client.GetDomain(data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read data domain, got error: %s", err))
		return
	}

	// Map response to model
	data.UUID = types.StringValue(domain.UUID)
	data.Name = types.StringValue(domain.Name)
	data.Email = types.StringValue(domain.Email)
	data.SlackChannel.Name = types.StringValue(domain.SlackChannel.Name)
	data.SlackChannel.ID = types.StringValue(domain.SlackChannel.ID)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DataDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DataDomainResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing domain
	domain := masthead.DataDomain{
		UUID:             data.UUID.ValueString(),
		Name:             data.Name.ValueString(),
		Email:            data.Email.ValueString(),
		SlackChannelName: data.SlackChannel.Name.ValueString(),
	}

	updatedDomain, err := r.client.UpdateDomain(domain)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update data domain, got error: %s", err))
		return
	}

	// Map response to model
	data.Name = types.StringValue(updatedDomain.Name)
	data.Email = types.StringValue(updatedDomain.Email)
	data.SlackChannel.Name = types.StringValue(updatedDomain.SlackChannel.Name)
	data.SlackChannel.ID = types.StringValue(updatedDomain.SlackChannel.ID)

	tflog.Trace(ctx, "updated a data domain resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DataDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var domain DataDomainResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &domain)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete domain
	err := r.client.DeleteDomain(domain.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete data domain, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted a data domain resource")
}

func (r *DataDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
