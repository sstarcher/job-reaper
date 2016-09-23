package config

import (
	log "github.com/Sirupsen/logrus"
	"github.com/sstarcher/job-reaper/alert"
	"github.com/sstarcher/job-reaper/alert/sensu"
	"github.com/sstarcher/job-reaper/alert/stdout"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

// Config data for alerters
type Config struct {
	Sensu  sensu.Service
	Stdout stdout.Service
}

var defaultConfig = []byte(`
stdout:
  level: info
`)

// NewConfig loads yaml configuration from path
func NewConfig(path *string) *[]alert.Alert {
	data, err := ioutil.ReadFile(*path)
	if os.IsNotExist(err) {
		log.Warn("Configuration file does not exist defaulting to stdout alerter")
		data = defaultConfig
	}
	return load(data)
}

func loadConfig(data []byte) Config {
	config := Config{}
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return config
}

func load(data []byte) *[]alert.Alert {
	config := loadConfig(data)

	var alerters = &[]alert.Alert{}
	process(config.Stdout, alerters)
	process(config.Sensu, alerters)

	return alerters
}

func process(alerter alert.Alert, alerters *[]alert.Alert) {
	if isEmpty(alerter) != true {
		structName := reflect.TypeOf(alerter).String()
		alerterName := strings.Split(structName, ".")[0]

		err := alerter.Validate()
		if err != nil {
			log.Errorf("error for %s - %v", alerterName, err)
		} else {
			log.Infof("Adding alerter for %s", alerterName)
			*alerters = append(*alerters, alerter)
		}
	}
}

func isEmpty(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
