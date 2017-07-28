package state

import "testing"
import "github.com/stretchr/testify/assert"

func TestState(t *testing.T) {
	assert := assert.New(t)
	SetCodeState(CODE_MISSING)
	SetCodeName("ls")
	SetCodeParams(map[string]interface{}{
		"blah":  1,
		"stuff": 2,
	})
	assert.Equal(state.CodeState, CODE_MISSING)
	assert.Equal(state.CodeName, "ls")
	assert.Equal(state.CodeParams, map[string]interface{}{
		"blah":  1,
		"stuff": 2,
	})

}
