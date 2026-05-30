package watcher_test

import (
	"errors"
	"testing"
	"time"

	"github.com/bhpcv252/portmap/internal/detector"
	"github.com/bhpcv252/portmap/internal/registry"
	"github.com/bhpcv252/portmap/internal/watcher"
)

type mockDet struct {
	ports []detector.ActivePort
	err   error
}

func (m *mockDet) ActivePorts() ([]detector.ActivePort, error) {
	return m.ports, m.err
}

func (m *mockDet) IsActive(port int) (bool, *detector.ActivePort, error) {
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

func makeWatcher(claims map[string]registry.Claim, mock *mockDet) *watcher.Watcher {
	return &watcher.Watcher{
		Registry: &registry.Registry{Version: 1, Claims: claims},
		Detector: mock,
		Interval: time.Second,
	}
}

func collect(
	w *watcher.Watcher,
	prev map[string]watcher.Status,
) ([]watcher.Change, map[string]watcher.Status) {
	var changes []watcher.Change
	next := w.Poll(prev, func(ch watcher.Change) {
		changes = append(changes, ch)
	})
	return changes, next
}

func TestWatcher_StoppedToRunning(t *testing.T) {
	mock := &mockDet{ports: []detector.ActivePort{{Port: 3000, PID: 1234}}}
	w := makeWatcher(map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend"},
	}, mock)

	prev := map[string]watcher.Status{"3000": watcher.StatusStopped}
	changes, _ := collect(w, prev)

	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].From != watcher.StatusStopped || changes[0].To != watcher.StatusRunning {
		t.Errorf("expected stopped->running, got %v->%v", changes[0].From, changes[0].To)
	}
	if changes[0].Project != "myapp" || changes[0].Service != "frontend" {
		t.Errorf("expected myapp/frontend, got %s/%s", changes[0].Project, changes[0].Service)
	}
	if changes[0].Port != 3000 {
		t.Errorf("expected port 3000, got %d", changes[0].Port)
	}
}

func TestWatcher_RunningToStopped(t *testing.T) {
	mock := &mockDet{ports: nil}
	w := makeWatcher(map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend"},
	}, mock)

	prev := map[string]watcher.Status{"3000": watcher.StatusRunning}
	changes, _ := collect(w, prev)

	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].From != watcher.StatusRunning || changes[0].To != watcher.StatusStopped {
		t.Errorf("expected running->stopped, got %v->%v", changes[0].From, changes[0].To)
	}
}

func TestWatcher_NoChange(t *testing.T) {
	mock := &mockDet{ports: []detector.ActivePort{{Port: 3000, PID: 1234}}}
	w := makeWatcher(map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend"},
	}, mock)

	prev := map[string]watcher.Status{"3000": watcher.StatusRunning}
	changes, _ := collect(w, prev)

	if len(changes) != 0 {
		t.Errorf("expected no changes when state is identical, got %d", len(changes))
	}
}

func TestWatcher_ProjectFilter(t *testing.T) {
	// only 4000 is active, but we're watching myapp only
	mock := &mockDet{ports: []detector.ActivePort{{Port: 4000, PID: 5678}}}
	w := makeWatcher(map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend"},
		"4000": {Project: "ml-service", Service: "api"},
	}, mock)
	w.Project = "myapp"

	prev := map[string]watcher.Status{"3000": watcher.StatusStopped}
	changes, _ := collect(w, prev)

	for _, ch := range changes {
		if ch.Port == 4000 {
			t.Error("expected no event for ml-service port when filtering to myapp")
		}
	}
}

func TestWatcher_MultipleChanges(t *testing.T) {
	mock := &mockDet{ports: []detector.ActivePort{
		{Port: 3000, PID: 100},
		{Port: 3001, PID: 101},
		{Port: 3002, PID: 102},
	}}
	w := makeWatcher(map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "svc1"},
		"3001": {Project: "myapp", Service: "svc2"},
		"3002": {Project: "myapp", Service: "svc3"},
	}, mock)

	prev := map[string]watcher.Status{
		"3000": watcher.StatusStopped,
		"3001": watcher.StatusStopped,
		"3002": watcher.StatusStopped,
	}
	changes, _ := collect(w, prev)

	if len(changes) != 3 {
		t.Errorf("expected 3 changes, got %d", len(changes))
	}
}

func TestWatcher_DetectorError(t *testing.T) {
	mock := &mockDet{err: errors.New("detector failed")}
	w := makeWatcher(map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend"},
	}, mock)

	// should not panic; error causes ports to appear stopped
	changes, _ := collect(w, map[string]watcher.Status{})
	_ = changes
}

func TestWatcher_NewPortFirstObservation(t *testing.T) {
	// when a port appears in the snapshot for the first time (not in prev),
	// no change is emitted since there is no prior state to compare against
	mock := &mockDet{ports: []detector.ActivePort{{Port: 3000, PID: 1234}}}
	w := makeWatcher(map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend"},
	}, mock)

	changes, _ := collect(w, map[string]watcher.Status{})

	if len(changes) != 0 {
		t.Errorf("expected no change on first observation, got %d", len(changes))
	}
}

func TestWatcher_WatchedCount_All(t *testing.T) {
	mock := &mockDet{}
	w := makeWatcher(map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend"},
		"3001": {Project: "myapp", Service: "api"},
		"8000": {Project: "other", Service: "svc"},
	}, mock)

	if got := w.WatchedCount(); got != 3 {
		t.Errorf("expected 3, got %d", got)
	}
}

func TestWatcher_WatchedCount_ProjectFilter(t *testing.T) {
	mock := &mockDet{}
	w := makeWatcher(map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend"},
		"3001": {Project: "myapp", Service: "api"},
		"8000": {Project: "other", Service: "svc"},
	}, mock)
	w.Project = "myapp"

	if got := w.WatchedCount(); got != 2 {
		t.Errorf("expected 2, got %d", got)
	}
}

func TestWatcher_PollReturnsCurrState(t *testing.T) {
	mock := &mockDet{ports: []detector.ActivePort{{Port: 3000, PID: 1234}}}
	w := makeWatcher(map[string]registry.Claim{
		"3000": {Project: "myapp", Service: "frontend"},
	}, mock)

	_, next := collect(w, map[string]watcher.Status{})

	if next["3000"] != watcher.StatusRunning {
		t.Errorf("expected returned state to show 3000 as running, got %v", next["3000"])
	}
}
