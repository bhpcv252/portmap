package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Registry struct {
	Version int              `json:"version"`
	Claims  map[string]Claim `json:"claims"`
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".portmap", "registry.json")
}

func Load(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// first use: no registry file yet, return a clean empty one
		return &Registry{Version: 1, Claims: make(map[string]Claim)}, nil
	}
	if err != nil {
		return nil, err
	}

	var r Registry
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	// guard against a registry file that has no claims key at all
	if r.Claims == nil {
		r.Claims = make(map[string]Claim)
	}
	return &r, nil
}

func (r *Registry) Save(path string) error {
	// create ~/.portmap/ if this is the first save
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (r *Registry) Get(port string) (Claim, bool) {
	c, ok := r.Claims[port]
	return c, ok
}

func (r *Registry) Set(port string, c Claim) {
	r.Claims[port] = c
}

func (r *Registry) Delete(port string) {
	delete(r.Claims, port)
}
