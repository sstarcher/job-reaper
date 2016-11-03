package sensu

import (
	"bytes"
	"encoding/json"
	"errors"

	"fmt"
	"github.com/sstarcher/job-reaper/alert"
	"io/ioutil"
	"net"
	"net/url"
	"regexp"
	"text/template"
	"time"
)

var validNamePattern = regexp.MustCompile(`^[\w\.-]+$`)
var prefix = "Jobs-"

// Service configuration data for sensu
type Service struct {
	Address   string
	Templates map[string]string
}

func generateTemplate(templateStr string, data alert.Data) (string, error) {
	tmpl, err := template.New("test").Parse(templateStr)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Validate sensu configuration
func (s Service) Validate() error {
	if s.Address == "" {
		return errors.New("Sensu requires an address")
	}
	return nil
}

// Send alert to sensu
func (s Service) Send(data alert.Data) error {
	var err error
	if !validNamePattern.MatchString(data.Name) {
		return fmt.Errorf("invalid name %q for sensu alert. Must match %v", data.Name, validNamePattern)
	}

	postData := make(map[string]interface{})
	postData["name"] = data.Name
	postData["source"] = prefix + data.Namespace
	postData["output"] = data.Message
	postData["exitCode"] = data.ExitCode

	postData["status"] = 0
	if data.ExitCode != 0 {
		postData["status"] = 1
	}

	for k, v := range s.Templates {
		value, err := generateTemplate(v, data)
		if err != nil {
			return err
		}

		postData[k], err = urlEncoded(value)
		if err != nil {
			return err
		}
	}

	conn, err := net.DialTimeout("tcp", s.Address, time.Second*3)
	if err != nil {
		return err
	}
	defer conn.Close()

	enc := json.NewEncoder(conn)
	err = enc.Encode(postData)
	if err != nil {
		return err
	}
	resp, err := ioutil.ReadAll(conn)
	if string(resp) != "ok" {
		return errors.New("sensu socket error: " + string(resp))
	}
	return nil
}

func urlEncoded(str string) (string, error) {
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
