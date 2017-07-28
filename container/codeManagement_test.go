package container

import "testing"
import "os"
import "github.com/mrmagooey/hpcaas-container-daemon/state"

var TEST_ENV_VAR = "HPCAAS_DAEMON_TEST_CONTAINER"

// test that a binary can be successfully started
func TestExecuteCode(t *testing.T) {

	state.SetCodeName("")
	ExecuteCode()

}

// test that a started binary can have its return
func TestWatchCode(t *testing.T) {
	//
}

func TestKillCode(t *testing.T) {
	//
}

func init() {
	if _, found := os.LookupEnv(TEST_ENV_VAR); found != true {
		panic("All tests should be run in the container, with " + TEST_ENV_VAR + " being set")
	}
}
