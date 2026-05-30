//go:build darwin

package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bhpcv252/portmap/internal/testutil"
)

func fixturesDir() string { return testutil.FixturesDir() }

func TestParseLsofOutput_Valid(t *testing.T) {
	content := testutil.MustReadFile(t, filepath.Join(fixturesDir(), "lsof_output.txt"))
	ports := parseLsofOutput(content)

	if len(ports) != 3 {
		t.Fatalf("expected 3 ports, got %d", len(ports))
	}
	if ports[0].Port != 3000 || ports[0].PID != 1234 || ports[0].Process != "node" {
		t.Errorf("unexpected first entry: %+v", ports[0])
	}
	if ports[1].Port != 8080 || ports[1].PID != 5678 {
		t.Errorf("unexpected second entry: %+v", ports[1])
	}
	if ports[2].Port != 4567 {
		t.Errorf("unexpected third entry: %+v", ports[2])
	}
}

func TestParseLsofOutput_Empty(t *testing.T) {
	ports := parseLsofOutput("")
	if len(ports) != 0 {
		t.Fatalf("expected 0 ports for empty input, got %d", len(ports))
	}
}

func TestParseLsofOutput_MalformedLine(t *testing.T) {
	ports := parseLsofOutput("COMMAND PID\ntoofew cols\n")
	if len(ports) != 0 {
		t.Error("expected malformed line to be skipped")
	}
}

func TestParseLsofOutput_WildcardAddress(t *testing.T) {
	line := "COMMAND   PID USER   FD   TYPE DEVICE SIZE/OFF NODE NAME\n" +
		"node     1234 user   22u  IPv4  12345      0t0  TCP *:3000 (LISTEN)\n"
	ports := parseLsofOutput(line)
	if len(ports) != 1 || ports[0].Port != 3000 {
		t.Errorf("expected port 3000 from *:3000, got %+v", ports)
	}
}

func TestParseLsofOutput_LoopbackAddress(t *testing.T) {
	line := "COMMAND   PID USER   FD   TYPE DEVICE SIZE/OFF NODE NAME\n" +
		"node     1234 user   22u  IPv4  12345      0t0  TCP 127.0.0.1:3000 (LISTEN)\n"
	ports := parseLsofOutput(line)
	if len(ports) != 1 || ports[0].Port != 3000 {
		t.Errorf("expected port 3000 from 127.0.0.1:3000, got %+v", ports)
	}
}

func TestParseLsofOutput_IPv6Address(t *testing.T) {
	line := "COMMAND   PID USER   FD   TYPE DEVICE SIZE/OFF NODE NAME\n" +
		"node     1234 user   22u  IPv6  12345      0t0  TCP [::1]:3000 (LISTEN)\n"
	ports := parseLsofOutput(line)
	if len(ports) != 1 || ports[0].Port != 3000 {
		t.Errorf("expected port 3000 from [::1]:3000, got %+v", ports)
	}
}

func TestRunLsof_NotFound(t *testing.T) {
	// override PATH to make lsof unfindable
	t.Setenv("PATH", t.TempDir())

	// re-create a temp dir with no lsof binary
	emptyDir := t.TempDir()
	t.Setenv("PATH", emptyDir)

	_, err := runLsof()
	if err == nil {
		// lsof may still be found if the test runs with a very permissive PATH.
		// only fail if PATH was truly empty and lsof ran anyway
		lsofPath, _ := filepath.Abs(emptyDir)
		if _, statErr := os.Stat(filepath.Join(lsofPath, "lsof")); os.IsNotExist(statErr) {
			t.Error("expected error when lsof is not in PATH")
		}
	}
}
