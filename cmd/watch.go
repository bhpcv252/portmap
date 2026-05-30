package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/bhpcv252/portmap/internal/iohelp"
	"github.com/bhpcv252/portmap/internal/registry"
	"github.com/bhpcv252/portmap/internal/watcher"
)

var (
	watchProject     string
	watchInterval    int
	watchNotify      bool
	watchCtxOverride context.Context // injected by tests to avoid blocking
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch claimed ports and print status changes",
	Args:  cobra.NoArgs,
	RunE:  runWatch,
}

func init() {
	watchCmd.Flags().StringVarP(&watchProject, "project", "p", "", "Watch only a specific project")
	watchCmd.Flags().IntVar(&watchInterval, "interval", 5, "Polling interval in seconds")
	watchCmd.Flags().BoolVar(&watchNotify, "notify", false, "Send a desktop notification on change")
}

func runWatch(cmd *cobra.Command, args []string) error {
	r, err := registry.Load(getRegistryPath())
	if err != nil {
		return err
	}

	w := &watcher.Watcher{
		Registry: r,
		Detector: getDetector(),
		Interval: time.Duration(watchInterval) * time.Second,
		Project:  watchProject,
	}

	ew := &iohelp.ErrWriter{W: out}
	ew.Printf("watching %d claimed port(s) (polling every %ds)\n", w.WatchedCount(), watchInterval)
	ew.Println("press Ctrl+C to stop\n")
	if ew.Err != nil {
		return ew.Err
	}

	ctx := watchCtxOverride
	if ctx == nil {
		var stop context.CancelFunc
		ctx, stop = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
	}

	return w.Run(ctx, func(ch watcher.Change) {
		from := watchStatusLabel(ch.From)
		to := watchStatusLabel(ch.To)
		line := fmt.Sprintf("[%s] port %d (%s/%s)  %s -> %s",
			time.Now().Format("15:04:05"),
			ch.Port, ch.Project, ch.Service,
			from, to,
		)
		ew2 := &iohelp.ErrWriter{W: out}
		ew2.Println(line)

		if watchNotify {
			sendNotification(fmt.Sprintf("port %d (%s/%s): %s -> %s",
				ch.Port, ch.Project, ch.Service, from, to))
		}
	})
}

func watchStatusLabel(s watcher.Status) string {
	if s == watcher.StatusRunning {
		return "● running"
	}
	return "○ stopped"
}

func sendNotification(msg string) {
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(`display notification %q with title "portmap"`, msg)
		_ = exec.Command("osascript", "-e", script).Run()
	case "linux":
		_ = exec.Command("notify-send", "portmap", msg).Run()
	}
}
