package audit

import (
	"testing"
	"time"

	"go.opencensus.io/stats/view"
)

func TestReportTotalViolations(t *testing.T) {
	const wantValue = 10
	const wantRowLength = 1
	wantTags := map[string]string{
		"enforcement_action": "deny",
	}

	r, err := newStatsReporter()
	if err != nil {
		t.Fatalf("newStatsReporter() error %v", err)
	}

	err = r.reportTotalViolations("deny", wantValue)
	if err != nil {
		t.Fatalf("ReportTotalViolations error %v", err)
	}

	row := checkData(t, violationsMetricName, wantRowLength)
	got, ok := row.Data.(*view.LastValueData)
	if !ok {
		t.Error("ReportTotalViolations should have aggregation LastValue()")
	}

	for _, tag := range row.Tags {
		if tag.Value != wantTags[tag.Key.Name()] {
			t.Errorf("ReportTotalViolations tags does not match for %v", tag.Key.Name())
		}
	}

	if int64(got.Value) != wantValue {
		t.Errorf("got %q = %v, want %v", violationsMetricName, got.Value, wantValue)
	}
}

func TestReportLatency(t *testing.T) {
	const minLatency = 100 * time.Second
	const maxLatency = 500 * time.Second

	const wantLatencyCount int64 = 2
	const wantLatencyMin float64 = 100
	const wantLatencyMax float64 = 500
	const wantRowLength int = 1

	r, err := newStatsReporter()
	if err != nil {
		t.Fatalf("got newStatsReporter() error %v", err)
	}

	err = r.reportLatency(minLatency)
	if err != nil {
		t.Fatalf("got reportLatency error %v", err)
	}

	err = r.reportLatency(maxLatency)
	if err != nil {
		t.Fatalf("got reportLatency error %v", err)
	}

	row := checkData(t, auditDurationMetricName, wantRowLength)
	gotLatency, ok := row.Data.(*view.DistributionData)
	if !ok {
		t.Fatalf("got reportLatency() type %T, want %T", row.Data, &view.DistributionData{})
	}

	if gotLatency.Count != wantLatencyCount {
		t.Errorf("got %q = %v, want %v", auditDurationMetricName, gotLatency.Count, wantLatencyCount)
	}

	if gotLatency.Min != wantLatencyMin {
		t.Errorf("got %q = %v, want %v", auditDurationMetricName, gotLatency.Min, wantLatencyMin)
	}

	if gotLatency.Max != wantLatencyMax {
		t.Errorf("got %q = %v, want %v", auditDurationMetricName, gotLatency.Max, wantLatencyMax)
	}
}

func checkData(t *testing.T, name string, wantRowLength int) *view.Row {
	row, err := view.RetrieveData(name)
	if err != nil {
		t.Fatalf("got RetrieveData error: %v from %v", err, name)
	}

	if len(row) != wantRowLength {
		t.Fatalf("got row length %v, want %v", len(row), wantRowLength)
	}

	if row[0].Data == nil {
		t.Fatalf("got row[0].Data = nil, want non-nil: %+v", row)
	}
	return row[0]
}

func TestLastRestartCheck(t *testing.T) {
	wantStartTime := time.Now()
	wantEndTime := wantStartTime.Add(1 * time.Minute)
	wantStartTs := float64(wantStartTime.Unix())
	wantEndTs := float64(wantEndTime.Unix())
	const wantRowLength = 1

	r, err := newStatsReporter()
	if err != nil {
		t.Fatalf("got newStatsReporter() error %v", err)
	}

	err = r.reportRunStart(wantStartTime)
	if err != nil {
		t.Fatalf("reportRunStart error %v", err)
	}
	row := checkData(t, lastRunStartTimeMetricName, wantRowLength)
	got, ok := row.Data.(*view.LastValueData)
	if !ok {
		t.Errorf("%s should have aggregation LastValue()", lastRunStartTimeMetricName)
	}

	if len(row.Tags) != 0 {
		t.Errorf("got %q tags %v, want empty", lastRunStartTimeMetricName, row.Tags)
	}

	if got.Value != wantStartTs {
		t.Errorf("got %q = %v, want %v", lastRunStartTimeMetricName, got.Value, wantStartTs)
	}

	err = r.reportRunEnd(wantEndTime)
	if err != nil {
		t.Fatalf("reportRunEnd error %v", err)
	}
	row = checkData(t, lastRunEndTimeMetricName, wantRowLength)
	got, ok = row.Data.(*view.LastValueData)
	if !ok {
		t.Errorf("%s should have aggregation LastValue()", lastRunEndTimeMetricName)
	}

	if len(row.Tags) != 0 {
		t.Errorf("got %q tags %v, want empty", lastRunEndTimeMetricName, row.Tags)
	}

	if got.Value != wantEndTs {
		t.Errorf("got %q = %v, want %v", lastRunEndTimeMetricName, got.Value, wantEndTs)
	}
}
