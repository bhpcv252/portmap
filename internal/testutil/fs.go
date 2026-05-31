package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

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
