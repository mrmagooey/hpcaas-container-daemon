#!/bin/bash

# install daemon deps
go get -t -v ./...

# install test deps
go get -u github.com/haya14busa/goverage

# run tests
# show coverage summary
# go test -cover ./...
# produce coverage report
goverage -coverprofile=coverage.out ./...

