package state

import "sync"
import "bytes"
import "fmt"
import "io/ioutil"
import "encoding/json"

var parameterJSONPath = "/hpcaas/parameters/parameters.json"
var parameterPath = "/hpcaas/parameters/parameters"

// contains a name of the port, what the container is binding today
// and what the address:port is for every other container
type extraPort struct {
	InternalPort           int
	Name                   string
	ExternalContainerPorts map[int]string
}

const (
	CODE_WAITING uint8 = iota + 1
	CODE_MISSING
	CODE_RUNNING
	CODE_STOPPED
	CODE_KILLED
	CODE_ERROR
)

const (
	RESULT_WAITING uint8 = iota + 1
	RESULT_UPLOADING
	RESULT_STOPPED
	RESULT_ERROR
)

type StateStruct struct {
	rw               *sync.RWMutex
	CodeParams       map[string]interface{}
	SharedFileSystem bool
	ExtraPorts       []extraPort
	CodeName         string
	CodeState        uint8
	ResultState      uint8
}

var state = StateStruct{rw: &sync.RWMutex{}}

func SetCodeName(name string) {
	state.rw.Lock()
	defer state.rw.Unlock()
	state.CodeName = name
}

func GetCodeName() string {
	state.rw.RLock()
	defer state.rw.RUnlock()
	return state.CodeName
}

func SetCodeState(codeState uint8) {
	state.rw.Lock()
	defer state.rw.Unlock()
	state.CodeState = codeState
}

func GetCodeState() uint8 {
	state.rw.RLock()
	defer state.rw.RUnlock()
	return state.CodeState
}

// merge two map[string]interface{}'s
func mergeCodeParams(original map[string]interface{}, second map[string]interface{}) map[string]interface{} {
	updated := make(map[string]interface{})
	for k, v := range original {
		updated[k] = v
	}
	for k, v := range second {
		updated[k] = v
	}
	return updated
}

func SetCodeParams(params map[string]interface{}) error {
	state.rw.Lock()
	defer state.rw.Unlock()
	state.CodeParams = mergeCodeParams(state.CodeParams, params)
	err := writeCodeParams()
	if err != nil {
		return err
	}
	return nil
}

// write code param state to disk
func writeCodeParams() error {
	// write json
	newJSON, err := json.Marshal(state.CodeParams)
	if err != nil {
		return err
	}
	ioutil.WriteFile(parameterJSONPath, newJSON, 0777)
	if err != nil {
		return err
	}
	// write newline separated file
	var buffer bytes.Buffer
	for k, v := range state.CodeParams {
		envLine := fmt.Sprintf("%s=%v\n", k, v)
		buffer.WriteString(envLine)
	}
	ioutil.WriteFile(parameterPath, buffer.Bytes(), 0777)
	return nil
}

func SetContainerParams(params map[string]interface{}) error {

	return nil
}

func SetSharedFileSystem(fs bool) {

}

func SetExtraPorts(ports []extraPort) {

}
