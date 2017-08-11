package container

import "io/ioutil"
import "github.com/stretchr/testify/assert"
import "testing"

func testWriteCodeParams(t *testing.T) {
	assert := assert.New(t)
	WriteCodeParams(map[string]string{
		"hello":     "world",
		"foo":       "bar",
		"something": "1",
	})
	json, err := ioutil.ReadFile(parameterJSONPath)
	assert.NoError(err)
	envs, err := ioutil.ReadFile(parameterPath)
	assert.NoError(err)
	assert.JSONEq("{\"hello\":\"world\",\"foo\":\"bar\",\"something\":1}", string(json))
	assert.Contains(envs, "hello=world\n")
	assert.Contains(envs, "foo=bar\n")
	assert.Contains(envs, "something=1\n")
}
