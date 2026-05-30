package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bhpcv252/portmap/internal/testutil"
)

func fixturesDir() string {
	return filepath.Join("..", "testutil", "fixtures")
}

func TestParseConfig_Valid(t *testing.T) {
	cfg, err := LoadConfig(filepath.Join(fixturesDir(), "portmap_full.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Project != "myapp" {
		t.Errorf("expected project myapp, got %q", cfg.Project)
	}
	if len(cfg.Ports) != 3 {
		t.Fatalf("expected 3 ports, got %d", len(cfg.Ports))
	}
	if cfg.Ports[0].Port != 3000 || cfg.Ports[0].Service != "frontend" {
		t.Errorf("unexpected first port: %+v", cfg.Ports[0])
	}
	if cfg.Ports[1].Port != 3001 || cfg.Ports[1].Service != "api" {
		t.Errorf("unexpected second port: %+v", cfg.Ports[1])
	}
	if cfg.Ports[2].Port != 5432 || cfg.Ports[2].Service != "postgres" {
		t.Errorf("unexpected third port: %+v", cfg.Ports[2])
	}
}

func TestParseConfig_MissingDescription(t *testing.T) {
	path := filepath.Join(t.TempDir(), "portmap.toml")
	testutil.WriteFile(t, path, `project = "myapp"

[[ports]]
port = 3000
service = "frontend"
`)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Ports[0].Description != "" {
		t.Errorf("expected empty description, got %q", cfg.Ports[0].Description)
	}
}

func TestParseConfig_InvalidTOML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "portmap.toml")
	testutil.WriteFile(t, path, "{not toml syntax")

	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid TOML")
	}
}

func TestParseConfig_Empty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "portmap.toml")
	testutil.WriteFile(t, path, "")

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Project != "" {
		t.Errorf("expected empty project, got %q", cfg.Project)
	}
	if len(cfg.Ports) != 0 {
		t.Errorf("expected 0 ports, got %d", len(cfg.Ports))
	}
}

func TestParseConfig_ExtraFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "portmap.toml")
	testutil.WriteFile(t, path, `project = "myapp"
unknown_field = "ignored"
`)

	_, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("expected no error for extra fields, got: %v", err)
	}
}

func TestWriteConfig_CreatesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "portmap.toml")
	cfg := &Config{
		Project: "myapp",
		Ports: []PortEntry{
			{Port: 3000, Service: "frontend", Description: "Next.js dev server"},
			{Port: 3001, Service: "api", Description: "Express REST API"},
		},
	}

	if err := WriteConfig(path, cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("expected file to exist after write")
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Project != "myapp" {
		t.Errorf("expected project myapp, got %q", loaded.Project)
	}
	if len(loaded.Ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(loaded.Ports))
	}
}

func TestWriteConfig_OverwritesExisting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "portmap.toml")
	testutil.WriteFile(t, path, `project = "old"`)

	cfg := &Config{Project: "new", Ports: []PortEntry{{Port: 9000, Service: "svc"}}}
	if err := WriteConfig(path, cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Project != "new" {
		t.Errorf("expected project overwritten to new, got %q", loaded.Project)
	}
}

func TestWriteConfig_Roundtrip(t *testing.T) {
	original, err := LoadConfig(filepath.Join(fixturesDir(), "portmap_full.toml"))
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "portmap.toml")
	if err := WriteConfig(path, original); err != nil {
		t.Fatal(err)
	}

	reloaded, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}

	if original.Project != reloaded.Project {
		t.Errorf("project mismatch: %q vs %q", original.Project, reloaded.Project)
	}
	if len(original.Ports) != len(reloaded.Ports) {
		t.Fatalf("port count mismatch: %d vs %d", len(original.Ports), len(reloaded.Ports))
	}
	for i, op := range original.Ports {
		rp := reloaded.Ports[i]
		if op.Port != rp.Port || op.Service != rp.Service || op.Description != rp.Description {
			t.Errorf("port %d mismatch: %+v vs %+v", i, op, rp)
		}
	}
}
