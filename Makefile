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

# make a container that is based off the hpcaas-container-base, but also has go installed
# replace CMD call to runit with our go installation and testing command
build-test-container:
ifeq ($(and $(DOCKER_BUILD_NETWORK),$(DOCKER_BUILD_PROXY)),)
	echo "No docker build network found!"
	docker build -t hpcaas-daemon-test-container -f tests/Dockerfile .
else
	echo "Docker build network found!"
	docker build -t hpcaas-daemon-test-container --network $(DOCKER_BUILD_NETWORK) --build-arg http_proxy=$(DOCKER_BUILD_PROXY) -f tests/Dockerfile .
endif

test:
	docker run --rm -e HPCAAS_DAEMON_TEST_CONTAINER=true -v "$(shell pwd)":/go/src/hpcaas-container-daemon -w /go/src/hpcaas-container-daemon hpcaas-daemon-test-container /bin/bash -c "go get -t -v; go test ./..."

