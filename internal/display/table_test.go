package display

import (
	"encoding/json"
	"strings"
	"testing"
)

func render(t *testing.T, rows []Row, unclaimed []Row) string {
	t.Helper()
	var buf strings.Builder
	if err := RenderTable(&buf, rows, unclaimed, true); err != nil {
		t.Fatalf("RenderTable: %v", err)
	}
	return buf.String()
}

func renderJSON(t *testing.T, rows []Row, unclaimed []Row) string {
	t.Helper()
	var buf strings.Builder
	if err := RenderJSON(&buf, rows, unclaimed); err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}
	return buf.String()
}

func TestRenderTable_AllStatuses(t *testing.T) {
	rows := []Row{
		{Port: "3000", Project: "myapp", Service: "frontend", Status: StatusRunning},
		{Port: "3001", Project: "myapp", Service: "api", Status: StatusStopped},
		{Port: "3002", Project: "myapp", Service: "worker", Status: StatusConflict},
	}
	unclaimed := []Row{{Port: "8080"}}

	got := render(t, rows, unclaimed)

	for _, want := range []string{"● running", "○ stopped", "⚠ conflict", "● running (unclaimed)"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

func TestRenderTable_Empty(t *testing.T) {
	got := render(t, nil, nil)

	if !strings.Contains(got, "no ports registered") {
		t.Errorf("expected empty state message, got: %q", got)
	}
}

func TestRenderTable_OnlyClaims_NoActive(t *testing.T) {
	rows := []Row{
		{Port: "3000", Project: "myapp", Service: "frontend", Status: StatusStopped},
		{Port: "3001", Project: "myapp", Service: "api", Status: StatusStopped},
		{Port: "3002", Project: "myapp", Service: "worker", Status: StatusStopped},
	}

	got := render(t, rows, nil)

	if strings.Contains(got, "Unclaimed") {
		t.Error("expected no unclaimed section when unclaimed is empty")
	}
	if strings.Count(got, "○ stopped") != 3 {
		t.Errorf("expected 3 stopped rows, got:\n%s", got)
	}
}

func TestRenderTable_UnclaimedSection(t *testing.T) {
	rows := []Row{
		{Port: "3000", Project: "myapp", Service: "frontend", Status: StatusStopped},
	}
	unclaimed := []Row{{Port: "8080"}}

	got := render(t, rows, unclaimed)

	if !strings.Contains(got, "Unclaimed active ports:") {
		t.Errorf("expected unclaimed section, got:\n%s", got)
	}
	if !strings.Contains(got, "8080") {
		t.Errorf("expected unclaimed port 8080, got:\n%s", got)
	}
}

func TestRenderTable_NoUnclaimedSection(t *testing.T) {
	rows := []Row{
		{Port: "3000", Project: "myapp", Service: "frontend", Status: StatusStopped},
	}

	got := render(t, rows, []Row{})

	if strings.Contains(got, "Unclaimed") {
		t.Error("expected no unclaimed section")
	}
}

func TestRenderTable_FilterProject(t *testing.T) {
	rows := []Row{
		{Port: "3000", Project: "myapp", Service: "frontend", Status: StatusStopped},
		{Port: "3001", Project: "myapp", Service: "api", Status: StatusStopped},
	}

	got := render(t, rows, nil)

	if strings.Contains(got, "otherapp") {
		t.Error("expected otherapp to be absent from output")
	}
	if strings.Count(got, "myapp") < 2 {
		t.Errorf("expected 2 myapp rows, got:\n%s", got)
	}
}

func TestRenderTable_FlagActive(t *testing.T) {
	rows := []Row{
		{Port: "3000", Project: "myapp", Service: "frontend", Status: StatusRunning},
	}

	got := render(t, rows, nil)

	if !strings.Contains(got, "● running") {
		t.Error("expected running row in output")
	}
	if strings.Contains(got, "○ stopped") {
		t.Error("expected no stopped rows")
	}
}

func TestRenderTable_FlagFree(t *testing.T) {
	rows := []Row{
		{Port: "3001", Project: "myapp", Service: "api", Status: StatusStopped},
	}

	got := render(t, rows, nil)

	if !strings.Contains(got, "○ stopped") {
		t.Error("expected stopped row in output")
	}
	if strings.Contains(got, "● running") {
		t.Error("expected no running rows")
	}
}

func TestRenderTable_FlagUnclaimed(t *testing.T) {
	unclaimed := []Row{{Port: "8080"}}

	got := render(t, []Row{}, unclaimed)

	if !strings.Contains(got, "8080") {
		t.Errorf("expected unclaimed port 8080, got: %q", got)
	}
}

func TestRenderTable_JSON(t *testing.T) {
	rows := []Row{
		{
			Port:        "3000",
			Project:     "myapp",
			Service:     "frontend",
			Status:      StatusStopped,
			Description: "Next.js",
		},
		{Port: "3001", Project: "myapp", Service: "api", Status: StatusRunning},
	}
	unclaimed := []Row{{Port: "8080"}}

	got := renderJSON(t, rows, unclaimed)

	var parsed []map[string]interface{}
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, got)
	}
	if len(parsed) != 3 {
		t.Fatalf("expected 3 entries (2 rows + 1 unclaimed), got %d", len(parsed))
	}
}

func TestRenderTable_JSON_Fields(t *testing.T) {
	rows := []Row{
		{
			Port:        "3000",
			Project:     "myapp",
			Service:     "frontend",
			Status:      StatusRunning,
			Description: "Next.js",
		},
	}

	got := renderJSON(t, rows, nil)

	var parsed []map[string]interface{}
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatal(err)
	}

	row := parsed[0]
	for _, field := range []string{"port", "project", "service", "status", "description"} {
		if _, ok := row[field]; !ok {
			t.Errorf("expected field %q in JSON output", field)
		}
	}
	if row["status"] != "running" {
		t.Errorf("expected status 'running', got %v", row["status"])
	}
}

func TestRenderTable_NoColor(t *testing.T) {
	rows := []Row{
		{Port: "3000", Project: "myapp", Service: "frontend", Status: StatusRunning},
	}

	got := render(t, rows, nil)

	if strings.Contains(got, "\x1b[") {
		t.Error("expected no ANSI escape codes in no-color output")
	}
}

func TestRenderTable_ColumnAlignment(t *testing.T) {
	rows := []Row{
		{Port: "80", Project: "myapp", Service: "frontend", Status: StatusStopped},
		{Port: "30000", Project: "myapp", Service: "api", Status: StatusStopped},
	}

	got := render(t, rows, nil)
	lines := strings.Split(strings.TrimSpace(got), "\n")

	if len(lines) < 3 {
		t.Fatalf("expected header + 2 data rows, got %d lines", len(lines))
	}

	// STATUS column must start at the same position in both data rows
	pos0 := strings.Index(lines[1], "○")
	pos1 := strings.Index(lines[2], "○")
	if pos0 != pos1 {
		t.Errorf("status column misaligned: row1 pos=%d, row2 pos=%d\n%s", pos0, pos1, got)
	}
}
