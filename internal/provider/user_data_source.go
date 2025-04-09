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
var _ datasource.DataSource = &UserDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

// UserDataSource defines the data source implementation.
type UserDataSource struct {
	client *masthead.Client
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch information about a Masthead user",
		Attributes: map[string]schema.Attribute{
			"email": schema.StringAttribute{
				MarkdownDescription: "Email address of the user",
				Required:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Role of the user (supported values: USER, OWNER)",
				Computed:            true,
			},
		},
	}
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config UserResourceModel
	var state UserResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all users from Masthead API
	usersResponse, err := d.client.ListUsers()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read users, got error: %s", err))
		return
	}

	// Find the user with the matching email
	found := false
	for _, user := range usersResponse {
		if config.Email.ValueString() == user.Email {
			state.Email = types.StringValue(user.Email)
			state.Role = user.Role
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"User Not Found",
			fmt.Sprintf("User with email %s was not found", config.Email.ValueString()),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
