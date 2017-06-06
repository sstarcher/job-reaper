package main

import (
	"flag"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/sstarcher/job-reaper/config"
	"github.com/sstarcher/job-reaper/kube"
)

var (
	masterURL   = flag.String("master", "", "url to kubernetes api server")
	configPath  = flag.String("config", "./config.yaml", "path to alerter configuration")
	failures    = flag.Int("failures", -1, "threshold of allowable failures for a job")
	interval    = flag.Int("interval", 30, "interval in seconds to wait between looking for jobs to reap")
	logLevel    = flag.String("log", "info", "log level - debug, info, warn, error, fatal, panic")
	reaperCount = flag.Int("reapers", 2, "Number of reaper routines to run")
	bufferRatio = flag.Int("buffer", 1, "Multiplier for buffer size compared to reaper count.")
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	flag.Parse()
	value, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Panic(err.Error())
	}
	log.SetLevel(value)

	alerters := config.NewConfig(configPath)
	if len(*alerters) == 0 {
		log.Fatal("No valid alerters")
	}

	if *reaperCount <= 0 {
		log.Fatal("reaper count must be greater than 0")
	}

	if *bufferRatio < 1 {
		log.Fatal("buffer ratio must be at least 1")
	}

	kube := kube.NewKubeClient(*masterURL, *failures, alerters, *reaperCount, *bufferRatio)

	everyTime := time.Duration(*interval) * time.Second
	for {
		current := time.Now()
		kube.Reap()
		sleepDur := everyTime - time.Now().Sub(current)
		time.Sleep(sleepDur)
	}
}
