package masthead

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const scanJSON = `{"id":42,"name":"Orders nulls","description":"d","project":"p","table":"p.dq.orders",` +
	`"type":"ANALYSIS","frequency":86400,"processDelayHours":6,"isActive":true,"source":"API",` +
	`"createdByType":"API_TOKEN","createdAt":"2026-11-23T10:00:00Z"}`

func newScanTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(h)
	t.Cleanup(server.Close)
	token := "test-token"
	c, err := NewClient(&token, 0)
	if err != nil {
		t.Fatal(err)
	}
	c.HostURL = server.URL
	return c
}

// A create is a dry run, then the real call, both with the frequency sent by name.
func TestCreateDataScanDryRunThenCreate(t *testing.T) {
	var calls []string
	c := newScanTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got DataScanCreateRequest
		if err := json.Unmarshal(body, &got); err != nil || got.Frequency != "DAILY" {
			t.Errorf("body %s: want frequency DAILY", body)
		}
		if r.Header.Get("X-API-TOKEN") != "test-token" {
			t.Error("missing X-API-TOKEN")
		}
		calls = append(calls, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		if r.URL.Query().Get("dryRun") == "true" {
			_, _ = w.Write([]byte(`{"summary":{"table":"p.dq.orders","series":12,"warnings":["few rows"]},"scan":null}`))
			return
		}
		_, _ = w.Write([]byte(`{"summary":{"table":"p.dq.orders"},"scan":` + scanJSON + `}`))
	})

	req := DataScanCreateRequest{Table: "p.dq.orders", Name: "Orders nulls", Frequency: "DAILY"}
	summary, err := c.ValidateDataScan(req)
	if err != nil || summary.Series != 12 || len(summary.Warnings) != 1 {
		t.Fatalf("validate: %+v, %v", summary, err)
	}
	scan, _, err := c.CreateDataScan(req)
	if err != nil || scan.ID != 42 || DataScanFrequencyName(scan.Frequency) != "DAILY" {
		t.Fatalf("create: %+v, %v", scan, err)
	}
	want := []string{"POST /clientApi/data-scans?dryRun=true", "POST /clientApi/data-scans?dryRun=false"}
	if strings.Join(calls, ",") != strings.Join(want, ",") {
		t.Fatalf("calls %v, want %v", calls, want)
	}
}

// A 404 is reported as not found, so the resource can drop it from state.
func TestGetDataScanNotFound(t *testing.T) {
	c := newScanTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Scan 7 not found"}`))
	})
	if _, err := c.GetDataScan(7); !IsNotFound(err) {
		t.Fatalf("want not found, got %v", err)
	}
}

// metadata-api answers 200 with success=false when data-scan-agent is down; that must be an error.
func TestDataScanServiceUnavailable(t *testing.T) {
	c := newScanTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":false,"error":"Data scan service unavailable"}`))
	})
	if _, err := c.GetDataScan(7); err == nil || IsNotFound(err) || !strings.Contains(err.Error(), "unavailable") {
		t.Fatalf("want unavailable error, got %v", err)
	}
	if err := c.DeleteDataScan(7); err == nil {
		t.Fatal("delete: want unavailable error")
	}
}

// Update sends only the fields that are set.
func TestUpdateDataScanSendsSetFields(t *testing.T) {
	c := newScanTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.Method != "PATCH" || string(body) != `{"active":false}` {
			t.Errorf("got %s %s", r.Method, body)
		}
		_, _ = w.Write([]byte(`{"scan":` + scanJSON + `,"summary":null}`))
	})
	inactive := false
	if _, err := c.UpdateDataScan(42, DataScanUpdateRequest{Active: &inactive}); err != nil {
		t.Fatal(err)
	}
}

// A create sends the type, severity and labels; the response carries them back.
func TestCreateDataScanSendsTypeSeverityLabels(t *testing.T) {
	c := newScanTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		want := `{"table":"p.dq.orders","name":"Orders nulls","frequency":"DAILY","type":"ANALYSIS_WITH_THRESHOLD",` +
			`"severity":"P2","labels":{"owner":"data-ops"}}`
		if string(body) != want {
			t.Errorf("body %s, want %s", body, want)
		}
		_, _ = w.Write([]byte(`{"summary":{"table":"p.dq.orders"},"scan":` +
			strings.Replace(scanJSON, `"source":"API"`, `"source":"API","severity":"P2","labels":{"owner":"data-ops"}`, 1) + `}`))
	})
	threshold, severity := "ANALYSIS_WITH_THRESHOLD", "P2"
	scan, _, err := c.CreateDataScan(DataScanCreateRequest{Table: "p.dq.orders", Name: "Orders nulls", Frequency: "DAILY",
		Type: &threshold, Severity: &severity, Labels: map[string]string{"owner": "data-ops"}})
	if err != nil || scan.Severity == nil || *scan.Severity != "P2" || scan.Labels["owner"] != "data-ops" {
		t.Fatalf("create: %+v, %v", scan, err)
	}
}

// A severity is left out of an update when unset, sent as null to remove it, and sent by name to change it.
func TestUpdateDataScanSeverity(t *testing.T) {
	for _, tc := range []struct {
		name   string
		update DataScanUpdateRequest
		want   string
	}{
		{"unchanged", DataScanUpdateRequest{}, `{}`},
		{"removed", DataScanUpdateRequest{Severity: DataScanSeverityUpdate(nil)}, `{"severity":null}`},
		{"changed", DataScanUpdateRequest{Severity: DataScanSeverityUpdate(func() *string { s := "P4"; return &s }())}, `{"severity":"P4"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newScanTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				if string(body) != tc.want {
					t.Errorf("body %s, want %s", body, tc.want)
				}
				_, _ = w.Write([]byte(`{"scan":` + scanJSON + `,"summary":null}`))
			})
			if _, err := c.UpdateDataScan(42, tc.update); err != nil {
				t.Fatal(err)
			}
		})
	}
}
