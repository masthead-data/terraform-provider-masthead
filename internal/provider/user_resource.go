package provider

import (
    "context"
    "fmt"

    masthead "github.com/masthead-data/terraform-provider-masthead/internal/client"
    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

func NewUserResource() resource.Resource {
    return &UserResource{}
}

// UserResource defines the resource implementation.
type UserResource struct {
    client *masthead.Client
}

// UserResourceModel describes the resource data model.
type UserResourceModel struct {
    Email types.String `tfsdk:"email"`
    Role  types.String `tfsdk:"role"`
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Manages a Masthead user",
        Attributes: map[string]schema.Attribute{
            "email": schema.StringAttribute{
                MarkdownDescription: "Email address of the user",
                Required:            true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            "role": schema.StringAttribute{
                MarkdownDescription: "Role of the user (e.g., USER, OWNER)",
                Required:            true,
            },
        },
    }
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data UserResourceModel

    // Read Terraform plan data into the model
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Create new user
    err := r.client.CreateUser(data.Email.ValueString(), data.Role.ValueString())
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user, got error: %s", err))
        return
    }

    // Set resource ID to the email address
    // For Masthead users, email is the unique identifier
    tflog.Trace(ctx, "created a user resource")

    // Save data into Terraform state
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var data UserResourceModel

    // Read Terraform prior state data into the model
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Get all users and find the one with matching email
    users, err := r.client.ListUsers()
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read users, got error: %s", err))
        return
    }

    // Find the user by email
    found := false
    for _, user := range users {
        if user.Email == data.Email.ValueString() {
            data.Role = types.StringValue(user.Role)
            found = true
            break
        }
    }

    // If user is not found, remove from state
    if !found {
        resp.State.RemoveResource(ctx)
        return
    }

    // Save updated data into Terraform state
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var data UserResourceModel

    // Read Terraform plan data into the model
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Update existing user role
    err := r.client.UpdateUserRole(data.Email.ValueString(), data.Role.ValueString())
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user, got error: %s", err))
        return
    }

    tflog.Trace(ctx, "updated a user resource")

    // Save updated data into Terraform state
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var data UserResourceModel

    // Read Terraform prior state data into the model
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Delete user
    err := r.client.DeleteUser(data.Email.ValueString())
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user, got error: %s", err))
        return
    }

    tflog.Trace(ctx, "deleted a user resource")
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("email"), req, resp)
}
