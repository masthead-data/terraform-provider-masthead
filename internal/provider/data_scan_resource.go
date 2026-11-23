package provider

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	masthead "github.com/masthead-data/terraform-provider-masthead/internal/client"
)

var (
	_ resource.Resource                   = &DataScanResource{}
	_ resource.ResourceWithImportState    = &DataScanResource{}
	_ resource.ResourceWithValidateConfig = &DataScanResource{}
)

func NewDataScanResource() resource.Resource {
	return &DataScanResource{}
}

// DataScanResource manages a Masthead data scan. The scan table and the job that fills it are managed
// elsewhere (for example with google_bigquery_table and a scheduled query); this resource only tells
// Masthead to read the table.
type DataScanResource struct {
	client *masthead.Client
}

type DataScanResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Table             types.String `tfsdk:"table"`
	Frequency         types.String `tfsdk:"frequency"`
	ProcessDelayHours types.Int64  `tfsdk:"process_delay_hours"`
	Active            types.Bool   `tfsdk:"active"`
	Project           types.String `tfsdk:"project"`
	Type              types.String `tfsdk:"type"`
	Source            types.String `tfsdk:"source"`
}

func (r *DataScanResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_scan"
}

func frequencyNames() []string {
	names := make([]string, 0, len(masthead.DataScanFrequencies))
	for n := range masthead.DataScanFrequencies {
		names = append(names, n)
	}
	sort.Slice(names, func(i, j int) bool {
		return masthead.DataScanFrequencies[names[i]] < masthead.DataScanFrequencies[names[j]]
	})
	return names
}

