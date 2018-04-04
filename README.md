Job Reaper
================

[![CircleCI](https://circleci.com/gh/sstarcher/job-reaper.svg?style=svg)](https://circleci.com/gh/sstarcher/job-reaper)
[![](https://imagelayers.io/badge/sstarcher/job-reaper:latest.svg)](https://imagelayers.io/?images=sstarcher/job-reaper:latest 'Get your own badge on imagelayers.io')
[![Docker Registry](https://img.shields.io/docker/pulls/sstarcher/job-reaper.svg)](https://registry.hub.docker.com/u/sstarcher/job-reaper)&nbsp;

This repo outputs reaps and alerts on finished kubernetes jobs.

Project: [https://github.com/sstarcher/job-reaper]
(https://github.com/sstarcher/job-reaper)

Docker image: [https://registry.hub.docker.com/u/sstarcher/job-reaper/]
(https://registry.hub.docker.com/u/sstarcher/job-reaper/)


## Usage

## Command Line Options

* `-master` - URL to kubernetes api server (default in-cluster kubernetes)
* `-config` - Path to alerter configuration (default ./config.yaml)
    - No configuration defaults to stdout alerter
* `-failures` - Threshold of allowable failures for a job:
    - `-failures=0`: The job will be reaped on any failures.
    - `-failures=-1`: The job will never be reaped on failures (default).
* `-interval` - Interval in seconds to wait between looking for jobs to reap (default 30 seconds)
* `-log` - Level to log - debug, info, warn, error, fatal, panic (default info)
* `-keep-completed=*duration*` - Duration to keep completed jobs (e.g. `-keep-completed=4h`).
* `-ignore-owned` - Ignore jobs owned by other resources, e.g. by `CronJob`s (which have their own reaping logic).

## Alerter options

Alerters are define in the configuration yaml file.  All alerters that are define will be used.

###Stdout
* Level - debug, info, warn, error, fatal, panic
```yaml
stdout:
    level: info
```

###Sensu
 Sensu has a special templates map that allows for adhoc key/value pairs to be passed to sensu.  The values are processed through golangs templating engine are are URL encoded.  Alerts in uchiwa show up as JIT clients via the name Jobs-NAMESPACE, where namespace is the kubernetes namespace the job was running under.

 The values availble for the template engine are as follows
*  Name
*  Message
*  Status
*  StartTime
*  EndTime
*  ExitCode
*  Namespace

## Examples
### Alerter Config
```yaml
sensu:
    address: localhost:3030
    templates:
      logs: "https://kibana/#/discover?_g=(time:(from:'{{ .StartTime }}',mode:absolute,to:'{{ .EndTime }}'))&empty_value"
      anykey: "{{ .ExitCode }}"
stdout:
    level: info
```

### Kubernetes Pod Definition
```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: job-reaper
spec:
  replicas: 1
  template:
    metadata:
      labels:
        name: job-reaper
    spec:
      containers:
      - name: job-reaper
        image: sstarcher/job-reaper:latest
```
