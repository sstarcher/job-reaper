package stdout

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/sstarcher/job-reaper/alert"
)

// Service structure for stdout
type Service struct {
	Level string
}

// Validate sensu configuration
func (s Service) Validate() error {
	if s.Level != "info" {
		return errors.New("info level is the only support level")
	}
	return nil
}

// Send alert to stdout
func (s Service) Send(data alert.Data) error {
	value := fmt.Sprintf("%s with exit code [%d] for %s", data.Status, data.ExitCode, data.Message)
	log.Infof("[%s] Reaping @ [%s] @ %s", data.Name, value, data.EndTime.String())

	return nil
}
