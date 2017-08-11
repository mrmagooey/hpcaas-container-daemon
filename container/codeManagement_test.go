package container

import "testing"
import "os"
import "github.com/stretchr/testify/assert"
import "time"
import "github.com/mrmagooey/hpcaas-container-daemon/state"
import "errors"
import "fmt"
import "os/exec"
import "bytes"

var TEST_ENV_VAR = "HPCAAS_DAEMON_TEST_CONTAINER"

func TestParent(t *testing.T) {
	fmt.Println("")
	t.Run("_testExecuteLs", _testExecuteLs)
	t.Run("_testStdout", _testStdout)
	t.Run("_testExecuteSleep", _testExecuteSleep)
	t.Run("_testCodeAlreadyStarted", _testCodeAlreadyStarted)
	t.Run("_testCodeMissing", _testCodeMissing)
	t.Run("_testEnvVars", _testEnvVars)
	t.Run("_testKillCode", _testKillCode)
	t.Run("_testCodeStartsThenReturnsError", _testCodeStartsThenReturnsError)
	t.Run("_testCodeFailToStart", _testCodeFailToStart)
	t.Run("_testCodeStartedExternally", _testCodeStartedExternally)
}

// test that a binary can be successfully started
func _testExecuteLs(t *testing.T) {
	assert := assert.New(t)
	state.InitState()
	os.Symlink("/bin/ls", "/hpcaas/code/myls")
	defer os.Remove("/hpcaas/code/myls")
	state.SetCodeName("myls")
	assert.Equal(state.CODE_WAITING, state.GetCodeState())
	err := ExecuteCode()
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(state.CODE_STOPPED, state.GetCodeState())
}

// test that we can read stdout
func _testStdout(t *testing.T) {
	assert := assert.New(t)
	state.InitState()
	if err := os.Symlink("/bin/echo", "/hpcaas/code/myecho"); err != nil {
		t.Error(err)
	}
	defer os.Remove("/hpcaas/code/myecho")
	state.SetCodeName("myecho")
	state.SetCodeArguments([]string{"hello"})
	if err := ExecuteCode(); err != nil {
		t.Error(err)
	}
	time.Sleep(1 * time.Second)
	assert.Equal("hello\n", state.GetCodeStdout())
}

// test that long running processes are tracked
func _testExecuteSleep(t *testing.T) {
	assert := assert.New(t)
	state.InitState()
	if err := os.Symlink("/bin/sleep", "/hpcaas/code/mysleep"); err != nil {
		t.Error(err)
	}
	defer os.Remove("/hpcaas/code/mysleep")
	state.SetCodeName("mysleep")
	state.SetCodeArguments([]string{"1"})
	assert.Equal(state.CODE_WAITING, state.GetCodeState())
	err := ExecuteCode()
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(state.CODE_RUNNING, state.GetCodeState())
	time.Sleep(2 * time.Second)
	assert.Equal(state.CODE_STOPPED, state.GetCodeState())
}

// test that only one binary can be running at one time
func _testCodeAlreadyStarted(t *testing.T) {
	// test that a started binary can have its return
	assert := assert.New(t)
	state.InitState()
	// reset code state from any other tests
	if err := os.Symlink("/bin/sleep", "/hpcaas/code/mysleep"); err != nil {
		t.Error(err)
	}
	defer os.Remove("/hpcaas/code/mysleep")
	state.SetCodeName("mysleep")
	state.SetCodeArguments([]string{"1"})
	assert.Equal(state.CODE_WAITING, state.GetCodeState())
	err := ExecuteCode()
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(state.CODE_RUNNING, state.GetCodeState())
	err = ExecuteCode()
	if assert.Error(err) {
		assert.Equal(errors.New("Code already started"), err)
	}
	// need to wait for sleep 1 to complete
	time.Sleep(2 * time.Second)
	assert.Equal(state.CODE_STOPPED, state.GetCodeState())
}

// test that missing code raises an error
func _testCodeMissing(t *testing.T) {
	assert := assert.New(t)
	state.InitState()
	state.SetCodeName("does_not_exist")
	assert.Equal(state.CODE_WAITING, state.GetCodeState())
	err := ExecuteCode()
	assert.Equal(state.CODE_MISSING, state.GetCodeState())
	if assert.Error(err) {
		assert.Equal(errors.New("Code executable is missing"), err)
	}
}

