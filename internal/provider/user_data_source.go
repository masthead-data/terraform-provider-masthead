package provider

import (
    "context"
    "fmt"

    masthead "github.com/masthead-data/terraform-provider-masthead/internal/client"
    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-log/tflog"
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

// UserDataSourceModel describes the data source data model.
type UserDataSourceModel struct {
    Email types.String `tfsdk:"email"`
    Role  types.String `tfsdk:"role"`
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
                MarkdownDescription: "Role of the user (e.g., admin, user)",
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
    var data UserDataSourceModel

    // Read Terraform configuration data into the model
    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Get all users from Masthead API
    users, err := d.client.ListUsers()
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read users, got error: %s", err))
        return
    }

    // Find the user with the matching email
    found := false
    for _, user := range users {
        if user.Email == data.Email.ValueString() {
            data.Role = types.StringValue(user.Role)
            found = true
            break
        }
    }

    if !found {
        resp.Diagnostics.AddError(
            "User Not Found",
            fmt.Sprintf("User with email %s was not found", data.Email.ValueString()),
        )
        return
    }

    tflog.Debug(ctx, "Read user data source", map[string]interface{}{
        "email": data.Email.ValueString(),
    })

    // Save data into Terraform state
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
