package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TempDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

func WriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("WriteFile: mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %s: %v", path, err)
	}
}

func MustReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("MustReadFile: %s: %v", path, err)
	}
	return string(data)
}

func FixturesDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "fixtures")
}
