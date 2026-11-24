package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	masthead "github.com/masthead-data/terraform-provider-masthead/internal/client"
)

var _ datasource.DataSource = &DataScanDataSource{}

func NewDataScanDataSource() datasource.DataSource {
	return &DataScanDataSource{}
}

// DataScanDataSource reads one data scan by ID or by name.
type DataScanDataSource struct {
	client *masthead.Client
}

func (d *DataScanDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_scan"
}

func (d *DataScanDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	computed := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{MarkdownDescription: desc, Computed: true}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch a Masthead data scan by `id` or by `name`. Set exactly one of them.",
		Attributes: map[string]schema.Attribute{
			"id":                  schema.StringAttribute{MarkdownDescription: "Scan ID.", Optional: true, Computed: true},
			"name":                schema.StringAttribute{MarkdownDescription: "Scan name.", Optional: true, Computed: true},
			"description":         computed("What the scan checks."),
			"table":               computed("Scan table, `project.dataset.table`. Empty for some scans set up by Masthead."),
			"frequency":           computed("The fixed frequency the schedule is the shorthand of, for example `DAILY`, or empty."),
			"schedule":            computed("Cron whose ticks cut the scan's periods, for example `0 0 * * *`."),
			"schedule_timezone":   computed("IANA time zone of the schedule."),
			"process_delay_hours": schema.Int64Attribute{MarkdownDescription: "Hours after a period closes before it is checked.", Computed: true},
			"active":              schema.BoolAttribute{MarkdownDescription: "Whether the scan runs.", Computed: true},
			"project":             computed("Project of the scan table."),
			"type":                computed("How Masthead evaluates the rows: `ANALYSIS` or `ANALYSIS_WITH_THRESHOLD`."),
			"severity":            computed("Severity of the scan's incidents, `P1`-`P5`, or empty."),
			"labels":              schema.MapAttribute{MarkdownDescription: "Key/value labels.", ElementType: types.StringType, Computed: true},
			"source":              computed("`API` or `MANUAL` (set up by Masthead)."),
		},
	}
}

func (d *DataScanDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DataScanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DataScanResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	byID, byName := !config.ID.IsNull(), !config.Name.IsNull()
	if byID == byName {
		resp.Diagnostics.AddError("Invalid lookup", "Set exactly one of id or name.")
		return
	}

	var scan *masthead.DataScan
	if byID {
		id, err := parseID(config.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid ID", err.Error())
			return
		}
		if scan, err = d.client.GetDataScan(id); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read data scan %d, got error: %s", id, err))
			return
		}
	} else {
		scans, err := d.client.ListDataScans()
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list data scans, got error: %s", err))
			return
		}
		for i := range scans {
			if scans[i].Name == config.Name.ValueString() {
				scan = &scans[i]
				break
			}
		}
		if scan == nil {
			resp.Diagnostics.AddError("Not found", fmt.Sprintf("No data scan named %q.", config.Name.ValueString()))
			return
		}
	}

	var state DataScanResourceModel
	toState(scan, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
