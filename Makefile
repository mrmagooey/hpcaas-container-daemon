.PHONY: all build 

all: build

build:
	docker run --rm -v "$(shell pwd)":/go/src/hpcaas-container-daemon -w /go/src/hpcaas-container-daemon golang:1.8 /bin/bash -c "go get -v; go build -v"

