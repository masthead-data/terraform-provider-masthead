package provider

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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
	Schedule          types.String `tfsdk:"schedule"`
	ScheduleTimezone  types.String `tfsdk:"schedule_timezone"`
	ProcessDelayHours types.Int64  `tfsdk:"process_delay_hours"`
	Active            types.Bool   `tfsdk:"active"`
	Project           types.String `tfsdk:"project"`
	Type              types.String `tfsdk:"type"`
	Severity          types.String `tfsdk:"severity"`
	Labels            types.Map    `tfsdk:"labels"`
	Source            types.String `tfsdk:"source"`
}

// Label rules, as data-scan-agent checks them (the Google Cloud label rules).
var (
	labelKeyPattern   = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,62}$`)
	labelValuePattern = regexp.MustCompile(`^[a-z0-9_-]{0,63}$`)
)

const labelsMax = 64

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
			"so contract errors fail the apply before anything is created. Changing `name`, `table`, `schedule`, " +
			"`schedule_timezone`, `frequency`, `type` or `labels` replaces the scan, which drops its history and incidents.",
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
			"schedule": schema.StringAttribute{
				MarkdownDescription: "When periods start and end: a five-field cron (minute hour day-of-month month " +
					"day-of-week) or `@hourly`, `@daily`, `@weekly`, `@monthly`. Each period runs from one tick to the " +
					"next and has one row per metric: `0 0 * * SUN-FRI` makes Friday's period Friday plus Saturday, " +
					"`0 0 * * MON` is weekly, `0 0 1 * *` monthly, `0 4,6,15 * * *` three periods a day. Periods are " +
					"1 hour to 31 days. Set this or `frequency`; read back when `frequency` is set. Changing it " +
					"replaces the scan.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schedule_timezone": schema.StringAttribute{
				MarkdownDescription: "IANA time zone the schedule runs in, for example `Europe/Paris`. Defaults to `UTC`. " +
					"Changing it replaces the scan.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"frequency": schema.StringAttribute{
				MarkdownDescription: "Shorthand for a UTC `schedule`: one of `" + strings.Join(frequencyNames(), "`, `") +
					"` (`DAILY` is midnight UTC). Set this or `schedule`; read back as the frequency the schedule is " +
					"the shorthand of, or empty for any other schedule. Changing it replaces the scan.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
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
				MarkdownDescription: "How Masthead evaluates the rows: `ANALYSIS` (the default; Masthead learns each metric's " +
					"normal range) or `ANALYSIS_WITH_THRESHOLD` (each row sets its own limits in `lower_threshold` and " +
					"`upper_threshold`; an empty one means no limit on that side). Changing it replaces the scan.",
				Optional:      true,
				Computed:      true,
				Default:       stringdefault.StaticString("ANALYSIS"),
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"severity": schema.StringAttribute{
				MarkdownDescription: "Severity of every incident the scan opens: one of `" +
					strings.Join(masthead.DataScanSeverities, "`, `") + "`. A change applies to new incidents. " +
					"Leave it out for no severity. Updated in place.",
				Optional: true,
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: "Key/value labels, for example for ownership. Keys: up to 63 lowercase letters, digits, " +
					"`_` and `-`, starting with a letter; values: up to 63 of the same characters. At most 64 labels. " +
					"Labels are set when the scan is created, so changing them replaces the scan.",
				ElementType:   types.StringType,
				Optional:      true,
				PlanModifiers: []planmodifier.Map{mapplanmodifier.RequiresReplace()},
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
	if !cfg.Schedule.IsUnknown() && !cfg.Frequency.IsUnknown() {
		switch {
		case cfg.Schedule.IsNull() && cfg.Frequency.IsNull():
			resp.Diagnostics.AddAttributeError(path.Root("schedule"), "Missing schedule", "Set schedule or frequency.")
		case !cfg.Schedule.IsNull() && !cfg.Frequency.IsNull():
			resp.Diagnostics.AddAttributeError(path.Root("frequency"), "Conflicting schedule",
				"Set schedule or frequency, not both.")
		}
	}
	if !cfg.Schedule.IsUnknown() && !cfg.Schedule.IsNull() && strings.TrimSpace(cfg.Schedule.ValueString()) == "" {
		resp.Diagnostics.AddAttributeError(path.Root("schedule"), "Invalid schedule", "The schedule is empty.")
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
	if !cfg.Type.IsUnknown() && !cfg.Type.IsNull() && !slices.Contains(masthead.DataScanTypes, cfg.Type.ValueString()) {
		resp.Diagnostics.AddAttributeError(path.Root("type"), "Invalid type",
			fmt.Sprintf("Use one of %s.", strings.Join(masthead.DataScanTypes, ", ")))
	}
	if !cfg.Severity.IsUnknown() && !cfg.Severity.IsNull() && !slices.Contains(masthead.DataScanSeverities, cfg.Severity.ValueString()) {
		resp.Diagnostics.AddAttributeError(path.Root("severity"), "Invalid severity",
			fmt.Sprintf("Use one of %s.", strings.Join(masthead.DataScanSeverities, ", ")))
	}
	if !cfg.Labels.IsUnknown() && !cfg.Labels.IsNull() {
		labels := map[string]types.String{}
		resp.Diagnostics.Append(cfg.Labels.ElementsAs(ctx, &labels, false)...)
		if len(labels) > labelsMax {
			resp.Diagnostics.AddAttributeError(path.Root("labels"), "Too many labels",
				fmt.Sprintf("A scan can have at most %d labels, got %d.", labelsMax, len(labels)))
		}
		for key, value := range labels {
			if !labelKeyPattern.MatchString(key) {
				resp.Diagnostics.AddAttributeError(path.Root("labels"), "Invalid label key",
					fmt.Sprintf("%q: keys are up to 63 lowercase letters, digits, _ and -, starting with a letter.", key))
			}
			if !value.IsUnknown() && !labelValuePattern.MatchString(value.ValueString()) {
				resp.Diagnostics.AddAttributeError(path.Root("labels"), "Invalid label value",
					fmt.Sprintf("%q: values are up to 63 lowercase letters, digits, _ and -.", value.ValueString()))
			}
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
	m.Schedule = types.StringValue(scan.Schedule)
	m.ScheduleTimezone = types.StringValue(scan.ScheduleTimezone)
	// the frequency a UTC schedule is the shorthand of; null for a custom schedule
	if name := masthead.DataScanScheduleFrequency(scan.Schedule, scan.ScheduleTimezone); name != "" {
		m.Frequency = types.StringValue(name)
	} else {
		m.Frequency = types.StringNull()
	}
	m.ProcessDelayHours = types.Int64Value(scan.ProcessDelayHours)
	m.Active = types.BoolValue(scan.IsActive)
	m.Project = types.StringValue(scan.Project)
	m.Type = types.StringValue(scan.Type)
	m.Severity = types.StringPointerValue(scan.Severity)
	// no labels reads as null, so a config without labels shows no diff
	if len(scan.Labels) == 0 {
		m.Labels = types.MapNull(types.StringType)
	} else {
		labels := make(map[string]attr.Value, len(scan.Labels))
		for k, v := range scan.Labels {
			labels[k] = types.StringValue(v)
		}
		m.Labels = types.MapValueMust(types.StringType, labels)
	}
	m.Source = types.StringValue(scan.Source)
}

// optionalString is the value of a planned attribute, or nil when it is null or not known yet.
func optionalString(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return v.ValueStringPointer()
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
		Table:            plan.Table.ValueString(),
		Name:             plan.Name.ValueString(),
		Frequency:        optionalString(plan.Frequency),
		Schedule:         optionalString(plan.Schedule),
		ScheduleTimezone: optionalString(plan.ScheduleTimezone),
		Description:      plan.Description.ValueStringPointer(),
	}
	if !plan.ProcessDelayHours.IsUnknown() && !plan.ProcessDelayHours.IsNull() {
		body.ProcessDelayHours = plan.ProcessDelayHours.ValueInt64Pointer()
	}
	if !plan.Type.IsUnknown() && !plan.Type.IsNull() {
		body.Type = plan.Type.ValueStringPointer()
	}
	body.Severity = plan.Severity.ValueStringPointer()
	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
		resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &body.Labels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
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

	// name, table, schedule, schedule_timezone, frequency, type and labels force replacement, so only these can
	// change here.
	update := masthead.DataScanUpdateRequest{Active: plan.Active.ValueBoolPointer()}
	if !plan.Description.Equal(state.Description) {
		d := plan.Description.ValueString() // null config clears the description
		update.Description = &d
	}
	if !plan.ProcessDelayHours.IsUnknown() && !plan.ProcessDelayHours.IsNull() {
		update.ProcessDelayHours = plan.ProcessDelayHours.ValueInt64Pointer()
	}
	if !plan.Severity.Equal(state.Severity) {
		update.Severity = masthead.DataScanSeverityUpdate(plan.Severity.ValueStringPointer()) // null config removes it
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
