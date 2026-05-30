//go:build integration

package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/bhpcv252/portmap/internal/registry"
)

func TestCleanCmd_NoneFound(t *testing.T) {
	regPath, _ := setupIntegration(t)
	stdout, _ := captureOutput(t)

	existingDir := t.TempDir()
	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {
			Project:   "myapp",
			Service:   "frontend",
			ClaimedAt: time.Now().UTC(),
			Path:      existingDir,
		},
	})

	if err := runClean(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "no stale claims found") {
		t.Errorf("expected 'no stale claims found', got: %q", stdout.String())
	}
}

func TestCleanCmd_StaleFound(t *testing.T) {
	regPath, _ := setupIntegration(t)
	stdout, _ := captureOutput(t)
	cleanYes = true

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {
			Project:   "myapp",
			Service:   "frontend",
			ClaimedAt: time.Now().UTC(),
			Path:      "/nonexistent/path/xyz",
		},
	})

	if err := runClean(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "3000") {
		t.Errorf("expected port 3000 listed as stale, got: %q", stdout.String())
	}
}

func TestCleanCmd_NotStale(t *testing.T) {
	regPath, _ := setupIntegration(t)
	stdout, _ := captureOutput(t)

	existingDir := t.TempDir()
	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {
			Project:   "myapp",
			Service:   "frontend",
			ClaimedAt: time.Now().UTC(),
			Path:      existingDir,
		},
	})

	if err := runClean(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "no stale claims found") {
		t.Errorf("expected no stale claims for existing path, got: %q", stdout.String())
	}
}

func TestCleanCmd_EmptyPath(t *testing.T) {
	regPath, _ := setupIntegration(t)
	stdout, _ := captureOutput(t)

	seedRegistry(t, regPath, map[string]registry.Claim{
		// empty path is skipped (e.g. shared/postgres with no project dir)
		"5432": {Project: "shared", Service: "postgres", ClaimedAt: time.Now().UTC(), Path: ""},
	})

	if err := runClean(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "no stale claims found") {
		t.Errorf("expected empty path to be skipped, got: %q", stdout.String())
	}
}

func TestCleanCmd_DryRun(t *testing.T) {
	regPath, reload := setupIntegration(t)
	stdout, _ := captureOutput(t)
	cleanDryRun = true

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {
			Project:   "myapp",
			Service:   "frontend",
			ClaimedAt: time.Now().UTC(),
			Path:      "/nonexistent/path/xyz",
		},
	})

	if err := runClean(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reload().GetClaim("3000") == nil {
		t.Error("expected registry unchanged with --dry-run")
	}
	if !strings.Contains(stdout.String(), "3000") {
		t.Errorf("expected stale claim listed in dry-run output, got: %q", stdout.String())
	}
}

func TestCleanCmd_Confirm_Yes(t *testing.T) {
	regPath, reload := setupIntegration(t)
	captureOutput(t)
	injectStdin(t, "y\n")

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {
			Project:   "myapp",
			Service:   "frontend",
			ClaimedAt: time.Now().UTC(),
			Path:      "/nonexistent/path/xyz",
		},
	})

	if err := runClean(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reload().GetClaim("3000") != nil {
		t.Error("expected claim removed after 'y' confirmation")
	}
}

func TestCleanCmd_Confirm_No(t *testing.T) {
	regPath, reload := setupIntegration(t)
	captureOutput(t)
	injectStdin(t, "n\n")

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {
			Project:   "myapp",
			Service:   "frontend",
			ClaimedAt: time.Now().UTC(),
			Path:      "/nonexistent/path/xyz",
		},
	})

	if err := runClean(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reload().GetClaim("3000") == nil {
		t.Error("expected claim to remain after 'n' confirmation")
	}
}

func TestCleanCmd_YesFlag(t *testing.T) {
	regPath, reload := setupIntegration(t)
	captureOutput(t)
	cleanYes = true

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {
			Project:   "myapp",
			Service:   "frontend",
			ClaimedAt: time.Now().UTC(),
			Path:      "/nonexistent/path/xyz",
		},
	})

	if err := runClean(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reload().GetClaim("3000") != nil {
		t.Error("expected claim removed with -y flag")
	}
}

func TestCleanCmd_MultipleStale(t *testing.T) {
	regPath, reload := setupIntegration(t)
	stdout, _ := captureOutput(t)
	cleanYes = true

	existingDir := t.TempDir()
	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {
			Project:   "myapp",
			Service:   "frontend",
			ClaimedAt: time.Now().UTC(),
			Path:      "/nonexistent/a",
		},
		"3001": {
			Project:   "myapp",
			Service:   "api",
			ClaimedAt: time.Now().UTC(),
			Path:      "/nonexistent/b",
		},
		"3002": {
			Project:   "myapp",
			Service:   "worker",
			ClaimedAt: time.Now().UTC(),
			Path:      existingDir,
		},
	})

	if err := runClean(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reload().GetClaim("3000") != nil {
		t.Error("expected stale port 3000 removed")
	}
	if reload().GetClaim("3001") != nil {
		t.Error("expected stale port 3001 removed")
	}
	if reload().GetClaim("3002") == nil {
		t.Error("expected port 3002 with existing path to remain")
	}
	if !strings.Contains(stdout.String(), "removed 2") {
		t.Errorf("expected 'removed 2' in output, got: %q", stdout.String())
	}
}
