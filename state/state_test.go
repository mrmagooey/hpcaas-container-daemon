package state

import "testing"
import "github.com/stretchr/testify/assert"
import "os"
import "io/ioutil"
import "time"

var codeName = "ls"
var codeState = CodeMissingState

var codeArgs = []string{"hi", "world"}

var params = map[string]string{
	"blah":  "1",
	"stuff": "2",
}

var sshAddrs = ContainerAddresses{
	1: "255.255.255.255:8000",
	2: "255.255.255.255:88",
}

func TestGetAndSetState(t *testing.T) {
	InitState()
	assert := assert.New(t)
	SetCodeState(codeState)
	SetCodeName(codeName)
	SetCodeParams(params)
	SetSSHAddresses(sshAddrs)
	SetCodeArguments(codeArgs)
	assert.Equal(codeState, GetCodeState())
	assert.Equal(codeName, GetCodeName())
	assert.Equal(params, GetCodeParams())
	assert.Equal(sshAddrs, GetSSHAddresses())
	assert.Equal(codeArgs, GetCodeArguments())
}

func TestHydration(t *testing.T) {
	InitState()
	assert := assert.New(t)
	// state file exists
	_, err := os.Stat(stateFile)
	assert.NoError(err)
	// check contents of file
	SetAuthorizationKey("lol")
	time.Sleep(10 * time.Millisecond)
	contents, err := ioutil.ReadFile(stateFile)
	assert.Contains(string(contents), "\"authorizationKey\":\"lol\"")
}
