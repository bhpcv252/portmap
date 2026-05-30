//go:build integration

package cmd

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/bhpcv252/portmap/internal/registry"
	"github.com/bhpcv252/portmap/internal/watcher"
)

func cancelledCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func TestWatchCmd_PrintsStartupMessage(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, nil)
	stdout, _ := captureOutput(t)
	watchCtxOverride = cancelledCtx()

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
		"3001": {Project: "myapp", Service: "api", ClaimedAt: time.Now().UTC()},
	})

	if err := runWatch(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "watching 2") {
		t.Errorf("expected 'watching 2' in output, got: %q", got)
	}
	if !strings.Contains(got, "Ctrl+C") {
		t.Errorf("expected Ctrl+C hint in output, got: %q", got)
	}
}

func TestWatchCmd_ProjectFilter(t *testing.T) {
	regPath, _ := setupIntegrationWithMock(t, nil)
	stdout, _ := captureOutput(t)
	watchCtxOverride = cancelledCtx()
	watchProject = "myapp"

	seedRegistry(t, regPath, map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend", ClaimedAt: time.Now().UTC()},
		"3001": {Project: "myapp", Service: "api", ClaimedAt: time.Now().UTC()},
		"8000": {Project: "ml-service", Service: "inference", ClaimedAt: time.Now().UTC()},
	})

	if err := runWatch(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "watching 2") {
		t.Errorf("expected 'watching 2' when filtering to myapp, got: %q", stdout.String())
	}
}

func TestWatchCmd_EmptyRegistry(t *testing.T) {
	setupIntegrationWithMock(t, nil)
	stdout, _ := captureOutput(t)
	watchCtxOverride = cancelledCtx()

	if err := runWatch(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "watching 0") {
		t.Errorf("expected 'watching 0' for empty registry, got: %q", stdout.String())
	}
}

func TestWatchCmd_IntervalInMessage(t *testing.T) {
	setupIntegrationWithMock(t, nil)
	stdout, _ := captureOutput(t)
	watchCtxOverride = cancelledCtx()
	watchInterval = 10

	if err := runWatch(nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "10s") {
		t.Errorf("expected interval in startup message, got: %q", stdout.String())
	}
}

func TestWatchStatusLabel_Running(t *testing.T) {
	if got := watchStatusLabel(watcher.StatusRunning); got != "● running" {
		t.Errorf("expected '● running', got %q", got)
	}
}

func TestWatchStatusLabel_Stopped(t *testing.T) {
	if got := watchStatusLabel(watcher.StatusStopped); got != "○ stopped" {
		t.Errorf("expected '○ stopped', got %q", got)
	}
}
