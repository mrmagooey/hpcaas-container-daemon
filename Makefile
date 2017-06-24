.PHONY: all build 

all: build

build:
	docker run --rm -v "$(shell pwd)":/go/src/hpcaas-cluster-daemon -w /go/src/hpcaas-cluster-daemon golang:1.8 /bin/bash -c "go get -v; go build -v"

