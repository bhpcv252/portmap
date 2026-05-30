//go:build linux

package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bhpcv252/portmap/internal/testutil"
)

func fixturesDir() string { return testutil.FixturesDir() }

func TestParseProcNetTCP_Valid(t *testing.T) {
	content := testutil.MustReadFile(t, filepath.Join(fixturesDir(), "proc_net_tcp.txt"))
	entries := parseProcNetTCP(content)

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestParseProcNetTCP_Empty(t *testing.T) {
	entries := parseProcNetTCP("  sl  local_address rem_address   st\n")
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestParseProcNetTCP_MalformedLine(t *testing.T) {
	content := "  sl  local_address\n   0: toofew\n"
	entries := parseProcNetTCP(content)
	if len(entries) != 0 {
		t.Error("expected malformed line to be skipped")
	}
}

func TestParseProcNetTCP_HexConversion(t *testing.T) {
	// 0BB8 hex = 3000 decimal
	content := "  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode\n" +
		"   0: 00000000:0BB8 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 99999 1 0000000000000000 100 0 0 10 0\n"
	entries := parseProcNetTCP(content)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].port != 3000 {
		t.Errorf("expected port 3000, got %d", entries[0].port)
	}
}

func TestParseProcNetTCP_OnlyListenState(t *testing.T) {
	content := "  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode\n" +
		"   0: 00000000:0BB8 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 11111 1 0000000000000000 100 0 0 10 0\n" +
		"   1: 00000000:1F90 00000000:0000 01 00000000:00000000 00:00000000 00000000     0        0 22222 1 0000000000000000 100 0 0 10 0\n"

	entries := parseProcNetTCP(content)
	if len(entries) != 1 {
		t.Fatalf("expected only LISTEN row, got %d entries", len(entries))
	}
	if entries[0].port != 3000 {
		t.Errorf("expected port 3000, got %d", entries[0].port)
	}
}

func TestParseProcNetTCP6_Valid(t *testing.T) {
	content := testutil.MustReadFile(t, filepath.Join(fixturesDir(), "proc_net_tcp6.txt"))
	// 11D7 hex = 4567 decimal
	entries := parseProcNetTCP(content)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry from tcp6 fixture, got %d", len(entries))
	}
	if entries[0].port != 4567 {
		t.Errorf("expected port 4567, got %d", entries[0].port)
	}
}

func TestMapInodeToPID_Found(t *testing.T) {
	// build a minimal /proc structure in a temp dir
	procRoot := t.TempDir()
	pidDir := filepath.Join(procRoot, "1234", "fd")
	if err := os.MkdirAll(pidDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// create a symlink named "3" pointing to "socket:[99999]"
	if err := os.Symlink("socket:[99999]", filepath.Join(pidDir, "3")); err != nil {
		t.Fatal(err)
	}

	pid, err := mapInodeToPID("99999", procRoot)
	if err != nil {
		t.Fatal(err)
	}
	if pid != 1234 {
		t.Errorf("expected PID 1234, got %d", pid)
	}
}

func TestMapInodeToPID_NotFound(t *testing.T) {
	procRoot := t.TempDir()
	pidDir := filepath.Join(procRoot, "1234", "fd")
	if err := os.MkdirAll(pidDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("socket:[11111]", filepath.Join(pidDir, "3")); err != nil {
		t.Fatal(err)
	}

	pid, err := mapInodeToPID("99999", procRoot)
	if err != nil {
		t.Fatal(err)
	}
	if pid != 0 {
		t.Errorf("expected 0 (not found), got %d", pid)
	}
}
