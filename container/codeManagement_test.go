package container

import "testing"
import "os"
import "github.com/stretchr/testify/assert"
import "time"
import "github.com/mrmagooey/hpcaas-container-daemon/state"
import "errors"
import "fmt"

var TEST_ENV_VAR = "HPCAAS_DAEMON_TEST_CONTAINER"

func TestParent(t *testing.T) {
	fmt.Println("")
	t.Run("", _testExecuteLs)
	t.Run("", _testStdout)
	t.Run("", _testExecuteSleep)
	t.Run("", _testCodeAlreadyStarted)
	t.Run("", _testCodeMissing)
	t.Run("", _testEnvVars)
	t.Run("", _testKillCode)
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
	state.SetCodeParams(map[string]interface{}{
		"hello": "world",
	})
	if err := ExecuteCode(); err != nil {
		t.Error(err)
	}
	time.Sleep(50 * time.Millisecond)
	assert.Equal(state.CODE_STOPPED, state.GetCodeState())
	assert.Equal("hello=world\n", state.GetCodeStdout())
}

// test that we can give environment variables to our binaries
func _testKillCode(t *testing.T) {
	// test that a started binary can have its return
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
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
	}
	// wait till it starts
	time.Sleep(100 * time.Millisecond)
	assert.Equal(state.CODE_RUNNING, state.GetCodeState())
	fmt.Println("before KillCode")
	if err := KillCode(); err != nil {
		t.Error(err)
		return
	}
	// wait till it dies
	time.Sleep(100 * time.Millisecond)

	assert.Equal(state.CODE_KILLED, state.GetCodeState())
	fmt.Println("testKillCodeFinished")
}

func init() {
	if _, found := os.LookupEnv(TEST_ENV_VAR); found != true {
		panic("All tests should be run in the container, with " + TEST_ENV_VAR + " being set")
	}
}