func (r *DataScanResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Masthead data scan: a BigQuery table of metric rows (`table_reference`, `rule_name`, " +
			"`timestamp`, `value`) that Masthead checks for anomalies every period. Creating a scan first dry-runs it, " +
			"so contract errors fail the apply before anything is created. Changing `name`, `table` or `frequency` " +
			"replaces the scan, which drops its history and incidents.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Scan ID.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Scan name, 1-100 characters, unique in the tenant. Changing it replaces the scan.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "What the scan checks. Updated in place.",
				Optional:            true,
			},
			"table": schema.StringAttribute{
				MarkdownDescription: "Scan table, `project.dataset.table`. The project must be connected to Masthead, and " +
					"Masthead's data scan service account needs read access to the table. Changing it replaces the scan.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"frequency": schema.StringAttribute{
				MarkdownDescription: "How often Masthead checks the scan: one of `" + strings.Join(frequencyNames(), "`, `") +
					"`. Changing it replaces the scan.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"process_delay_hours": schema.Int64Attribute{
				MarkdownDescription: "Hours to wait after a period closes before checking it, 0-24. Defaults to a value " +
					"based on the frequency. Updated in place.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "`false` pauses the scan. Defaults to `true`. Updated in place.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"project": schema.StringAttribute{
				MarkdownDescription: "Project of the scan table.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "How Masthead evaluates the rows, for example `ANALYSIS` (learned expected range).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"source": schema.StringAttribute{
				MarkdownDescription: "Who set the scan up: `API` for scans created through the API, `MANUAL` for scans set up by Masthead.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *DataScanResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg DataScanResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !cfg.Name.IsUnknown() && !cfg.Name.IsNull() {
		if n := len([]rune(cfg.Name.ValueString())); n < 1 || n > 100 {
			resp.Diagnostics.AddAttributeError(path.Root("name"), "Invalid name", "The name must be 1-100 characters.")
		}
	}
	if !cfg.Table.IsUnknown() && !cfg.Table.IsNull() {
		if parts := strings.Split(cfg.Table.ValueString(), "."); len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
			resp.Diagnostics.AddAttributeError(path.Root("table"), "Invalid table", "Use project.dataset.table.")
		}
	}
	if !cfg.Frequency.IsUnknown() && !cfg.Frequency.IsNull() {
		if _, ok := masthead.DataScanFrequencies[cfg.Frequency.ValueString()]; !ok {
			resp.Diagnostics.AddAttributeError(path.Root("frequency"), "Invalid frequency",
				fmt.Sprintf("Use one of %s.", strings.Join(frequencyNames(), ", ")))
		}
	}
	if !cfg.ProcessDelayHours.IsUnknown() && !cfg.ProcessDelayHours.IsNull() {
		if d := cfg.ProcessDelayHours.ValueInt64(); d < 0 || d > 24 {
			resp.Diagnostics.AddAttributeError(path.Root("process_delay_hours"), "Invalid process_delay_hours", "Use 0-24.")
		}
	}
}

func (r *DataScanResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// toState maps an API scan onto the model. Description stays null when the config leaves it out.
func toState(scan *masthead.DataScan, m *DataScanResourceModel) {
	m.ID = types.StringValue(strconv.FormatInt(scan.ID, 10))
	m.Name = types.StringValue(scan.Name)
	if scan.Description != nil && *scan.Description != "" {
		m.Description = types.StringValue(*scan.Description)
	} else {
		m.Description = types.StringNull()
	}
	if scan.Table != nil {
		m.Table = types.StringValue(*scan.Table)
	}
	if name := masthead.DataScanFrequencyName(scan.Frequency); name != "" {
		m.Frequency = types.StringValue(name)
	} else {
		m.Frequency = types.StringValue(strconv.FormatInt(scan.Frequency, 10))
	}
	m.ProcessDelayHours = types.Int64Value(scan.ProcessDelayHours)
	m.Active = types.BoolValue(scan.IsActive)
	m.Project = types.StringValue(scan.Project)
	m.Type = types.StringValue(scan.Type)
	m.Source = types.StringValue(scan.Source)
}

func parseID(id string) (int64, error) {
	v, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("data scan ID %q isn't a number", id)
	}
	return v, nil
}

func (r *DataScanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DataScanResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := masthead.DataScanCreateRequest{
		Table:       plan.Table.ValueString(),
		Name:        plan.Name.ValueString(),
		Frequency:   plan.Frequency.ValueString(),
		Description: plan.Description.ValueStringPointer(),
	}
	if !plan.ProcessDelayHours.IsUnknown() && !plan.ProcessDelayHours.IsNull() {
		body.ProcessDelayHours = plan.ProcessDelayHours.ValueInt64Pointer()
	}

	// Dry run first: contract errors (missing access, wrong columns, too little history) fail here,
	// before a scan exists.
	summary, err := r.client.ValidateDataScan(body)
	if err != nil {
		resp.Diagnostics.AddError("Data scan validation failed",
			fmt.Sprintf("Masthead can't create scan %q on %s: %s", body.Name, body.Table, err))
		return
	}
	for _, w := range summary.Warnings {
		resp.Diagnostics.AddWarning("Data scan warning", fmt.Sprintf("%s: %s", body.Name, w))
	}

	scan, _, err := r.client.CreateDataScan(body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create data scan, got error: %s", err))
		return
	}

	// The API creates scans active; pause right away when the config says so.
	if !plan.Active.ValueBool() && scan.IsActive {
		inactive := false
		paused, err := r.client.UpdateDataScan(scan.ID, masthead.DataScanUpdateRequest{Active: &inactive})
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Created data scan %d but couldn't pause it: %s", scan.ID, err))
			return
		}
		scan = paused
	}

	state := plan
	toState(scan, &state)
	if plan.Description.IsNull() {
		state.Description = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DataScanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DataScanResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := parseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", err.Error())
		return
	}

	scan, err := r.client.GetDataScan(id)
	if masthead.IsNotFound(err) {
		resp.State.RemoveResource(ctx) // deleted outside Terraform: plan a re-create
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read data scan %d, got error: %s", id, err))
		return
	}

	toState(scan, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DataScanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state DataScanResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := parseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", err.Error())
		return
	}

	// name, table and frequency force replacement, so only these can change here.
	update := masthead.DataScanUpdateRequest{Active: plan.Active.ValueBoolPointer()}
	if !plan.Description.Equal(state.Description) {
		d := plan.Description.ValueString() // null config clears the description
		update.Description = &d
	}
	if !plan.ProcessDelayHours.IsUnknown() && !plan.ProcessDelayHours.IsNull() {
		update.ProcessDelayHours = plan.ProcessDelayHours.ValueInt64Pointer()
	}

	scan, err := r.client.UpdateDataScan(id, update)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update data scan %d, got error: %s", id, err))
		return
	}

	toState(scan, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DataScanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DataScanResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := parseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", err.Error())
		return
	}
	if err := r.client.DeleteDataScan(id); err != nil && !masthead.IsNotFound(err) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete data scan %d, got error: %s", id, err))
	}
}

func (r *DataScanResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
