//go:build darwin

package detector

import (
	"testing"
)

func TestParseLsofFieldOutput_Valid(t *testing.T) {
	// field output from: lsof -nP -iTCP -sTCP:LISTEN -Fpcn
	input := "p1234\ncnode\nn*:3000\n" +
		"p5678\ncpython\nn127.0.0.1:8080\n" +
		"p9012\ncruby\nn[::1]:4567\n"

	ports := parseLsofFieldOutput(input)

	if len(ports) != 3 {
		t.Fatalf("expected 3 ports, got %d", len(ports))
	}
	if ports[0].Port != 3000 || ports[0].PID != 1234 || ports[0].Process != "node" {
		t.Errorf("unexpected first entry: %+v", ports[0])
	}
	if ports[1].Port != 8080 || ports[1].PID != 5678 || ports[1].Process != "python" {
		t.Errorf("unexpected second entry: %+v", ports[1])
	}
	if ports[2].Port != 4567 || ports[2].PID != 9012 || ports[2].Process != "ruby" {
		t.Errorf("unexpected third entry: %+v", ports[2])
	}
}

func TestParseLsofFieldOutput_Empty(t *testing.T) {
	ports := parseLsofFieldOutput("")
	if len(ports) != 0 {
		t.Fatalf("expected 0 ports for empty input, got %d", len(ports))
	}
}

func TestParseLsofFieldOutput_Dedup(t *testing.T) {
	// same pid:port from both IPv4 and IPv6 listeners should appear once
	input := "p1234\ncnode\nn*:3000\n" +
		"p1234\ncnode\nn[::1]:3000\n"

	ports := parseLsofFieldOutput(input)
	if len(ports) != 1 {
		t.Fatalf("expected 1 port after dedup, got %d", len(ports))
	}
	if ports[0].Port != 3000 {
		t.Errorf("expected port 3000, got %d", ports[0].Port)
	}
}

func TestParseLsofFieldOutput_WildcardAddress(t *testing.T) {
	input := "p1234\ncnode\nn*:3000\n"
	ports := parseLsofFieldOutput(input)
	if len(ports) != 1 || ports[0].Port != 3000 {
		t.Errorf("expected port 3000 from *:3000, got %+v", ports)
	}
}

func TestParseLsofFieldOutput_LoopbackAddress(t *testing.T) {
	input := "p1234\ncnode\nn127.0.0.1:3000\n"
	ports := parseLsofFieldOutput(input)
	if len(ports) != 1 || ports[0].Port != 3000 {
		t.Errorf("expected port 3000 from 127.0.0.1:3000, got %+v", ports)
	}
}

func TestParseLsofFieldOutput_IPv6Address(t *testing.T) {
	input := "p1234\ncnode\nn[::1]:3000\n"
	ports := parseLsofFieldOutput(input)
	if len(ports) != 1 || ports[0].Port != 3000 {
		t.Errorf("expected port 3000 from [::1]:3000, got %+v", ports)
	}
}

func TestParseLsofFieldOutput_SkipsInvalidPort(t *testing.T) {
	// n line with no colon should be skipped
	input := "p1234\ncnode\nnjust-a-name\np5678\ncpython\nn*:8080\n"
	ports := parseLsofFieldOutput(input)
	if len(ports) != 1 || ports[0].Port != 8080 {
		t.Errorf("expected only port 8080, got %+v", ports)
	}
}

func TestRunLsofPorts_NotFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	ports, err := runLsofPorts()
	// when lsof is not in PATH, exec returns an error (not an ExitError)
	// so runLsofPorts should return nil, err
	if err == nil && len(ports) > 0 {
		t.Error("expected error or empty result when lsof is not in PATH")
	}
}

func TestExtractPort_Valid(t *testing.T) {
	cases := []struct {
		addr string
		want int
	}{
		{"*:3000", 3000},
		{"127.0.0.1:8080", 8080},
		{"[::1]:4567", 4567},
		{"0.0.0.0:443", 443},
	}
	for _, tc := range cases {
		got, err := extractPort(tc.addr)
		if err != nil {
			t.Errorf("extractPort(%q) error: %v", tc.addr, err)
		}
		if got != tc.want {
			t.Errorf("extractPort(%q) = %d, want %d", tc.addr, got, tc.want)
		}
	}
}

func TestExtractPort_NoColon(t *testing.T) {
	_, err := extractPort("noport")
	if err == nil {
		t.Error("expected error for address with no colon")
	}
}
