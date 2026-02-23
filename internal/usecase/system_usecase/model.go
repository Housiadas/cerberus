package system_usecase

// Status represents a readiness status.
type Status struct {
	Status string `json:"status"`
}

// Info represents information about the usecase.
type Info struct {
	Status     string `json:"status,omitempty"`
	Build      string `json:"build,omitempty"`
	Host       string `json:"host,omitempty"`
	Name       string `json:"name,omitempty"`
	PodIP      string `json:"podIp,omitempty"`
	Node       string `json:"node,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	GOMAXPROCS int    `json:"gomaxprocs,omitempty"`
}
