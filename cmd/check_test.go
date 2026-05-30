//go:build integration

package cmd

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bhpcv252/portmap/internal/registry"
)

func TestCheckCmd_Free(t *testing.T) {
	setupIntegration(t)
	stdout, _ := captureOutput(t)

	err := runCheck(nil, []string{"3000"})
	if err != nil {
		t.Fatalf("expected exit 0 for free port, got: %v", err)
	}
	if !strings.Contains(stdout.String(), "free") {
		t.Errorf("expected 'free' in output, got: %q", stdout.String())
	}
}

func TestCheckCmd_ClaimedStopped(t *testing.T) {
	regPath, _ := setupIntegration(t)
	stdout, _ := captureOutput(t)

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {
			Project:     "myapp",
			Service:     "frontend",
			Description: "Next.js dev server",
			ClaimedAt:   time.Date(2025, 5, 10, 14, 22, 0, 0, time.UTC),
			Path:        "/home/user/projects/myapp",
		},
	})

	noColor = true
	err := runCheck(nil, []string{"3000"})

	var exitErr *ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("expected ExitError{1}, got: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{"myapp", "frontend", "Next.js dev server", "2025-05-10 14:22"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
	if strings.Contains(got, "pid:") {
		t.Errorf("expected no pid line for stopped port, got:\n%s", got)
	}
}

func TestCheckCmd_MissingPort(t *testing.T) {
	setupIntegration(t)

	err := runCheck(nil, []string{"notaport"})
	if err == nil {
		t.Fatal("expected error for invalid port argument")
	}
}
