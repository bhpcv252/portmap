//go:build integration

package cmd

import (
	"strings"
	"testing"

	"github.com/bhpcv252/portmap/internal/detector"
)

func TestSuggestCmd_Default(t *testing.T) {
	setupMock(t, nil)
	stdout, _ := captureOutput(t)

	if err := runSuggest(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "3000") {
		t.Errorf("expected 3000 to be suggested, got: %q", stdout.String())
	}
}

func TestSuggestCmd_SkipsActivePort(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 3000, PID: 9999},
	})
	stdout, _ := captureOutput(t)

	if err := runSuggest(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "3001") {
		t.Errorf("expected 3001 (skipping active 3000), got: %q", stdout.String())
	}
}

func TestSuggestCmd_SkipsMultipleActivePorts(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 3000, PID: 1},
		{Port: 3001, PID: 2},
		{Port: 3002, PID: 3},
	})
	stdout, _ := captureOutput(t)

	if err := runSuggest(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "3003") {
		t.Errorf("expected 3003 after skipping 3000-3002, got: %q", stdout.String())
	}
}

func TestSuggestCmd_FromTo(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 8000, PID: 1},
	})
	stdout, _ := captureOutput(t)
	suggestFrom = 8000
	suggestTo = 8002

	if err := runSuggest(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "8001") {
		t.Errorf("expected 8001 (8000 is active), got: %q", stdout.String())
	}
}

func TestSuggestCmd_Count(t *testing.T) {
	setupMock(t, nil)
	stdout, _ := captureOutput(t)
	suggestCount = 3

	if err := runSuggest(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{"3000", "3001", "3002"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %s in output, got: %q", want, got)
		}
	}
}

func TestSuggestCmd_RangeExhausted(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 8000, PID: 1},
		{Port: 8001, PID: 2},
	})
	_, stderr := captureOutput(t)
	suggestFrom = 8000
	suggestTo = 8001

	err := runSuggest(nil, nil)
	if err == nil {
		t.Fatal("expected error when range exhausted")
	}
	if !strings.Contains(stderr.String(), "error:") {
		t.Errorf("expected error message on stderr, got: %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "8000-8001") {
		t.Errorf("expected range in error message, got: %q", stderr.String())
	}
}

func TestSuggestCmd_CountExceedsAvailable(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 8000, PID: 1},
	})
	stdout, _ := captureOutput(t)
	suggestFrom = 8000
	suggestTo = 8001
	suggestCount = 3

	if err := runSuggest(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "8001") {
		t.Errorf("expected 8001 in output, got: %q", got)
	}
	if !strings.Contains(got, "only") {
		t.Errorf("expected 'only' note about limited availability, got: %q", got)
	}
}

func TestSuggestCmd_HintOnExhaustion(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 9000, PID: 1},
	})
	_, stderr := captureOutput(t)
	suggestFrom = 9000
	suggestTo = 9000

	runSuggest(nil, nil)

	if !strings.Contains(stderr.String(), "--from") {
		t.Errorf("expected --from hint in error output, got: %q", stderr.String())
	}
}
