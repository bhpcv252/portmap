package project

import (
	"path/filepath"
	"testing"

	"github.com/bhpcv252/portmap/internal/testutil"
)

func TestInferName_TomlInCwd(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, filepath.Join(dir, "portmap.toml"), `project = "myapp"`)

	if got := InferName(dir); got != "myapp" {
		t.Errorf("expected myapp, got %q", got)
	}
}

func TestInferName_TomlInParent(t *testing.T) {
	parent := t.TempDir()
	testutil.WriteFile(t, filepath.Join(parent, "portmap.toml"), `project = "parentapp"`)
	child := filepath.Join(parent, "child")
	testutil.WriteFile(t, filepath.Join(child, ".keep"), "")

	if got := InferName(child); got != "parentapp" {
		t.Errorf("expected parentapp, got %q", got)
	}
}

func TestInferName_TomlInGrandparent(t *testing.T) {
	root := t.TempDir()
	testutil.WriteFile(t, filepath.Join(root, "portmap.toml"), `project = "rootapp"`)
	grandchild := filepath.Join(root, "level1", "level2")
	testutil.WriteFile(t, filepath.Join(grandchild, ".keep"), "")

	if got := InferName(grandchild); got != "rootapp" {
		t.Errorf("expected rootapp, got %q", got)
	}
}

func TestInferName_NoToml_FolderName(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "myproject")
	testutil.WriteFile(t, filepath.Join(dir, ".keep"), "")

	if got := InferName(dir); got != "myproject" {
		t.Errorf("expected myproject, got %q", got)
	}
}

func TestInferName_TomlMissingProjectField(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "myproject")
	testutil.WriteFile(t, filepath.Join(dir, "portmap.toml"), `project = ""`)

	if got := InferName(dir); got != "myproject" {
		t.Errorf("expected fallback to folder name myproject, got %q", got)
	}
}

func TestFindConfigFile_InCwd(t *testing.T) {
	dir := t.TempDir()
	expected := filepath.Join(dir, "portmap.toml")
	testutil.WriteFile(t, expected, `project = "myapp"`)

	path, _, err := FindConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

func TestFindConfigFile_InParent(t *testing.T) {
	parent := t.TempDir()
	expected := filepath.Join(parent, "portmap.toml")
	testutil.WriteFile(t, expected, `project = "myapp"`)
	child := filepath.Join(parent, "child")
	testutil.WriteFile(t, filepath.Join(child, ".keep"), "")

	path, _, err := FindConfig(child)
	if err != nil {
		t.Fatal(err)
	}
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

func TestFindConfigFile_NotFound(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "notoml")
	testutil.WriteFile(t, filepath.Join(dir, ".keep"), "")

	path, cfg, err := FindConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if path != "" {
		t.Errorf("expected empty path, got %q", path)
	}
	if cfg != nil {
		t.Errorf("expected nil config, got %+v", cfg)
	}
}
