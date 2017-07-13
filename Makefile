.PHONY: all build build-docker

compress= @if hash upx 2>/dev/null; then\
					upx hpcaas-container-daemon;\
					fi

all: build

build:
	go get -v
	go build -v
	$(call compress)

build-docker:
	docker run --rm -v "$(shell pwd)":/go/src/hpcaas-container-daemon -w /go/src/hpcaas-container-daemon golang:1.8 /bin/bash -c "go get -v; go build -v"
	$(call compress)

test-docker:
	docker run --rm -v "$(shell pwd)":/go/src/hpcaas-container-daemon -w /go/src/hpcaas-container-daemon golang:1.8 /bin/bash -c "go get -v; go test"

