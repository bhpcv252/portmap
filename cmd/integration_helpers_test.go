//go:build integration

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/bhpcv252/portmap/internal/registry"
)

func setupIntegration(t *testing.T) (string, func() *registry.Registry) {
	t.Helper()

	regPath := filepath.Join(t.TempDir(), "registry.json")
	registryPathOverride = regPath
	t.Cleanup(func() { registryPathOverride = "" })

	resetCmdState(t)

	reload := func() *registry.Registry {
		t.Helper()
		r, err := registry.Load(regPath)
		if err != nil {
			t.Fatalf("reload registry: %v", err)
		}
		return r
	}
	return regPath, reload
}

func captureOutput(t *testing.T) (stdout, stderr *bytes.Buffer) {
	t.Helper()
	stdout = &bytes.Buffer{}
	stderr = &bytes.Buffer{}
	out = stdout
	errOut = stderr
	t.Cleanup(func() {
		out = os.Stdout
		errOut = os.Stderr
	})
	return
}

func resetCmdState(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		claimProject = ""
		claimService = ""
		claimDesc = ""
		claimForce = false
		freeForce = false
		lsProject = ""
		lsActive = false
		lsFree = false
		lsUnclaimed = false
		lsJSON = false
		noColor = false
		cwdOverride = ""
		out = os.Stdout
		errOut = os.Stderr
	})
}

func seedRegistry(t *testing.T, regPath string, claims map[string]registry.Claim) {
	t.Helper()
	r, err := registry.Load(regPath)
	if err != nil {
		t.Fatalf("seedRegistry load: %v", err)
	}
	for port, c := range claims {
		r.Set(port, c)
	}
	if err := r.Save(regPath); err != nil {
		t.Fatalf("seedRegistry save: %v", err)
	}
}
