//go:build integration

package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/bhpcv252/portmap/internal/detector"
)

func TestCheckCmd_Free(t *testing.T) {
	setupMock(t, nil)
	stdout, _ := captureOutput(t)

	if err := runCheck(nil, []string{"3000"}); err != nil {
		t.Fatalf("expected exit 0 for free port, got: %v", err)
	}
	if !strings.Contains(stdout.String(), "not running") {
		t.Errorf("expected 'not running' in output, got: %q", stdout.String())
	}
}

func TestCheckCmd_Running(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 3000, PID: 84312, Process: "node", CWD: "/home/user/projects/myapp"},
	})
	stdout, _ := captureOutput(t)
	noColor = true

	err := runCheck(nil, []string{"3000"})
	var exitErr *ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("expected ExitError{1}, got: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{"● running", "84312", "node", "/home/user/projects/myapp"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

func TestCheckCmd_Running_NoCWD(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 3000, PID: 84312, Process: "node"},
	})
	stdout, _ := captureOutput(t)
	noColor = true

	err := runCheck(nil, []string{"3000"})
	var exitErr *ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("expected ExitError{1}, got: %v", err)
	}

	got := stdout.String()
	if strings.Contains(got, "cwd:") {
		t.Errorf("expected no cwd line when CWD is empty, got:\n%s", got)
	}
}

func TestCheckCmd_Running_NoProcess(t *testing.T) {
	setupMock(t, []detector.ActivePort{
		{Port: 3000, PID: 84312},
	})
	stdout, _ := captureOutput(t)
	noColor = true

	err := runCheck(nil, []string{"3000"})
	var exitErr *ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("expected ExitError{1}, got: %v", err)
	}

	got := stdout.String()
	if strings.Contains(got, "process:") {
		t.Errorf("expected no process line when Process is empty, got:\n%s", got)
	}
}

func TestCheckCmd_ExitCode_Free(t *testing.T) {
	setupMock(t, nil)
	captureOutput(t)

	err := runCheck(nil, []string{"3000"})
	if err != nil {
		t.Errorf("expected exit 0 (nil error) for free port, got: %v", err)
	}
}

func TestCheckCmd_ExitCode_InUse(t *testing.T) {
	setupMock(t, []detector.ActivePort{{Port: 3000, PID: 1}})
	captureOutput(t)

	err := runCheck(nil, []string{"3000"})
	var exitErr *ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Errorf("expected ExitError{1} for in-use port, got: %v", err)
	}
}

func TestCheckCmd_InvalidPort(t *testing.T) {
	setupMock(t, nil)
	captureOutput(t)

	if err := runCheck(nil, []string{"notaport"}); err == nil {
		t.Fatal("expected error for invalid port argument")
	}
}

func TestCheckCmd_PortOutOfRange(t *testing.T) {
	setupMock(t, nil)
	captureOutput(t)

	if err := runCheck(nil, []string{"99999"}); err == nil {
		t.Fatal("expected error for out-of-range port")
	}
}
