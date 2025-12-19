package services

type ServiceStatus struct {
	Running bool `json:"running"`
	PID     int  `json:"pid"`
	Port    int  `json:"port"`
}

type Service interface {
	Name() string
	Start() error
	Stop() error
	Status() ServiceStatus
}
