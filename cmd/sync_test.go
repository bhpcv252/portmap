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

func TestSyncCmd_Fresh(t *testing.T) {
	_, reload := setupIntegration(t)
	stdout, _ := captureOutput(t)

	dir := t.TempDir()
	testutil.WriteFile(t, filepath.Join(dir, "portmap.toml"), `project = "myapp"

[[ports]]
port = 3000
service = "frontend"
description = "Next.js dev server"

[[ports]]
port = 3001
service = "api"
description = "Express REST API"

[[ports]]
port = 5432
service = "postgres"
description = "Local PostgreSQL"
`)
	cwdOverride = dir

	if err := runSync(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := reload()
	for _, port := range []string{"3000", "3001", "5432"} {
		if r.GetClaim(port) == nil {
			t.Errorf("expected claim for port %s", port)
		}
	}
	if !strings.Contains(stdout.String(), "synced 3") {
		t.Errorf("expected 'synced 3' in output, got: %q", stdout.String())
	}
}

func TestSyncCmd_AlreadyClaimedSameProject(t *testing.T) {
	regPath, reload := setupIntegration(t)
	stdout, _ := captureOutput(t)

	dir := t.TempDir()
	testutil.WriteFile(t, filepath.Join(dir, "portmap.toml"), `project = "myapp"

[[ports]]
port = 3000
service = "frontend"
`)
	cwdOverride = dir

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC(), Path: dir},
	})

	if err := runSync(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := reload().GetClaim("3000")
	if c == nil || c.Project != "myapp" {
		t.Error("expected claim unchanged for same project")
	}
	if !strings.Contains(stdout.String(), "skipping") {
		t.Errorf("expected 'skipping' in output, got: %q", stdout.String())
	}
}

func TestSyncCmd_AlreadyClaimedDiffProject_NoForce(t *testing.T) {
	regPath, reload := setupIntegration(t)
	stdout, _ := captureOutput(t)

	dir := t.TempDir()
	testutil.WriteFile(t, filepath.Join(dir, "portmap.toml"), `project = "myapp"

[[ports]]
port = 3000
service = "frontend"
`)
	cwdOverride = dir

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "otherapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

	if err := runSync(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := reload().GetClaim("3000")
	if c == nil || c.Project != "otherapp" {
		t.Error("expected claim unchanged when conflicting without --force")
	}
	if !strings.Contains(stdout.String(), "skipping") {
		t.Errorf("expected 'skipping' in output, got: %q", stdout.String())
	}
}

func TestSyncCmd_AlreadyClaimedDiffProject_Force(t *testing.T) {
	regPath, reload := setupIntegration(t)
	captureOutput(t)
	syncForce = true

	dir := t.TempDir()
	testutil.WriteFile(t, filepath.Join(dir, "portmap.toml"), `project = "myapp"

[[ports]]
port = 3000
service = "frontend"
`)
	cwdOverride = dir

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "otherapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

	if err := runSync(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c := reload().GetClaim("3000")
	if c == nil || c.Project != "myapp" {
		t.Errorf("expected claim overwritten with --force, got: %+v", c)
	}
}

func TestSyncCmd_DryRun(t *testing.T) {
	_, reload := setupIntegration(t)
	stdout, _ := captureOutput(t)
	syncDryRun = true

	dir := t.TempDir()
	testutil.WriteFile(t, filepath.Join(dir, "portmap.toml"), `project = "myapp"

[[ports]]
port = 3000
service = "frontend"
`)
	cwdOverride = dir

	if err := runSync(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reload().GetClaim("3000") != nil {
		t.Error("expected registry unchanged with --dry-run")
	}
	if !strings.Contains(stdout.String(), "would claim") {
		t.Errorf("expected 'would claim' in output, got: %q", stdout.String())
	}
}

func TestSyncCmd_NoToml(t *testing.T) {
	setupIntegration(t)
	_, stderr := captureOutput(t)

	cwdOverride = t.TempDir()

	err := runSync(nil, nil)
	if err == nil {
		t.Fatal("expected error when no portmap.toml found")
	}
	if !strings.Contains(stderr.String(), "error:") {
		t.Errorf("expected 'error:' on stderr, got: %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "portmap init") {
		t.Errorf("expected 'portmap init' hint, got: %q", stderr.String())
	}
}

func TestSyncCmd_PartialSync(t *testing.T) {
	regPath, reload := setupIntegration(t)
	stdout, _ := captureOutput(t)

	dir := t.TempDir()
	testutil.WriteFile(t, filepath.Join(dir, "portmap.toml"), `project = "myapp"

[[ports]]
port = 3000
service = "frontend"

[[ports]]
port = 3001
service = "api"

[[ports]]
port = 3002
service = "worker"
`)
	cwdOverride = dir

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3002": {Project: "otherapp", Service: "worker", ClaimedAt: time.Now().UTC()},
	})

	if err := runSync(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "synced 2") {
		t.Errorf("expected 'synced 2' in output, got: %q", got)
	}
	if !strings.Contains(got, "1 skipped") {
		t.Errorf("expected '1 skipped' in output, got: %q", got)
	}

	r := reload()
	if r.GetClaim("3000") == nil {
		t.Error("expected port 3000 to be claimed")
	}
	if r.GetClaim("3001") == nil {
		t.Error("expected port 3001 to be claimed")
	}
	if c := r.GetClaim("3002"); c == nil || c.Project != "otherapp" {
		t.Error("expected port 3002 to remain with otherapp")
	}
}
