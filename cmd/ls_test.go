//go:build integration

package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/bhpcv252/portmap/internal/detector"
	"github.com/bhpcv252/portmap/internal/registry"
)

func TestLsCmd_FullTable(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, []detector.ActivePort{
		{Port: 3000, PID: 1234},
		{Port: 3001, PID: 5678},
	})
	stdout, _ := captureOutput(t)
	noColor = true

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
		"3001": {Project: "myapp", Service: "api", ClaimedAt: time.Now().UTC()},
		"8000": {Project: "ml-service", Service: "inference", ClaimedAt: time.Now().UTC()},
		"5432": {Project: "shared", Service: "postgres", ClaimedAt: time.Now().UTC()},
	})

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if strings.Count(got, "● running") < 2 {
		t.Errorf("expected at least 2 running rows, got:\n%s", got)
	}
	if strings.Count(got, "○ stopped") < 2 {
		t.Errorf("expected at least 2 stopped rows, got:\n%s", got)
	}
}

func TestLsCmd_EmptyRegistry_NoActive(t *testing.T) {
	setupIntegrationWithMock(t, nil)
	stdout, _ := captureOutput(t)

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "no ports registered") {
		t.Errorf("expected empty state message, got: %q", stdout.String())
	}
}

func TestLsCmd_EmptyRegistry_WithUnclaimed(t *testing.T) {
	setupIntegrationWithMock(t, []detector.ActivePort{
		{Port: 8080, PID: 9999, Process: "nginx"},
	})
	stdout, _ := captureOutput(t)
	noColor = true

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "Unclaimed active ports:") {
		t.Errorf("expected unclaimed section, got:\n%s", got)
	}
	if !strings.Contains(got, "8080") {
		t.Errorf("expected unclaimed port 8080, got:\n%s", got)
	}
}

func TestLsCmd_ProjectFilter(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, nil)
	stdout, _ := captureOutput(t)
	lsProject = "myapp"
	noColor = true

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
		"3001": {Project: "myapp", Service: "api", ClaimedAt: time.Now().UTC()},
		"8000": {Project: "ml-service", Service: "inference", ClaimedAt: time.Now().UTC()},
	})

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if strings.Contains(got, "ml-service") {
		t.Errorf("expected ml-service filtered out, got:\n%s", got)
	}
}

func TestLsCmd_ActiveFilter(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, []detector.ActivePort{
		{Port: 3000, PID: 1234},
		{Port: 3001, PID: 5678},
	})
	stdout, _ := captureOutput(t)
	lsActive = true
	noColor = true

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
		"3001": {Project: "myapp", Service: "api", ClaimedAt: time.Now().UTC()},
		"8000": {Project: "ml-service", Service: "inference", ClaimedAt: time.Now().UTC()},
	})

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if strings.Contains(got, "ml-service") {
		t.Errorf("expected stopped ml-service filtered out, got:\n%s", got)
	}
	if strings.Count(got, "● running") < 2 {
		t.Errorf("expected 2 running rows, got:\n%s", got)
	}
}

func TestLsCmd_FreeFilter(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, []detector.ActivePort{
		{Port: 3000, PID: 1234},
		{Port: 3001, PID: 5678},
	})
	stdout, _ := captureOutput(t)
	lsFree = true
	noColor = true

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
		"3001": {Project: "myapp", Service: "api", ClaimedAt: time.Now().UTC()},
		"8000": {Project: "ml-service", Service: "inference", ClaimedAt: time.Now().UTC()},
	})

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if strings.Contains(got, "● running") {
		t.Errorf("expected running rows filtered out, got:\n%s", got)
	}
	if !strings.Contains(got, "ml-service") {
		t.Errorf("expected stopped ml-service in output, got:\n%s", got)
	}
}

func TestLsCmd_UnclaimedFilter(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, []detector.ActivePort{
		{Port: 8080, PID: 9999},
	})
	stdout, _ := captureOutput(t)
	lsUnclaimed = true
	noColor = true

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "8080") {
		t.Errorf("expected unclaimed port 8080, got:\n%s", got)
	}
}

func TestLsCmd_JSON_ValidParse(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, nil)
	stdout, _ := captureOutput(t)
	lsJSON = true

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed []map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, stdout.String())
	}
}

func TestLsCmd_JSON_StatusField(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, []detector.ActivePort{
		{Port: 3000, PID: 1234},
	})
	stdout, _ := captureOutput(t)
	lsJSON = true

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed []map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatal(err)
	}
	if len(parsed) == 0 || parsed[0]["status"] != "running" {
		t.Errorf("expected status 'running', got: %v", parsed)
	}
}

func TestLsCmd_NoColor(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, nil)
	stdout, _ := captureOutput(t)
	noColor = true

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

	if err := runLs(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(stdout.String(), "\x1b[") {
		t.Error("expected no ANSI codes with --no-color")
	}
}
