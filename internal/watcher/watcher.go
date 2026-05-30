package watcher

import (
	"context"
	"strconv"
	"time"

	"github.com/bhpcv252/portmap/internal/detector"
	"github.com/bhpcv252/portmap/internal/registry"
)

type Status string

const (
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
)

type Change struct {
	Port    int
	Project string
	Service string
	From    Status
	To      Status
}

type Watcher struct {
	Registry *registry.Registry
	Detector detector.Detector
	Interval time.Duration
	Project  string // empty means watch all projects
}

func (w *Watcher) Poll(prev map[string]Status, onChange func(Change)) map[string]Status {
	curr := w.snapshot()
	for port, to := range curr {
		from, ok := prev[port]
		if !ok {
			from = to
		}
		if from != to {
			n, _ := strconv.Atoi(port)
			ch := Change{Port: n, From: from, To: to}
			if c := w.Registry.GetClaim(port); c != nil {
				ch.Project = c.Project
				ch.Service = c.Service
			}
			onChange(ch)
		}
	}
	return curr
}

func (w *Watcher) Run(ctx context.Context, onChange func(Change)) error {
	prev := w.snapshot()
	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			prev = w.Poll(prev, onChange)
		}
	}
}

func (w *Watcher) WatchedCount() int {
	count := 0
	for _, c := range w.Registry.Claims {
		if w.Project == "" || c.Project == w.Project {
			count++
		}
	}
	return count
}

func (w *Watcher) snapshot() map[string]Status {
	result := make(map[string]Status)
	for port, c := range w.Registry.Claims {
		if w.Project != "" && c.Project != w.Project {
			continue
		}
		n, err := strconv.Atoi(port)
		if err != nil {
			continue
		}
		active, _, _ := w.Detector.IsActive(n)
		if active {
			result[port] = StatusRunning
		} else {
			result[port] = StatusStopped
		}
	}
	return result
}
