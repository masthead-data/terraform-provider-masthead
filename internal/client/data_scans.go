package masthead

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// DataScanFrequencies maps the API's frequency names to their period in seconds. Requests send the name;
// responses carry the seconds.
var DataScanFrequencies = map[string]int64{
	"HOURLY":         3600,
	"EVERY_6_HOURS":  21600,
	"EVERY_12_HOURS": 43200,
	"DAILY":          86400,
}

// DataScanTypes are how Masthead evaluates a scan's rows: ANALYSIS learns each metric's normal range,
// ANALYSIS_WITH_THRESHOLD takes the limits from each row's lower_threshold / upper_threshold columns.
var DataScanTypes = []string{"ANALYSIS", "ANALYSIS_WITH_THRESHOLD"}

// DataScanSeverities are the severities a scan can give its incidents, as on Masthead incidents.
var DataScanSeverities = []string{"P1", "P2", "P3", "P4", "P5"}

// DataScanFrequencyName returns the API name of a period in seconds, or "" if there is none.
func DataScanFrequencyName(seconds int64) string {
	for name, s := range DataScanFrequencies {
		if s == seconds {
			return name
		}
	}
	return ""
}

// DataScan is a Masthead data scan: a BigQuery table of metric rows that Masthead checks every period.
type DataScan struct {
	ID                     int64             `json:"id"`
	Name                   string            `json:"name"`
	Description            *string           `json:"description"`
	Project                string            `json:"project"`
	Table                  *string           `json:"table"`
	Type                   string            `json:"type"`
	Frequency              int64             `json:"frequency"`
	ProcessDelayHours      int64             `json:"processDelayHours"`
	IsActive               bool              `json:"isActive"`
	Source                 string            `json:"source"`
	Severity               *string           `json:"severity"`
	Labels                 map[string]string `json:"labels"`
	CreatedByType          *string           `json:"createdByType"`
	CreatedAt              time.Time         `json:"createdAt"`
	LastProcessedDatetime  *time.Time        `json:"lastProcessedDatetime"`
	NextProcessingDatetime *time.Time        `json:"nextProcessingDatetime"`
}

// DataScanSummary is what a create or a table change reads from the scan table: the dry-run result.
type DataScanSummary struct {
	Table          string   `json:"table"`
	RowCount       int64    `json:"rowCount"`
	Tables         int64    `json:"tables"`
	Rules          int64    `json:"rules"`
	Series         int64    `json:"series"`
	BytesProcessed int64    `json:"bytesProcessed"`
	Warnings       []string `json:"warnings"`
}

// DataScanCreateRequest is the body of POST /clientApi/data-scans.
type DataScanCreateRequest struct {
	Table             string            `json:"table"`
	Name              string            `json:"name"`
	Frequency         string            `json:"frequency"`
	ProcessDelayHours *int64            `json:"processDelayHours,omitempty"`
	Description       *string           `json:"description,omitempty"`
	Type              *string           `json:"type,omitempty"`
	Severity          *string           `json:"severity,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
}

// DataScanUpdateRequest is the body of PATCH /clientApi/data-scans/{id}. Nil fields are left unchanged.
// Changing Table or Frequency resets the scan's history and incidents.
type DataScanUpdateRequest struct {
	Table             *string `json:"table,omitempty"`
	Frequency         *string `json:"frequency,omitempty"`
	ProcessDelayHours *int64  `json:"processDelayHours,omitempty"`
	Description       *string `json:"description,omitempty"`
	Active            *bool   `json:"active,omitempty"`
	// Severity is left out when nil; DataScanSeverityUpdate sets a new one or removes it.
	Severity json.RawMessage `json:"severity,omitempty"`
}

// DataScanSeverityUpdate is the update value for a severity: the new severity, or JSON null to remove it.
// The API keeps the severity when the field is missing, so a removal must send an explicit null.
func DataScanSeverityUpdate(severity *string) json.RawMessage {
	if severity == nil {
		return json.RawMessage("null")
	}
	v, _ := json.Marshal(*severity)
	return v
}

type dataScanWriteResponse struct {
	Scan    *DataScan        `json:"scan"`
	Summary *DataScanSummary `json:"summary"`
	Error   *string          `json:"error"`
}

// unavailable returns the error metadata-api sends with a 200 when data-scan-agent can't be reached.
func unavailable(body []byte) error {
	var r struct {
		Success *bool   `json:"success"`
		Error   *string `json:"error"`
	}
	if json.Unmarshal(body, &r) == nil && r.Success != nil && !*r.Success && r.Error != nil {
		return fmt.Errorf("data scans: %s", *r.Error)
	}
	return nil
}

func (c *Client) writeDataScan(method, url string, payload any) (*dataScanWriteResponse, error) {
	rb, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	if err := unavailable(body); err != nil {
		return nil, err
	}
	out := dataScanWriteResponse{}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ValidateDataScan dry-runs a create: the API checks the table and its rows against the scan contract and
// returns what the scan would read, without creating it. Contract violations come back as an error.
func (c *Client) ValidateDataScan(scan DataScanCreateRequest) (*DataScanSummary, error) {
	out, err := c.writeDataScan("POST", fmt.Sprintf("%s/clientApi/data-scans?dryRun=true", c.HostURL), scan)
	if err != nil {
		return nil, err
	}
	if out.Summary == nil {
		return nil, ErrEmptyValue
	}
	return out.Summary, nil
}

// CreateDataScan creates a data scan. Call ValidateDataScan first to fail without side effects.
func (c *Client) CreateDataScan(scan DataScanCreateRequest) (*DataScan, *DataScanSummary, error) {
	out, err := c.writeDataScan("POST", fmt.Sprintf("%s/clientApi/data-scans?dryRun=false", c.HostURL), scan)
	if err != nil {
		return nil, nil, err
	}
	if out.Scan == nil || out.Scan.ID == 0 {
		return nil, nil, ErrEmptyValue
	}
	return out.Scan, out.Summary, nil
}

// GetDataScan returns one data scan. A missing scan is an *HTTPError with IsNotFound(err) true.
func (c *Client) GetDataScan(id int64) (*DataScan, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/clientApi/data-scans/%d", c.HostURL, id), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	if err := unavailable(body); err != nil {
		return nil, err
	}
	scan := DataScan{}
	if err := json.Unmarshal(body, &scan); err != nil {
		return nil, err
	}
	if scan.ID == 0 {
		return nil, ErrEmptyValue
	}
	return &scan, nil
}

// UpdateDataScan changes a data scan in place.
func (c *Client) UpdateDataScan(id int64, update DataScanUpdateRequest) (*DataScan, error) {
	out, err := c.writeDataScan("PATCH", fmt.Sprintf("%s/clientApi/data-scans/%d", c.HostURL, id), update)
	if err != nil {
		return nil, err
	}
	if out.Scan == nil || out.Scan.ID == 0 {
		return nil, ErrEmptyValue
	}
	return out.Scan, nil
}

// DeleteDataScan deletes a data scan with its results and open incidents. The scan table stays.
func (c *Client) DeleteDataScan(id int64) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/clientApi/data-scans/%d", c.HostURL, id), nil)
	if err != nil {
		return err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return err
	}
	return unavailable(body)
}

// ListDataScans returns the tenant's data scans, ordered by name.
func (c *Client) ListDataScans() ([]DataScan, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/clientApi/data-scans/list", c.HostURL), nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	if err := unavailable(body); err != nil {
		return nil, err
	}
	out := struct {
		Values []DataScan `json:"values"`
	}{}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out.Values, nil
}
