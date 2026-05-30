//go:build integration

package cmd

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bhpcv252/portmap/internal/project"
	"github.com/bhpcv252/portmap/internal/registry"
	"github.com/bhpcv252/portmap/internal/testutil"
)

func TestExportCmd_WritesFile(t *testing.T) {
	regPath, _ := setupIntegration(t)
	captureOutput(t)

	dir := t.TempDir()
	outPath := filepath.Join(dir, "portmap.toml")
	exportProject = "myapp"
	exportOutput = outPath

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {
			Project:     "myapp",
			Service:     "frontend",
			Description: "Next.js dev server",
			ClaimedAt:   time.Now().UTC(),
		},
		"3001": {
			Project:     "myapp",
			Service:     "api",
			Description: "Express REST API",
			ClaimedAt:   time.Now().UTC(),
		},
	})

	if err := runExport(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := project.LoadConfig(outPath)
	if err != nil {
		t.Fatalf("parse exported file: %v", err)
	}
	if cfg.Project != "myapp" {
		t.Errorf("expected project myapp, got %q", cfg.Project)
	}
	if len(cfg.Ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(cfg.Ports))
	}
}

func TestExportCmd_OutputFlag(t *testing.T) {
	regPath, _ := setupIntegration(t)
	captureOutput(t)

	dir := t.TempDir()
	outPath := filepath.Join(dir, "custom", "output.toml")
	testutil.WriteFile(t, filepath.Join(dir, "custom", ".keep"), "") // create parent dir
	exportProject = "myapp"
	exportOutput = outPath

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

	if err := runExport(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := project.LoadConfig(outPath)
	if err != nil {
		t.Fatalf("expected file at custom output path: %v", err)
	}
	if len(cfg.Ports) != 1 {
		t.Errorf("expected 1 port, got %d", len(cfg.Ports))
	}
}

func TestExportCmd_Stdout(t *testing.T) {
	regPath, _ := setupIntegration(t)
	stdout, _ := captureOutput(t)
	exportProject = "myapp"
	exportStdout = true

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {
			Project:     "myapp",
			Service:     "frontend",
			Description: "Next.js dev server",
			ClaimedAt:   time.Now().UTC(),
		},
	})

	if err := runExport(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "myapp") {
		t.Errorf("expected project name in stdout, got: %q", got)
	}
	if !strings.Contains(got, "3000") {
		t.Errorf("expected port 3000 in stdout, got: %q", got)
	}
	if !strings.Contains(got, "frontend") {
		t.Errorf("expected service name in stdout, got: %q", got)
	}
}

func TestExportCmd_NoClaimsForProject(t *testing.T) {
	regPath, _ := setupIntegration(t)
	stdout, _ := captureOutput(t)
	exportProject = "unknownproject"

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
	})

	if err := runExport(nil, nil); err != nil {
		t.Fatalf("expected clean exit for unknown project, got: %v", err)
	}
	if !strings.Contains(stdout.String(), "no claims found") {
		t.Errorf("expected 'no claims found' message, got: %q", stdout.String())
	}
}

func TestExportCmd_RoundTrip(t *testing.T) {
	_, _ = setupIntegration(t)
	captureOutput(t)

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
`)
	cwdOverride = dir

	if err := runSync(nil, nil); err != nil {
		t.Fatalf("sync: %v", err)
	}

	syncDryRun, syncForce = false, false

	stdout, _ := captureOutput(t)
	exportProject = "myapp"
	exportStdout = true

	if err := runExport(nil, nil); err != nil {
		t.Fatalf("export: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{"myapp", "3000", "frontend", "3001", "api"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in exported output, got:\n%s", want, got)
		}
	}
}
