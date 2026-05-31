//go:build integration

package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/bhpcv252/portmap/internal/detector"
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

func setupMock(t *testing.T, active []detector.ActivePort) {
	t.Helper()
	det = &testMock{ports: active}
	t.Cleanup(func() {
		det = nil
		resetCmdFlags()
		out = os.Stdout
		errOut = os.Stderr
		in = os.Stdin
	})
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

func resetCmdFlags() {
	lsJSON = false
	noColor = false
	killYes = false
	suggestFrom, suggestTo, suggestCount = 3000, 9999, 1
}
