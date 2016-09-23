package main

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"github.com/sstarcher/job-reaper/config"
	"github.com/sstarcher/job-reaper/kube"
	"time"
)

var (
	masterURL  = flag.String("master", "", "url to kubernetes api server")
	configPath = flag.String("config", "./config.yaml", "path to alerter configuration")
	threshold  = flag.Uint("threshold", 0, "threshold in seconds for reaping stuck jobs")
	failures   = flag.Int("failures", -1, "threshold of allowable failures for a job")
	interval   = flag.Int("interval", 30, "interval in seconds to wait between looking for jobs to reap")
	logLevel   = flag.String("log", "info", "log level - debug, info, warn, error, fatal, panic")
)

func main() {
	flag.Parse()
	value, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Panic(err.Error())
	}
	log.SetLevel(value)

	alerters := config.NewConfig(configPath)
	if len(*alerters) == 0 {
		log.Fatalf("No valid alerters")
	}

	kube := kube.NewKubeClient(*masterURL, *threshold, *failures, alerters)

	everyTime := time.Duration(*interval) * time.Second
	for {
		current := time.Now()
		kube.Reap()
		sleepDur := everyTime - time.Now().Sub(current)
		time.Sleep(sleepDur)
	}
}
