package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bhpcv252/portmap/internal/testutil"
)

func fixturesDir() string {
	return filepath.Join("..", "testutil", "fixtures")
}

func TestLoad_ValidFile(t *testing.T) {
	r, err := Load(filepath.Join(fixturesDir(), "registry_full.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Claims) != 4 {
		t.Fatalf("expected 4 claims, got %d", len(r.Claims))
	}
	if r.Version != 1 {
		t.Fatalf("expected version 1, got %d", r.Version)
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does_not_exist.json")
	r, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Claims) != 0 {
		t.Fatalf("expected 0 claims, got %d", len(r.Claims))
	}
	if r.Version != 1 {
		t.Fatalf("expected version 1, got %d", r.Version)
	}
}

func TestLoad_CorruptJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "corrupt.json")
	testutil.WriteFile(t, path, "{not valid json")

	r, err := Load(path)
	if err == nil {
		t.Fatal("expected error for corrupt JSON")
	}
	if r != nil {
		t.Fatal("expected nil registry on error")
	}
}

func TestLoad_EmptyClaims(t *testing.T) {
	r, err := Load(filepath.Join(fixturesDir(), "registry_empty.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Claims) != 0 {
		t.Fatalf("expected 0 claims, got %d", len(r.Claims))
	}
}

func TestLoad_VersionPreserved(t *testing.T) {
	r, err := Load(filepath.Join(fixturesDir(), "registry_full.json"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Version != 1 {
		t.Fatalf("expected version 1, got %d", r.Version)
	}
}

func TestSave_WritesJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "registry.json")
	r := &Registry{Version: 1, Claims: make(map[string]Claim)}
	r.Set("3000", Claim{Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()})
	r.Set("3001", Claim{Project: "myapp", Service: "api", ClaimedAt: time.Now().UTC()})

	if err := r.Save(path); err != nil {
		t.Fatal(err)
	}

	data := testutil.MustReadFile(t, path)
	var parsed Registry
	if err := json.Unmarshal([]byte(data), &parsed); err != nil {
		t.Fatalf("saved file is not valid JSON: %v", err)
	}
	if _, ok := parsed.Claims["3000"]; !ok {
		t.Error("expected claim for port 3000")
	}
	if _, ok := parsed.Claims["3001"]; !ok {
		t.Error("expected claim for port 3001")
	}
}

func TestSave_CreatesParentDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "dir", "registry.json")
	r := &Registry{Version: 1, Claims: make(map[string]Claim)}

	if err := r.Save(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("expected file to exist after save")
	}
}

func TestSave_Roundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "registry.json")
	r, _ := Load(path)
	r.Set("3000", Claim{
		Project:     "myapp",
		Service:     "frontend",
		Description: "Next.js dev server",
		ClaimedAt:   time.Now().UTC().Truncate(time.Second),
		Path:        "/home/user/projects/myapp",
	})

	if err := r.Save(path); err != nil {
		t.Fatal(err)
	}

	r2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	c, ok := r2.Get("3000")
	if !ok {
		t.Fatal("expected claim for port 3000 after reload")
	}
	if c.Project != "myapp" || c.Service != "frontend" || c.Description != "Next.js dev server" {
		t.Errorf("claim fields mismatch after roundtrip: %+v", c)
	}
}

func TestSave_PreservesVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "registry.json")
	r := &Registry{Version: 1, Claims: make(map[string]Claim)}

	if err := r.Save(path); err != nil {
		t.Fatal(err)
	}

	data := testutil.MustReadFile(t, path)
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(data), &parsed); err != nil {
		t.Fatal(err)
	}
	v, ok := parsed["version"]
	if !ok {
		t.Fatal("expected version field in saved JSON")
	}
	if v.(float64) != 1 {
		t.Fatalf("expected version 1, got %v", v)
	}
}
