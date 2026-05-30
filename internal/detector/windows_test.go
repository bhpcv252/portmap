//go:build windows

package detector

import (
	"path/filepath"
	"testing"

	"github.com/bhpcv252/portmap/internal/testutil"
)

func fixturesDir() string { return testutil.FixturesDir() }

func TestParseNetstatOutput_Valid(t *testing.T) {
	content := testutil.MustReadFile(t, filepath.Join(fixturesDir(), "netstat_output.txt"))
	entries := parseNetstatOutput(content)

	if len(entries) != 2 {
		t.Fatalf("expected 2 LISTENING entries, got %d", len(entries))
	}
	if entries[0].port != 3000 || entries[0].pid != 1234 {
		t.Errorf("unexpected first entry: %+v", entries[0])
	}
	if entries[1].port != 8080 || entries[1].pid != 5678 {
		t.Errorf("unexpected second entry: %+v", entries[1])
	}
}

func TestParseNetstatOutput_OnlyListening(t *testing.T) {
	content := testutil.MustReadFile(t, filepath.Join(fixturesDir(), "netstat_output.txt"))
	entries := parseNetstatOutput(content)
	for _, e := range entries {
		if e.port == 49152 {
			t.Error("expected ESTABLISHED row to be filtered out")
		}
	}
}

func TestParseNetstatOutput_Empty(t *testing.T) {
	content := "\nActive Connections\n\n  Proto  Local Address  Foreign Address  State  PID\n"
	entries := parseNetstatOutput(content)
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries for header-only input, got %d", len(entries))
	}
}

func TestParseNetstatOutput_MalformedLine(t *testing.T) {
	content := "  Proto  Local\ntoofew\n"
	entries := parseNetstatOutput(content)
	if len(entries) != 0 {
		t.Error("expected malformed line to be skipped")
	}
}

func TestParseTasklist_Valid(t *testing.T) {
	output := `"node.exe","1234","Console","1","10,000 K"` + "\n" +
		`"python.exe","5678","Console","1","20,000 K"` + "\n"
	pids := map[int]bool{1234: true, 5678: true}

	names := parseTasklist(output, pids)
	if names[1234] != "node.exe" {
		t.Errorf("expected node.exe for PID 1234, got %q", names[1234])
	}
	if names[5678] != "python.exe" {
		t.Errorf("expected python.exe for PID 5678, got %q", names[5678])
	}
}

func TestParseTasklist_PIDNotFound(t *testing.T) {
	output := `"node.exe","1234","Console","1","10,000 K"` + "\n"
	pids := map[int]bool{9999: true}

	names := parseTasklist(output, pids)
	if names[9999] != "" {
		t.Errorf("expected empty string for unknown PID, got %q", names[9999])
	}
}
