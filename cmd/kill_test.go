//go:build integration

package cmd

import (
	"net"
	"strings"
	"testing"
)

func bindPort(t *testing.T) (net.Listener, int) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("bindPort: %v", err)
	}
	return l, l.Addr().(*net.TCPAddr).Port
}

func TestKillCmd_NotRunning(t *testing.T) {
	setupIntegrationWithMock(t, nil) // no active ports
	_, stderr := captureOutput(t)

	err := runKill(nil, []string{"3000"})
	if err == nil {
		t.Fatal("expected error when nothing running")
	}
	if !strings.Contains(stderr.String(), "nothing is running on port 3000") {
		t.Errorf("expected error message, got: %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "nothing to kill") {
		t.Errorf("expected hint, got: %q", stderr.String())
	}
}
