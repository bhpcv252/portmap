package cmdtest

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/bhpcv252/portmap/internal/registry"
)

func TempRegistry(t *testing.T) (string, *registry.Registry) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "registry.json")
	r, err := registry.Load(path)
	if err != nil {
		t.Fatalf("TempRegistry: %v", err)
	}
	return path, r
}

func RegistryWithClaims(t *testing.T, claims map[string]registry.Claim) *registry.Registry {
	t.Helper()
	path, r := TempRegistry(t)
	for port, c := range claims {
		if c.ClaimedAt.IsZero() {
			c.ClaimedAt = time.Now().UTC()
		}
		r.Set(port, c)
	}
	if err := r.Save(path); err != nil {
		t.Fatalf("RegistryWithClaims: save: %v", err)
	}
	return r
}
