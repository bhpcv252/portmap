//go:build integration

package cmd

import (
	"strings"
	"testing"

	"github.com/bhpcv252/portmap/internal/detector"
)

func TestKillCmd_NotRunning(t *testing.T) {
	setupMock(t, nil)
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

func TestKillCmd_ShowsPortInfo(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 3000, PID: 84312, Process: "node", CWD: "/home/user/myapp"},
	})
	stdout, _ := captureOutput(t)
	injectStdin(t, "n\n")
	noColor = true

	if err := runKill(nil, []string{"3000"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{"● running", "84312", "node", "/home/user/myapp"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

func TestKillCmd_Confirm_No(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 3000, PID: 84312, Process: "node"},
	})
	captureOutput(t)
	injectStdin(t, "n\n")

	if err := runKill(nil, []string{"3000"}); err != nil {
		t.Fatalf("expected clean exit on 'n', got: %v", err)
	}
}

func TestKillCmd_InvalidPort(t *testing.T) {
	setupMock(t, nil)
	captureOutput(t)

	if err := runKill(nil, []string{"notaport"}); err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestKillCmd_ShowsProcessInfo(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 3000, PID: 12345, Process: "node"},
	})
	stdout, _ := captureOutput(t)
	injectStdin(t, "n\n")

	if err := runKill(nil, []string{"3000"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "pid:") {
		t.Errorf("expected 'pid:' in output, got: %q", got)
	}
	if !strings.Contains(got, "12345") {
		t.Errorf("expected PID 12345 in output, got: %q", got)
	}
}

func TestKillCmd_YesFlag_SkipsConfirmation(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 3000, PID: 99999, Process: "node"},
	})
	stdout, _ := captureOutput(t)
	killYes = true

	// PID 99999 almost certainly doesn't exist so KillProcess returns an error
	// we don't care about that error; we only verify the [y/N] prompt was never shown
	runKill(nil, []string{"3000"})

	if strings.Contains(stdout.String(), "[y/N]") {
		t.Error("expected confirmation prompt to be skipped with --yes flag")
	}
}
