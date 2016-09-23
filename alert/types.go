package alert

import (
	"time"
)

// Alert Interface for sending alerts
type Alert interface {
	Send(data Data) error
	Validate() error
}

// Data structure available for the alerter
type Data struct {
	Name      string
	Message   string
	Status    string
	StartTime time.Time
	EndTime   time.Time
	ExitCode  int
	Namespace string
}
