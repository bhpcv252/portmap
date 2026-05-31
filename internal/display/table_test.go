package display

import (
	"encoding/json"
	"strings"
	"testing"
)

func render(t *testing.T, rows []Row) string {
	t.Helper()
	var buf strings.Builder
	if err := RenderTable(&buf, rows, true); err != nil {
		t.Fatalf("RenderTable: %v", err)
	}
	return buf.String()
}

func renderJSON(t *testing.T, rows []Row) string {
	t.Helper()
	var buf strings.Builder
	if err := RenderJSON(&buf, rows); err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}
	return buf.String()
}

func TestRenderTable_Empty(t *testing.T) {
	got := render(t, nil)
	if !strings.Contains(got, "no active ports") {
		t.Errorf("expected 'no active ports', got: %q", got)
	}
}

func TestRenderTable_ShowsData(t *testing.T) {
	rows := []Row{
		{Port: "3000", PID: 1234, Process: "node", CWD: "/home/user/myapp"},
	}
	got := render(t, rows)

	for _, want := range []string{"3000", "1234", "node", "/home/user/myapp"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

func TestRenderTable_Headers(t *testing.T) {
	rows := []Row{
		{Port: "3000", PID: 1234, Process: "node"},
	}
	got := render(t, rows)

	for _, header := range []string{"PORT", "PID", "PROCESS", "CWD"} {
		if !strings.Contains(got, header) {
			t.Errorf("expected column header %q in output, got:\n%s", header, got)
		}
	}
}

func TestRenderTable_MultipleRows(t *testing.T) {
	rows := []Row{
		{Port: "3000", PID: 1234, Process: "node", CWD: "/home/user/a"},
		{Port: "8080", PID: 5678, Process: "python", CWD: "/home/user/b"},
	}
	got := render(t, rows)

	for _, want := range []string{"3000", "8080", "node", "python"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

func TestRenderTable_ColumnAlignment(t *testing.T) {
	rows := []Row{
		{Port: "80", PID: 1, Process: "nginx"},
		{Port: "30000", PID: 99999, Process: "python"},
	}
	got := render(t, rows)
	lines := strings.Split(strings.TrimSpace(got), "\n")

	if len(lines) < 3 {
		t.Fatalf("expected header + 2 data rows, got %d lines:\n%s", len(lines), got)
	}

	// PROCESS column must start at the same position in both data rows
	pos0 := strings.Index(lines[1], "nginx")
	pos1 := strings.Index(lines[2], "python")
	if pos0 != pos1 {
		t.Errorf("PROCESS column misaligned: row1=%d, row2=%d\n%s", pos0, pos1, got)
	}
}

func TestRenderTable_NoColor(t *testing.T) {
	rows := []Row{
		{Port: "3000", PID: 1234, Process: "node"},
	}
	got := render(t, rows)

	if strings.Contains(got, "\x1b[") {
		t.Error("expected no ANSI escape codes with noColor=true")
	}
}

func TestRenderJSON_ValidArray(t *testing.T) {
	rows := []Row{
		{Port: "3000", PID: 1234, Process: "node", CWD: "/home/user/myapp"},
	}
	got := renderJSON(t, rows)

	var parsed []map[string]interface{}
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, got)
	}
	if len(parsed) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(parsed))
	}
}

func TestRenderJSON_Fields(t *testing.T) {
	rows := []Row{
		{Port: "3000", PID: 1234, Process: "node", CWD: "/home/user/myapp"},
	}
	got := renderJSON(t, rows)

	var parsed []map[string]interface{}
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatal(err)
	}

	row := parsed[0]
	for _, field := range []string{"port", "pid", "process", "cwd"} {
		if _, ok := row[field]; !ok {
			t.Errorf("expected field %q in JSON, got: %v", field, row)
		}
	}
	if row["port"] != "3000" {
		t.Errorf("expected port '3000', got %v", row["port"])
	}
}

func TestRenderJSON_EmptyArray(t *testing.T) {
	got := renderJSON(t, []Row{})

	var parsed []map[string]interface{}
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("expected valid JSON array, got: %v", err)
	}
	if len(parsed) != 0 {
		t.Errorf("expected empty array, got %d entries", len(parsed))
	}
}

func TestRenderJSON_MultipleRows(t *testing.T) {
	rows := []Row{
		{Port: "3000", PID: 1234, Process: "node", CWD: "/a"},
		{Port: "8080", PID: 5678, Process: "python", CWD: "/b"},
	}
	got := renderJSON(t, rows)

	var parsed []map[string]interface{}
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 2 {
		t.Errorf("expected 2 entries, got %d", len(parsed))
	}
}
