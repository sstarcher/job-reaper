TEST? = $$(glide nv)

GO_SOURCE_FILES = $(shell find . -type f -name "*.go" | grep -v /vendor/)

vendor:
	glide install

all: vendor build test docker

docker:
	docker build .

clean:
	rm -rf build

test: vendor
	go test -v $(TEST)

build: vendor $(GO_SOURCE_FILES)
	go build -o build/job-reaper cmd/main.go

run: build
	./build/job-reaper --config=./myconfig.yaml --master=localhost:8080 --failures=0

mini:
	minikube ip || $$(pkill -9 -f 'kubectl proxy'; minikube start; kubectl proxy --port=8080 &)

kube_clean: mini
	kubectl delete scheduledjobs --all
	kubectl delete jobs --all
	kubectl delete pods --all

jobs: kube_clean
	# Success
	kubectl run succeed --schedule="*/5 * * * ?" --image=perl --restart=OnFailure -- perl -Mbignum=bpi -wle 'print bpi(2000)'
	# Always Fail
	kubectl run always-fail --schedule="*/1 * * * ?" --image=busybox --restart=Never -- /bin/sh -c 'sleep 10; exit 5'
	kubectl annotate always-fail succeed job-reaper.github.sstarcher.io/channel='#listhub-syseng-status'
	# Image pull Error
	#kubectl run image-pull-error --schedule="*/1 * * * ?" --image=buxybox --restart=Never -- /bin/sh
	# RunContainerError
	#kubectl run run-container-error --schedule="*/1 * * * ?" --image=busybox --restart=Never -- /bin/bash

annotate:
	kubectl annotate scheduledjob succeed job-reaper.github.sstarcher.io/channel=listhub-syseng

.PHONY: all docker test run