// test that we can give environment variables to our binaries
func _testEnvVars(t *testing.T) {
	assert := assert.New(t)
	state.InitState()
	if err := os.Symlink("/usr/bin/env", "/hpcaas/code/myenv"); err != nil {
		t.Error(err)
	}
	defer os.Remove("/hpcaas/code/myenv")
	state.SetCodeName("myenv")
	state.SetCodeParams(map[string]string{
		"hello": "world",
	})
	if err := ExecuteCode(); err != nil {
		t.Error(err)
	}
	time.Sleep(50 * time.Millisecond)
	assert.Equal(state.CODE_STOPPED, state.GetCodeState())
	assert.Equal("hello=world\n", state.GetCodeStdout())
}

// test that we can kill a running binary
func _testKillCode(t *testing.T) {
	assert := assert.New(t)
	state.InitState()
	// reset code state from any other tests
	if err := os.Symlink("/bin/sleep", "/hpcaas/code/mysleep"); err != nil {
		t.Error(err)
	}
	defer os.Remove("/hpcaas/code/mysleep")
	state.SetCodeName("mysleep")
	state.SetCodeArguments([]string{"1000"})

	assert.Equal(state.CODE_WAITING, state.GetCodeState())
	err := ExecuteCode()
	if err != nil {
		t.Error(err)
	}
	// wait till it starts
	time.Sleep(100 * time.Millisecond)
	assert.Equal(state.CODE_RUNNING, state.GetCodeState())
	// kill it
	if err := KillCode(); err != nil {
		t.Error(err)
		return
	}
	// TODO parse the process tree to check that we are actually killing the process
	// wait till it dies
	time.Sleep(100 * time.Millisecond)
	assert.Equal(state.CODE_KILLED, state.GetCodeState())
}

func _testCodeStartsThenReturnsError(t *testing.T) {
	// test that a started binary can have its return
	assert := assert.New(t)
	state.InitState()
	// reset code state from any other tests
	if err := os.Symlink("/bin/bash", "/hpcaas/code/bash"); err != nil {
		t.Error(err)
	}
	defer os.Remove("/hpcaas/code/bash")
	state.SetCodeName("bash")
	state.SetCodeArguments([]string{"-c \"sleep 1 && exit 1\""})
	assert.Equal(state.CODE_WAITING, state.GetCodeState())
	err := ExecuteCode()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(state.CODE_RUNNING, state.GetCodeState())
	// wait till it errors out
	time.Sleep(1 * time.Second)
	assert.Equal(state.CODE_ERROR, state.GetCodeState())
}

func _testCodeFailToStart(t *testing.T) {
	// test that a started binary can have its return
	assert := assert.New(t)
	state.InitState()
	// reset code state from any other tests
	if err := os.Symlink("/bin/bash", "/hpcaas/code/bash"); err != nil {
		t.Error(err)
	}
	defer os.Remove("/hpcaas/code/bash")
	state.SetCodeName("bash")
	// make binary unusable
	os.Chmod("/hpcaas/code/bash", 0x000)
	// make it usable again after the test has finished
	defer os.Chmod("/hpcaas/code/bash", 0x755)
	err := ExecuteCode()
	assert.Error(err)
	assert.Equal(state.CODE_FAILED_TO_START, state.GetCodeState())
}

// test that an externally started binary can be managed
func _testCodeStartedExternally(t *testing.T) {
	// test that a started binary can have its return
	assert := assert.New(t)
	state.InitState()
	// start a command that will disown and be inherited by the root process
	cmd := exec.Command("bash", "-c", "sleep 3 & disown")
	var out bytes.Buffer
	var err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &err
	cmd.Start()
	state.SetCodeName("sleep")
	time.Sleep(2 * time.Second)
	// the daemon should pick up that there is a sleep command running
	assert.Equal(state.CODE_RUNNING, state.GetCodeState())
	time.Sleep(4 * time.Second)
	// the daemon should pick up that the sleep command has stopped
	assert.Equal(state.CODE_STOPPED, state.GetCodeState())
}

func init() {
	if _, found := os.LookupEnv(TEST_ENV_VAR); found != true {
		panic("All tests should be run in the container, with " + TEST_ENV_VAR + " being set")
	}
}
