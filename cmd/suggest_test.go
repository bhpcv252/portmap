//go:build integration

package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/bhpcv252/portmap/internal/detector"
	"github.com/bhpcv252/portmap/internal/registry"
)

func TestSuggestCmd_Default(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, nil)
	stdout, _ := captureOutput(t)

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
		"3001": {Project: "myapp", Service: "api", ClaimedAt: time.Now().UTC()},
	})

	if err := runSuggest(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "3002") {
		t.Errorf("expected 3002 to be suggested, got: %q", stdout.String())
	}
}

func TestSuggestCmd_SkipsActive(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, []detector.ActivePort{
		{Port: 3001, PID: 9999},
	})
	stdout, _ := captureOutput(t)

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

	if err := runSuggest(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "3002") {
		t.Errorf("expected 3002 (skipping active 3001), got: %q", stdout.String())
	}
}

func TestSuggestCmd_FromTo(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, nil)
	stdout, _ := captureOutput(t)
	suggestFrom = 8000
	suggestTo = 8002

	seedRegistry(t, regPath, map[string]registry.Claim{
		"8000": {Project: "myapp", Service: "svc", ClaimedAt: time.Now().UTC()},
	})

	if err := runSuggest(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "8001") {
		t.Errorf("expected 8001, got: %q", stdout.String())
	}
}

func TestSuggestCmd_Count(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, nil)
	stdout, _ := captureOutput(t)
	suggestCount = 3

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "svc", ClaimedAt: time.Now().UTC()},
		"3001": {Project: "myapp", Service: "svc2", ClaimedAt: time.Now().UTC()},
	})

	if err := runSuggest(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{"3002", "3003", "3004"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %s in output, got: %q", want, got)
		}
	}
}

func TestSuggestCmd_RangeExhausted(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, nil)
	_, stderr := captureOutput(t)
	suggestFrom = 8000
	suggestTo = 8001

	seedRegistry(t, regPath, map[string]registry.Claim{
		"8000": {Project: "myapp", Service: "a", ClaimedAt: time.Now().UTC()},
		"8001": {Project: "myapp", Service: "b", ClaimedAt: time.Now().UTC()},
	})

	err := runSuggest(nil, nil)
	if err == nil {
		t.Fatal("expected error when range exhausted")
	}
	if !strings.Contains(stderr.String(), "error:") {
		t.Errorf("expected error message, got: %q", stderr.String())
	}
}

func TestSuggestCmd_CountExceedsAvailable(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, nil)
	stdout, _ := captureOutput(t)
	suggestFrom = 8000
	suggestTo = 8001
	suggestCount = 3

	seedRegistry(t, regPath, map[string]registry.Claim{
		"8000": {Project: "myapp", Service: "a", ClaimedAt: time.Now().UTC()},
	})

	if err := runSuggest(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "8001") {
		t.Errorf("expected 8001 in output, got: %q", got)
	}
	if !strings.Contains(got, "only") {
		t.Errorf("expected note about limited availability, got: %q", got)
	}
}

func TestSuggestCmd_SkipReason(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, []detector.ActivePort{
		{Port: 3001, PID: 9999},
	})
	stdout, _ := captureOutput(t)
	suggestFrom = 3000
	suggestTo = 3003

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

	if err := runSuggest(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "3000") {
		t.Errorf("expected skip reason for claimed 3000, got: %q", got)
	}
	if !strings.Contains(got, "3001") {
		t.Errorf("expected skip reason for active 3001, got: %q", got)
	}
}
