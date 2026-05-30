//go:build integration

package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/bhpcv252/portmap/internal/registry"
)

func TestLsCmd_EmptyRegistry_NoActive(t *testing.T) {
	setupIntegration(t)
	stdout, _ := captureOutput(t)

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "no ports registered") {
		t.Errorf("expected empty state message, got: %q", stdout.String())
	}
}

func TestLsCmd_ProjectFilter(t *testing.T) {
	regPath, _ := setupIntegration(t)
	stdout, _ := captureOutput(t)

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
		"3001": {Project: "myapp", Service: "api", ClaimedAt: time.Now().UTC()},
		"8000": {Project: "ml-service", Service: "inference", ClaimedAt: time.Now().UTC()},
	})

	lsProject = "myapp"
	noColor = true

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if strings.Contains(got, "ml-service") {
		t.Errorf("expected ml-service filtered out, got:\n%s", got)
	}
	if strings.Count(got, "myapp") < 2 {
		t.Errorf("expected 2 myapp rows, got:\n%s", got)
	}
}

func TestLsCmd_FreeFilter(t *testing.T) {
	regPath, _ := setupIntegration(t)
	stdout, _ := captureOutput(t)

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
		"3001": {Project: "myapp", Service: "api", ClaimedAt: time.Now().UTC()},
	})

	lsFree = true
	noColor = true

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if strings.Count(got, "○ stopped") < 2 {
		t.Errorf("expected 2 stopped rows with --free, got:\n%s", got)
	}
}

func TestLsCmd_JSON_ValidParse(t *testing.T) {
	regPath, _ := setupIntegration(t)
	stdout, _ := captureOutput(t)

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

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

func TestLsCmd_NoColor(t *testing.T) {
	regPath, _ := setupIntegration(t)
	stdout, _ := captureOutput(t)

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

	noColor = true

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(stdout.String(), "\x1b[") {
		t.Error("expected no ANSI codes with --no-color")
	}
}
