//go:build !linux && !darwin && !windows

package detector

import "fmt"

type stubDetector struct{}

func New() Detector { return &stubDetector{} }

func (d *stubDetector) ActivePorts() ([]ActivePort, error) {
	return nil, fmt.Errorf("port detection not supported on this platform")
}

func (d *stubDetector) IsActive(port int) (bool, *ActivePort, error) {
	return false, nil, fmt.Errorf("port detection not supported on this platform")
}
