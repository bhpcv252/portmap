//go:build integration

package cmd

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bhpcv252/portmap/internal/registry"
	"github.com/bhpcv252/portmap/internal/testutil"
)

func TestClaimCmd_NewPort(t *testing.T) {
	_, reload := setupIntegration(t)
	stdout, _ := captureOutput(t)

	claimProject = "myapp"
	claimService = "frontend"
	claimDesc = "Next.js dev server"

	if err := runClaim(nil, []string{"3000"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := reload()
	c := r.GetClaim("3000")
	if c == nil {
		t.Fatal("expected claim for port 3000")
	}
	if c.Project != "myapp" || c.Service != "frontend" || c.Description != "Next.js dev server" {
		t.Errorf("unexpected claim fields: %+v", c)
	}
	if !strings.Contains(stdout.String(), "claimed port 3000") {
		t.Errorf("expected success message, got: %q", stdout.String())
	}
}

func TestClaimCmd_UpdatesDescription(t *testing.T) {
	regPath, reload := setupIntegration(t)
	stdout, _ := captureOutput(t)

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {
			Project:     "myapp",
			Service:     "frontend",
			Description: "old",
			ClaimedAt:   time.Now().UTC(),
		},
	})

	claimProject = "myapp"
	claimService = "frontend"
	claimDesc = "new desc"

	if err := runClaim(nil, []string{"3000"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := reload().GetClaim("3000")
	if c == nil {
		t.Fatal("expected claim")
	}
	if c.Description != "new desc" {
		t.Errorf("expected description updated, got %q", c.Description)
	}
	if !strings.Contains(stdout.String(), "updated port 3000") {
		t.Errorf("expected updated message, got: %q", stdout.String())
	}
}

func TestClaimCmd_SameServiceNoDesc(t *testing.T) {
	regPath, _ := setupIntegration(t)

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

	claimProject = "myapp"
	claimService = "frontend"

	if err := runClaim(nil, []string{"3000"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClaimCmd_ConflictNoForce(t *testing.T) {
	regPath, reload := setupIntegration(t)
	_, stderr := captureOutput(t)

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

	claimProject = "otherapp"
	claimService = "frontend"

	err := runClaim(nil, []string{"3000"})
	if err == nil {
		t.Fatal("expected error for conflict")
	}

	// registry must be unchanged
	c := reload().GetClaim("3000")
	if c == nil || c.Project != "myapp" {
		t.Error("expected registry unchanged after conflict")
	}
	if !strings.Contains(stderr.String(), "error:") {
		t.Errorf("expected error message on stderr, got: %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "--force") {
		t.Errorf("expected --force hint on stderr, got: %q", stderr.String())
	}
}

func TestClaimCmd_ConflictWithForce(t *testing.T) {
	regPath, reload := setupIntegration(t)

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

	claimProject = "otherapp"
	claimService = "frontend"
	claimForce = true

	if err := runClaim(nil, []string{"3000"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := reload().GetClaim("3000")
	if c == nil || c.Project != "otherapp" {
		t.Errorf("expected claim overwritten, got: %+v", c)
	}
}

func TestClaimCmd_InfersProjectFromToml(t *testing.T) {
	_, reload := setupIntegration(t)

	dir := t.TempDir()
	testutil.WriteFile(t, filepath.Join(dir, "portmap.toml"), `project = "tomlapp"`)
	cwdOverride = dir

	claimService = "api"

	if err := runClaim(nil, []string{"3000"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := reload().GetClaim("3000")
	if c == nil || c.Project != "tomlapp" {
		t.Errorf("expected project inferred from toml, got: %+v", c)
	}
}

func TestClaimCmd_InfersProjectFromFolder(t *testing.T) {
	_, reload := setupIntegration(t)

	dir := filepath.Join(t.TempDir(), "myproject")
	testutil.WriteFile(t, filepath.Join(dir, ".keep"), "")
	cwdOverride = dir

	claimService = "api"

	if err := runClaim(nil, []string{"3000"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := reload().GetClaim("3000")
	if c == nil || c.Project != "myproject" {
		t.Errorf("expected project inferred from folder name, got: %+v", c)
	}
}

func TestClaimCmd_StoresPath(t *testing.T) {
	_, reload := setupIntegration(t)

	dir := t.TempDir()
	cwdOverride = dir
	claimProject = "myapp"
	claimService = "api"

	if err := runClaim(nil, []string{"3000"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := reload().GetClaim("3000")
	if c == nil || c.Path != dir {
		t.Errorf("expected path %q, got %+v", dir, c)
	}
}

func TestClaimCmd_StoresClaimedAt(t *testing.T) {
	_, reload := setupIntegration(t)

	claimProject = "myapp"
	claimService = "api"

	if err := runClaim(nil, []string{"3000"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := reload().GetClaim("3000")
	if c == nil || c.ClaimedAt.IsZero() {
		t.Error("expected claimed_at to be set")
	}
}

func TestClaimCmd_InvalidPort_Zero(t *testing.T) {
	regPath, reload := setupIntegration(t)

	seedRegistry(t, regPath, map[string]registry.Claim{})

	err := runClaim(nil, []string{"0"})
	if err == nil {
		t.Fatal("expected error for port 0")
	}
	if len(reload().Claims) != 0 {
		t.Error("expected registry unchanged")
	}
}

func TestClaimCmd_InvalidPort_TooHigh(t *testing.T) {
	_, reload := setupIntegration(t)

	err := runClaim(nil, []string{"65536"})
	if err == nil {
		t.Fatal("expected error for port 65536")
	}
	if len(reload().Claims) != 0 {
		t.Error("expected registry unchanged")
	}
}
