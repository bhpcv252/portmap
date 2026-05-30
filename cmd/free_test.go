//go:build integration

package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/bhpcv252/portmap/internal/detector"
	"github.com/bhpcv252/portmap/internal/registry"
)

func TestFreeCmd_ClaimedNotRunning(t *testing.T) {
	regPath, reload := setupIntegrationWithMock(t, nil)
	stdout, _ := captureOutput(t)

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

	if err := runFree(nil, []string{"3000"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reload().GetClaim("3000") != nil {
		t.Error("expected claim to be removed")
	}
	if !strings.Contains(stdout.String(), "released port 3000") {
		t.Errorf("expected release message, got: %q", stdout.String())
	}
}

func TestFreeCmd_ClaimedAndRunning_Confirm(t *testing.T) {
	regPath, reload := setupIntegrationWithMock(t, []detector.ActivePort{
		{Port: 3000, PID: 1234, Process: "node"},
	})
	stdout, _ := captureOutput(t)
	injectStdin(t, "y\n")

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

	if err := runFree(nil, []string{"3000"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reload().GetClaim("3000") != nil {
		t.Error("expected claim to be removed after y confirmation")
	}
	if !strings.Contains(stdout.String(), "still running") {
		t.Errorf("expected warning about running port, got: %q", stdout.String())
	}
}

func TestFreeCmd_ClaimedAndRunning_Decline(t *testing.T) {
	regPath, reload := setupIntegrationWithMock(t, []detector.ActivePort{
		{Port: 3000, PID: 1234, Process: "node"},
	})
	captureOutput(t)
	injectStdin(t, "n\n")

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

	if err := runFree(nil, []string{"3000"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reload().GetClaim("3000") == nil {
		t.Error("expected claim to remain after n confirmation")
	}
}

func TestFreeCmd_ClaimedAndRunning_Force(t *testing.T) {
	regPath, reload := setupIntegrationWithMock(t, []detector.ActivePort{
		{Port: 3000, PID: 1234, Process: "node"},
	})
	captureOutput(t)
	freeForce = true

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

	if err := runFree(nil, []string{"3000"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reload().GetClaim("3000") != nil {
		t.Error("expected claim to be removed with --force")
	}
}

func TestFreeCmd_NoClaim(t *testing.T) {
	setupIntegrationWithMock(t, nil)
	stdout, _ := captureOutput(t)

	if err := runFree(nil, []string{"9999"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "port 9999 has no registered claim") {
		t.Errorf("expected no-claim message, got: %q", stdout.String())
	}
}

func TestFreeCmd_NoClaim_Force(t *testing.T) {
	setupIntegrationWithMock(t, nil)
	stdout, _ := captureOutput(t)
	freeForce = true

	if err := runFree(nil, []string{"9999"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "port 9999 has no registered claim") {
		t.Errorf("expected no-claim message, got: %q", stdout.String())
	}
}
