package main

import "testing"
import "os"
import "github.com/stretchr/testify/assert"

var TEST_ENV_VAR = "HPCAAS_DAEMON_TEST_CONTAINER"

func TestServer(t *testing.T) {
	//
	assert.Equal(t, 1, 1)
}

func init() {
	if _, found := os.LookupEnv(TEST_ENV_VAR); found != true {
		panic("All tests should be run in the container, with " + TEST_ENV_VAR + " being set")
	}
}
