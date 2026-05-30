//go:build integration

package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/bhpcv252/portmap/internal/registry"
)

func TestFreeCmd_ClaimedNotRunning(t *testing.T) {
	regPath, reload := setupIntegration(t)
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

func TestFreeCmd_NoClaim(t *testing.T) {
	setupIntegration(t)
	stdout, _ := captureOutput(t)

	if err := runFree(nil, []string{"9999"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "port 9999 has no registered claim") {
		t.Errorf("expected no-claim message, got: %q", stdout.String())
	}
}

func TestFreeCmd_NoClaim_Force(t *testing.T) {
	setupIntegration(t)
	stdout, _ := captureOutput(t)

	freeForce = true
	if err := runFree(nil, []string{"9999"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "port 9999 has no registered claim") {
		t.Errorf("expected no-claim message, got: %q", stdout.String())
	}
}
