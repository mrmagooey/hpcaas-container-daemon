package state

import "sync"
import "encoding/json"
import "io/ioutil"

var stateFile = "/hpcaas/daemon/state.json"

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

type url struct {
	address string
}

type StateStruct struct {
	rw               *sync.RWMutex
	CodeParams       map[string]interface{} `json:"codeParams"`
	SharedFileSystem bool                   `json:"sharedFileSystem"`
	ExtraPorts       []extraPort            `json:"extraPorts"`
	CodeName         string                 `json:"codeName"`
	CodeState        uint8                  `json:"codeState"`
	ResultState      uint8                  `json:"resultState"`
	SSHAddresses     portMapping            `json:"sshAddresses"`
	AuthorizationKey string                 `json:"authorizationKey"`
	WorldRank        int                    `json:"worldRank"`
	WorldSize        int                    `json:"worldSize"`
	ResultsDirectory string                 `json:"resultsDirectory"`
	ResultsUrl       url                    `json:"resultsUrl"`
}

var state = StateStruct{rw: &sync.RWMutex{}}

func StateJson() []byte {
	sj, _ := json.Marshal(state)
	return sj
}

var dehydrateMut = sync.Mutex{}

// saves the state to disk, is thread-safe
// used as a recovery strategy if the daemon has been killed or crashed
func dehydrateToDisk() {
	dehydrateMut.Lock()
	defer dehydrateMut.Unlock()
	err := ioutil.WriteFile(stateFile, StateJson(), 777)
	if err != nil {
		//TODO
	}
}

// reads from the state.json file on disk and recreates the internal state of the daemon
// used as a recovery strategy if the daemon has been killed or crashed
// best-effort attempt, if the file is bad or missing this function not complain
func RehydrateFromDisk() {
	file, err := ioutil.ReadFile(stateFile)
	if err != nil {
		// the file doesn't exist or is unreadable
		// could happen if the daemon previously started but didn't manage to write any state
	}
	if e := json.Unmarshal(file, state); e != nil {
		// the file has been corrupted
	}
}

func SetCodeName(name string) {
	state.rw.Lock()
	defer state.rw.Unlock()
	go dehydrateToDisk()
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
	go dehydrateToDisk()
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

func SetAuthorizationKey(key string) {
	state.rw.Lock()
	defer state.rw.Unlock()
	go dehydrateToDisk()
	state.AuthorizationKey = key
}

func SetSSHAddresses(addrs map[int]string) error {
	state.rw.Lock()
	defer state.rw.Unlock()
	go dehydrateToDisk()
	state.SSHAddresses = addrs
	return nil
}

func GetAuthorizationKey() string {
	state.rw.RLock()
	defer state.rw.RUnlock()
	return state.AuthorizationKey
}
