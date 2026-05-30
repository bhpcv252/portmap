//go:build integration

package cmd

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bhpcv252/portmap/internal/detector"
	"github.com/bhpcv252/portmap/internal/registry"
)

func TestCheckCmd_Free(t *testing.T) {
	setupIntegrationWithMock(t, nil)
	stdout, _ := captureOutput(t)

	if err := runCheck(nil, []string{"3000"}); err != nil {
		t.Fatalf("expected exit 0 for free port, got: %v", err)
	}
	if !strings.Contains(stdout.String(), "free") {
		t.Errorf("expected 'free' in output, got: %q", stdout.String())
	}
}

func TestCheckCmd_ClaimedRunning(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, []detector.ActivePort{
		{Port: 3000, PID: 84312, Process: "node"},
	})
	stdout, _ := captureOutput(t)
	noColor = true

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {
			Project:     "myapp",
			Service:     "frontend",
			Description: "Next.js dev server",
			ClaimedAt:   time.Date(2025, 5, 10, 14, 22, 0, 0, time.UTC),
			Path:        "/home/user/projects/myapp",
		},
	})

	err := runCheck(nil, []string{"3000"})
	var exitErr *ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("expected ExitError{1}, got: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{"● running", "myapp", "frontend", "84312"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

func TestCheckCmd_ClaimedStopped(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, nil)
	stdout, _ := captureOutput(t)
	noColor = true

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {
			Project:   "myapp",
			Service:   "frontend",
			ClaimedAt: time.Date(2025, 5, 10, 14, 22, 0, 0, time.UTC),
		},
	})

	err := runCheck(nil, []string{"3000"})
	var exitErr *ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("expected ExitError{1}, got: %v", err)
	}

	got := stdout.String()
	if strings.Contains(got, "pid:") {
		t.Errorf("expected no pid line for stopped port, got:\n%s", got)
	}
	if !strings.Contains(got, "myapp") {
		t.Errorf("expected claim fields in output, got:\n%s", got)
	}
}

func TestCheckCmd_ActiveUnclaimed(t *testing.T) {
	setupIntegrationWithMock(t, []detector.ActivePort{
		{Port: 3000, PID: 84312, Process: "node"},
	})
	stdout, _ := captureOutput(t)
	noColor = true

	err := runCheck(nil, []string{"3000"})
	var exitErr *ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("expected ExitError{1}, got: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "running (unclaimed)") {
		t.Errorf("expected 'running (unclaimed)' in output, got:\n%s", got)
	}
	if !strings.Contains(got, "84312") {
		t.Errorf("expected pid in output, got:\n%s", got)
	}
}

func TestCheckCmd_MissingPort(t *testing.T) {
	setupIntegrationWithMock(t, nil)

	if err := runCheck(nil, []string{"notaport"}); err == nil {
		t.Fatal("expected error for invalid port argument")
	}
}
