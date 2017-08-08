package main

import "testing"
import "os"
import "github.com/stretchr/testify/assert"

var TEST_ENV_VAR = "HPCAAS_DAEMON_TEST_CONTAINER"

func TestServer(t *testing.T) {
	assert.Equal(t, 1, 1)
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
