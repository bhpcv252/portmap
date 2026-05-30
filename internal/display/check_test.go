package display

import (
	"strings"
	"testing"
	"time"
)

func renderCheck(t *testing.T, info CheckInfo) string {
	t.Helper()
	var buf strings.Builder
	if err := RenderCheck(&buf, info, true); err != nil {
		t.Fatalf("RenderCheck: %v", err)
	}
	return buf.String()
}

func TestRenderCheck_ClaimedAndRunning(t *testing.T) {
	info := CheckInfo{
		Port:        "3000",
		Project:     "myapp",
		Service:     "frontend",
		Description: "Next.js dev server",
		ClaimedAt:   time.Date(2025, 5, 10, 14, 22, 0, 0, time.UTC),
		Path:        "/home/user/projects/myapp",
		PID:         84312,
		Status:      StatusRunning,
	}

	got := renderCheck(t, info)

	for _, want := range []string{
		"port 3000", "● running", "myapp", "frontend",
		"Next.js dev server", "2025-05-10 14:22", "/home/user/projects/myapp", "84312",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

func TestRenderCheck_ClaimedNotRunning(t *testing.T) {
	info := CheckInfo{
		Port:        "3000",
		Project:     "myapp",
		Service:     "frontend",
		Description: "Next.js dev server",
		ClaimedAt:   time.Now().UTC(),
		Path:        "/home/user/projects/myapp",
		Status:      StatusStopped,
	}

	got := renderCheck(t, info)

	if strings.Contains(got, "pid:") {
		t.Errorf("expected no pid line when not running, got:\n%s", got)
	}
	if !strings.Contains(got, "myapp") {
		t.Errorf("expected claim fields in output, got:\n%s", got)
	}
}

func TestRenderCheck_ActiveUnclaimed(t *testing.T) {
	info := CheckInfo{
		Port:    "3000",
		PID:     84312,
		Process: "node",
		Status:  StatusRunning,
	}

	got := renderCheck(t, info)

	for _, want := range []string{"running (unclaimed)", "84312", "node"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
	if strings.Contains(got, "project:") {
		t.Errorf("expected no claim fields for unclaimed port, got:\n%s", got)
	}
}

func TestRenderCheck_Free(t *testing.T) {
	got := renderCheck(t, CheckInfo{Port: "3000", Status: StatusFree})

	if !strings.Contains(got, "free") {
		t.Errorf("expected 'free' in output, got:\n%s", got)
	}
	if !strings.Contains(got, "no claim registered, nothing running") {
		t.Errorf("expected free detail message, got:\n%s", got)
	}
}

func TestRenderCheck_ClaimedAtFormatting(t *testing.T) {
	info := CheckInfo{
		Port:      "3000",
		Project:   "myapp",
		Service:   "frontend",
		ClaimedAt: time.Date(2025, 5, 10, 14, 22, 0, 0, time.UTC),
		Status:    StatusStopped,
	}

	got := renderCheck(t, info)

	if !strings.Contains(got, "2025-05-10 14:22") {
		t.Errorf("expected formatted date '2025-05-10 14:22', got:\n%s", got)
	}
}
