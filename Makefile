.PHONY: all build build-docker

all: build

build:
	go get -v
	go build -v

build-docker:
	docker run --rm -v "$(shell pwd)":/go/src/hpcaas-container-daemon -w /go/src/hpcaas-container-daemon golang:1.8 /bin/bash -c "go get -v; go build -v"

test-docker:
	docker run --rm -v "$(shell pwd)":/go/src/hpcaas-container-daemon -w /go/src/hpcaas-container-daemon golang:1.8 /bin/bash -c "go get -v; go test"

