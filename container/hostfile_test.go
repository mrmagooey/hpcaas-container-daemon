package container

import "io/ioutil"
import "github.com/stretchr/testify/assert"
import "testing"

import "github.com/mrmagooey/hpcaas-container-daemon/state"

func TestWriteHostFile(t *testing.T) {
	assert := assert.New(t)
	state.SetSSHAddresses(state.ContainerAddresses{
		1: "127.0.0.1:8000",
		2: "127.0.0.1:8002",
	})
	WriteHostFile()
	contents, err := ioutil.ReadFile(hostFilePath)
	assert.NoError(err)
	assert.Equal(
		"container_1 slots 1\ncontainer_2 slots 1\n",
		string(contents),
	)
}
