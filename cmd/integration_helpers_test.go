//go:build integration

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bhpcv252/portmap/internal/detector"
	"github.com/bhpcv252/portmap/internal/registry"
)

type testMock struct {
	ports []detector.ActivePort
	err   error
}

func (m *testMock) ActivePorts() ([]detector.ActivePort, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.ports, nil
}

func (m *testMock) IsActive(port int) (bool, *detector.ActivePort, error) {
	if m.err != nil {
		return false, nil, m.err
	}
	for _, ap := range m.ports {
		if ap.Port == port {
			cp := ap
			return true, &cp, nil
		}
	}
	return false, nil, nil
}

func setupIntegration(t *testing.T) (string, func() *registry.Registry) {
	t.Helper()

	regPath := filepath.Join(t.TempDir(), "registry.json")
	registryPathOverride = regPath
	det = nil

	t.Cleanup(func() {
		registryPathOverride = ""
		det = nil
		resetCmdFlags()
		out = os.Stdout
		errOut = os.Stderr
		in = os.Stdin
		cwdOverride = ""
	})

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

func setupIntegrationWithMock(
	t *testing.T,
	active []detector.ActivePort,
) (string, func() *registry.Registry) {
	t.Helper()
	regPath, reload := setupIntegration(t)
	det = &testMock{ports: active}
	return regPath, reload
}

func captureOutput(t *testing.T) (stdout, stderr *bytes.Buffer) {
	t.Helper()
	stdout = &bytes.Buffer{}
	stderr = &bytes.Buffer{}
	out = stdout
	errOut = stderr
	return
}

func injectStdin(t *testing.T, s string) {
	t.Helper()
	in = strings.NewReader(s)
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

func resetCmdFlags() {
	claimProject, claimService, claimDesc = "", "", ""
	claimForce = false
	freeForce = false
	lsProject = ""
	lsActive, lsFree, lsUnclaimed, lsJSON = false, false, false, false
	noColor = false
	killYes = false
	suggestFrom, suggestTo, suggestCount = 3000, 9999, 1
}
