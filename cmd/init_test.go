//go:build integration

package cmd

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhpcv252/portmap/internal/project"
	"github.com/bhpcv252/portmap/internal/testutil"
)

func TestInitCmd_TwoPorts(t *testing.T) {
	setupIntegration(t)
	stdout, _ := captureOutput(t)

	dir := t.TempDir()
	cwdOverride = dir
	injectStdin(t, "myapp\n3000\nfrontend\nNext.js dev server\n3001\napi\nExpress backend\n\n")

	if err := runInit(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := project.LoadConfig(filepath.Join(dir, "portmap.toml"))
	if err != nil {
		t.Fatalf("parse portmap.toml: %v", err)
	}
	if cfg.Project != "myapp" {
		t.Errorf("expected project myapp, got %q", cfg.Project)
	}
	if len(cfg.Ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(cfg.Ports))
	}
	if cfg.Ports[0].Port != 3000 || cfg.Ports[0].Service != "frontend" {
		t.Errorf("unexpected first port: %+v", cfg.Ports[0])
	}
	if cfg.Ports[1].Port != 3001 || cfg.Ports[1].Service != "api" {
		t.Errorf("unexpected second port: %+v", cfg.Ports[1])
	}
	_ = stdout
}

func TestInitCmd_NoPorts(t *testing.T) {
	setupIntegration(t)
	captureOutput(t)

	dir := t.TempDir()
	cwdOverride = dir
	injectStdin(t, "myapp\n\n")

	if err := runInit(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := project.LoadConfig(filepath.Join(dir, "portmap.toml"))
	if err != nil {
		t.Fatalf("parse portmap.toml: %v", err)
	}
	if cfg.Project != "myapp" {
		t.Errorf("expected project myapp, got %q", cfg.Project)
	}
	if len(cfg.Ports) != 0 {
		t.Errorf("expected 0 ports, got %d", len(cfg.Ports))
	}
}

func TestInitCmd_PrintsSyncHint(t *testing.T) {
	setupIntegration(t)
	stdout, _ := captureOutput(t)

	dir := t.TempDir()
	cwdOverride = dir
	injectStdin(t, "myapp\n\n")

	if err := runInit(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "portmap sync") {
		t.Errorf("expected sync hint in output, got: %q", stdout.String())
	}
}

func TestInitCmd_InfersProjectFromFolder(t *testing.T) {
	setupIntegration(t)
	captureOutput(t)

	dir := filepath.Join(t.TempDir(), "myproject")
	testutil.WriteFile(t, filepath.Join(dir, ".keep"), "")
	cwdOverride = dir
	// empty project name -> falls back to folder name
	injectStdin(t, "\n\n")

	if err := runInit(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := project.LoadConfig(filepath.Join(dir, "portmap.toml"))
	if err != nil {
		t.Fatalf("parse portmap.toml: %v", err)
	}
	if cfg.Project != "myproject" {
		t.Errorf("expected project inferred as 'myproject', got %q", cfg.Project)
	}
}

func TestInitCmd_TomlAlreadyExists_Overwrite(t *testing.T) {
	setupIntegration(t)
	captureOutput(t)

	dir := t.TempDir()
	cwdOverride = dir
	testutil.WriteFile(t, filepath.Join(dir, "portmap.toml"), `project = "old"`)

	// project name, no ports, then "y" to overwrite
	injectStdin(t, "newapp\n\ny\n")

	if err := runInit(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := project.LoadConfig(filepath.Join(dir, "portmap.toml"))
	if err != nil {
		t.Fatalf("parse portmap.toml: %v", err)
	}
	if cfg.Project != "newapp" {
		t.Errorf("expected project overwritten to 'newapp', got %q", cfg.Project)
	}
}

func TestInitCmd_TomlAlreadyExists_Decline(t *testing.T) {
	setupIntegration(t)
	captureOutput(t)

	dir := t.TempDir()
	cwdOverride = dir
	testutil.WriteFile(t, filepath.Join(dir, "portmap.toml"), `project = "old"`)

	// project name, no ports, then "n" to decline
	injectStdin(t, "newapp\n\nn\n")

	if err := runInit(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := project.LoadConfig(filepath.Join(dir, "portmap.toml"))
	if err != nil {
		t.Fatalf("parse portmap.toml: %v", err)
	}
	if cfg.Project != "old" {
		t.Errorf("expected original project 'old' preserved, got %q", cfg.Project)
	}
}

func TestInitCmd_InvalidPortSkipped(t *testing.T) {
	setupIntegration(t)
	stdout, _ := captureOutput(t)

	dir := t.TempDir()
	cwdOverride = dir
	// port 99999 is invalid, then a valid port, then done
	injectStdin(t, "myapp\n99999\n3000\napi\ndesc\n\n")

	if err := runInit(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := project.LoadConfig(filepath.Join(dir, "portmap.toml"))
	if err != nil {
		t.Fatalf("parse portmap.toml: %v", err)
	}
	if len(cfg.Ports) != 1 || cfg.Ports[0].Port != 3000 {
		t.Errorf("expected only port 3000, got: %+v", cfg.Ports)
	}
	if !strings.Contains(stdout.String(), "invalid port") {
		t.Errorf("expected invalid port warning, got: %q", stdout.String())
	}
}
