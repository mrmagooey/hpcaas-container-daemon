.PHONY: all build build-docker

compress= @if hash upx 2>/dev/null; then\
					upx hpcaas-container-daemon;\
					fi

all: build

build:
	dep ensure
	go build -v
	$(call compress)

build-docker:
	docker run --rm -v "$(shell pwd)":/go/src/hpcaas-container-daemon -w /go/src/hpcaas-container-daemon golang:1.9.2 /bin/bash -c "go get -u github.com/golang/dep/cmd/dep; dep ensure; go build -v"
	$(call compress)

test-image-name=hpcaas-daemon-test-image
test-container-name=hpcaas-daemon-test
# make a container that is based off the hpcaas-container-base, but also has go installed
# replace CMD call to runit with our go installation and testing command
build-test-container:
ifeq ($(and $(DOCKER_BUILD_NETWORK),$(DOCKER_BUILD_PROXY)),)
	echo "No docker build network found!"
	docker build -t $(test-image-name) -f tests/Dockerfile .
else
	echo "Docker build network found!"
	docker build -t $(test-image-name) --network $(DOCKER_BUILD_NETWORK) --build-arg http_proxy=$(DOCKER_BUILD_PROXY) -f tests/Dockerfile .
endif

docker_test_container_exists=$(shell docker ps -a | grep $(test-container-name) > /dev/null 2>&1; echo $$?)

docker_go_path=/go/src/github.com/mrmagooey/hpcaas-container-daemon

test:
ifeq ($(docker_test_container_exists), 0)
	docker start -a $(test-container-name)
else
	docker run --name $(test-container-name) -e HPCAAS_DAEMON_TEST_CONTAINER=true -v "$(shell pwd)":$(docker_go_path) -w $(docker_go_path) $(test-image-name):latest /bin/bash -c "tests/run_tests.sh"
endif

test-show-coverage-report:
	go tool cover -html=coverage.out

