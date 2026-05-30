package detector

import (
	"errors"
	"testing"
)

type mockDetector struct {
	ports []ActivePort
	err   error
}

func (m *mockDetector) ActivePorts() ([]ActivePort, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.ports, nil
}

func (m *mockDetector) IsActive(port int) (bool, *ActivePort, error) {
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

func TestMockDetector_ActivePorts(t *testing.T) {
	m := &mockDetector{ports: []ActivePort{
		{Port: 3000, PID: 1234, Process: "node"},
		{Port: 8080, PID: 5678, Process: "python"},
	}}

	ports, err := m.ActivePorts()
	if err != nil {
		t.Fatal(err)
	}
	if len(ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(ports))
	}
}

func TestMockDetector_IsActive_Found(t *testing.T) {
	m := &mockDetector{ports: []ActivePort{{Port: 3000, PID: 1234, Process: "node"}}}

	ok, ap, err := m.IsActive(3000)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected port 3000 to be active")
	}
	if ap == nil || ap.PID != 1234 {
		t.Errorf("expected PID 1234, got %+v", ap)
	}
}

func TestMockDetector_IsActive_NotFound(t *testing.T) {
	m := &mockDetector{ports: []ActivePort{{Port: 3000, PID: 1234, Process: "node"}}}

	ok, ap, err := m.IsActive(9999)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expected port 9999 to be inactive")
	}
	if ap != nil {
		t.Errorf("expected nil ActivePort, got %+v", ap)
	}
}

func TestMockDetector_Error(t *testing.T) {
	want := errors.New("detector failed")
	m := &mockDetector{err: want}

	_, err := m.ActivePorts()
	if !errors.Is(err, want) {
		t.Errorf("expected %v, got %v", want, err)
	}
}
