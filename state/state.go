package state

import "sync"

// contains a name of the port, what the container is binding today
type portMapping map[int]string

// and what the address:port is for every other container
type extraPort struct {
	InternalPort           int
	Name                   string
	ExternalContainerPorts portMapping
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
	SSHAddresses     portMapping
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
// second argument overrides the first
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
	return nil
}

func GetCodeParams() map[string]interface{} {
	state.rw.RLock()
	defer state.rw.RUnlock()
	return state.CodeParams
}

func SetContainerParams(params map[string]interface{}) error {

	return nil
}

func SetSharedFileSystem(fs bool) {

}

func SetExtraPorts(ports []extraPort) {

}

func SetSSHAddresses(addrs map[int]string) error {
	state.rw.Lock()
	defer state.rw.Unlock()
	state.SSHAddresses = addrs
	return nil
}
