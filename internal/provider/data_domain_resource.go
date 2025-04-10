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
var (
	_ resource.Resource                = &DataDomainResource{}
	_ resource.ResourceWithImportState = &DataDomainResource{}
)

func NewDataDomainResource() resource.Resource {
	return &DataDomainResource{}
}

// DataDomainResource defines the resource implementation.
type DataDomainResource struct {
	client *masthead.Client
}

type DataDomainResourceModel struct {
	UUID             types.String `tfsdk:"uuid"`
	Name             types.String `tfsdk:"name"`
	Email            types.String `tfsdk:"email"`
	SlackChannelName types.String `tfsdk:"slack_channel_name"`
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
				MarkdownDescription: "Slack channel name associated with the data domain",
				Optional:            true,
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
	var plan DataDomainResourceModel
	var state DataDomainResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new domain
	domainRequest := masthead.DataDomain{
		Name:             plan.Name.ValueString(),
		Email:            plan.Email.ValueString(),
		SlackChannelName: plan.SlackChannelName.ValueString(),
	}

	domainResponse, err := r.client.CreateDomain(domainRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create data domain, got error: %s", err))
		return
	}

	// Map response to model
	state.UUID = types.StringValue(domainResponse.UUID)
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

func (r *DataDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var plan DataDomainResourceModel
	var state DataDomainResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get domain by UUID
	domainResponse, err := r.client.GetDomain(plan.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read data domain, got error: %s", err))
		return
	}

	// Map response to model
	state.UUID = types.StringValue(domainResponse.UUID)
	state.Name = types.StringValue(domainResponse.Name)
	state.Email = types.StringValue(domainResponse.Email)
	if domainResponse.SlackChannel != (masthead.SlackChannel{}) {
		state.SlackChannelName = types.StringValue(domainResponse.SlackChannel.Name)
	} else {
		state.SlackChannelName = types.StringNull()
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DataDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DataDomainResourceModel
	var state DataDomainResourceModel

	// Read the current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing domain
	domainRequest := masthead.DataDomain{
		UUID:             plan.UUID.ValueString(),
		Name:             plan.Name.ValueString(),
		Email:            plan.Email.ValueString(),
		SlackChannelName: plan.SlackChannelName.ValueString(),
	}

	domainResponse, err := r.client.UpdateDomain(domainRequest)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update data domain, got error: %s", err))
		return
	}

	// Map response to model
	state.UUID = types.StringValue(domainResponse.UUID)
	state.Name = types.StringValue(domainResponse.Name)
	state.Email = types.StringValue(domainResponse.Email)
	if domainResponse.SlackChannel != (masthead.SlackChannel{}) {
		state.SlackChannelName = types.StringValue(domainResponse.SlackChannel.Name)
	} else {
		state.SlackChannelName = types.StringNull()
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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
}

func (r *DataDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
