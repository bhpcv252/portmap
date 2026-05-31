//go:build integration

package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/bhpcv252/portmap/internal/detector"
)

func TestLsCmd_NoActivePorts(t *testing.T) {
	setupMock(t, nil)
	stdout, _ := captureOutput(t)

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "no active ports") {
		t.Errorf("expected 'no active ports', got: %q", stdout.String())
	}
}

func TestLsCmd_ShowsActivePorts(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 3000, PID: 1234, Process: "node", CWD: "/home/user/myapp"},
		{Port: 8080, PID: 5678, Process: "python", CWD: "/home/user/api"},
	})
	stdout, _ := captureOutput(t)
	noColor = true

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{"3000", "1234", "node", "8080", "5678", "python"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

func TestLsCmd_ShowsColumns(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 3000, PID: 1234, Process: "node", CWD: "/home/user/myapp"},
	})
	stdout, _ := captureOutput(t)

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	for _, header := range []string{"PORT", "PID", "PROCESS", "CWD"} {
		if !strings.Contains(got, header) {
			t.Errorf("expected column header %q, got:\n%s", header, got)
		}
	}
}

func TestLsCmd_ShowsCWD(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 3000, PID: 1234, Process: "node", CWD: "/home/user/projects/myapp"},
	})
	stdout, _ := captureOutput(t)

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "/home/user/projects/myapp") {
		t.Errorf("expected CWD in output, got: %q", stdout.String())
	}
}

func TestLsCmd_JSON_Valid(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 3000, PID: 1234, Process: "node", CWD: "/home/user/myapp"},
	})
	stdout, _ := captureOutput(t)
	lsJSON = true

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed []map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, stdout.String())
	}
	if len(parsed) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(parsed))
	}
}

func TestLsCmd_JSON_Fields(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 3000, PID: 1234, Process: "node", CWD: "/home/user/myapp"},
	})
	stdout, _ := captureOutput(t)
	lsJSON = true

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed []map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatal(err)
	}

	row := parsed[0]
	for _, field := range []string{"port", "pid", "process", "cwd"} {
		if _, ok := row[field]; !ok {
			t.Errorf("expected field %q in JSON output, got: %v", field, row)
		}
	}
}

func TestLsCmd_JSON_EmptyArray(t *testing.T) {
	setupMock(t, nil)
	stdout, _ := captureOutput(t)
	lsJSON = true

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed []map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("expected valid JSON array, got: %v\n%s", err, stdout.String())
	}
	if len(parsed) != 0 {
		t.Errorf("expected empty JSON array, got %d entries", len(parsed))
	}
}

func TestLsCmd_NoColor(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 3000, PID: 1234, Process: "node"},
	})
	stdout, _ := captureOutput(t)
	noColor = true

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(stdout.String(), "\x1b[") {
		t.Error("expected no ANSI codes with --no-color")
	}
}

func TestLsCmd_ColumnAlignment(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 80, PID: 1, Process: "nginx"},
		{Port: 30000, PID: 99999, Process: "python"},
	})
	stdout, _ := captureOutput(t)
	noColor = true

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected header + 2 data rows, got %d lines:\n%s", len(lines), stdout.String())
	}

	// PID column position must be the same in both data rows
	pos0 := strings.Index(lines[1], "1 ")
	pos1 := strings.Index(lines[2], "99999")
	if pos0 != pos1 {
		t.Errorf("PID column misaligned: row1 pos=%d, row2 pos=%d\n%s", pos0, pos1, stdout.String())
	}
}
