package detector

type Detector interface {
	ActivePorts() ([]ActivePort, error)
	IsActive(port int) (bool, *ActivePort, error)
}

type ActivePort struct {
	Port    int
	PID     int
	Process string
}
