package state

import "sync"
import "encoding/json"
import "io/ioutil"

// import "os/exec"
// import "reflect"

var stateFile = "/hpcaas/daemon/state.json"

// contains a name of the port, what the container is binding today
type ContainerAddresses map[int]string

// and what the address:port is for every other container
type extraPort struct {
	InternalPort           int
	Name                   string
	ExternalContainerPorts ContainerAddresses
}

const (
	DAEMON_STARTED uint8 = iota + 1
	DAEMON_RUNNING
	DAEMON_ERROR
)

const (
	CODE_WAITING uint8 = iota + 1
	CODE_MISSING
	CODE_FAILED_TO_START
	CODE_RUNNING
	CODE_STOPPED
	CODE_KILLED
	CODE_FAILED_TO_KILL
	CODE_ERROR
)

const (
	RESULT_WAITING uint8 = iota + 1
	RESULT_UPLOADING
	RESULT_STOPPED
	RESULT_ERROR
)

type StateStruct struct {
	CodeParams       map[string]interface{} `json:"codeParams"`
	SharedFileSystem bool                   `json:"sharedFileSystem"`
	ExtraPorts       []extraPort            `json:"extraPorts"`
	CodeName         string                 `json:"codeName"`
	CodeArguments    []string               `json:"codeArguments"`
	CodeState        uint8                  `json:"codeState"`
	DaemonState      uint8                  `json:"daemonState"`
	ResultState      uint8                  `json:"resultState"`
	SSHAddresses     ContainerAddresses     `json:"sshAddresses"`
	WorldRank        int                    `json:"worldRank"`
	WorldSize        int                    `json:"worldSize"`
	ResultsDirectory string                 `json:"resultsDirectory"`
	ResultsUrl       string                 `json:"resultsUrl"`
	CodeExitStatus   int                    `json:"codeExitStatus"`
	AuthorizationKey string                 `json:"authorizationKey"`
	CodeStdout       string
	CodeStderr       string
	ErrorMessages    []string
	CodePID          int
}

// set defaults
var daemonState = StateStruct{}
var stateRWMutex = sync.RWMutex{}

// resets the state of the daemonState to a set of default values
func InitState() {
	stateRWMutex.Lock()
	daemonState = StateStruct{
		SharedFileSystem: false,
		CodeName:         "hpc-code",
		CodeState:        CODE_WAITING,
		DaemonState:      DAEMON_STARTED,
		ResultState:      RESULT_WAITING,
		ResultsDirectory: "/hpcaas/results",
	}
	stateRWMutex.Unlock()
}

func init() {
	InitState()
}

func GetStateJson() []byte {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	sj, _ := json.Marshal(daemonState)
	return sj
}

var dehydrateMut = sync.Mutex{}

// saves the daemonState to disk, is thread-safe
// used as a recovery strategy if the daemon has been killed or crashed
func dehydrateToDisk() {
	dehydrateMut.Lock()
	defer dehydrateMut.Unlock()
	err := ioutil.WriteFile(stateFile, GetStateJson(), 777)
	if err != nil {
		panic("Can't write dehydrate file to disk")
	}
}

// reads from the state.json file on disk and recreates the internal daemonState of the daemon
// used as a recovery strategy if the daemon has been killed or crashed
// best-effort attempt, if the file is bad or missing this function not complain
func RehydrateFromDisk() {
	file, err := ioutil.ReadFile(stateFile)
	if err != nil {
		// the file doesn't exist or is unreadable
		// could happen if the daemon previously started but didn't manage to write any daemonState
	}
	if e := json.Unmarshal(file, daemonState); e != nil {
		// TODO the file has been corrupted
	}
}

func SetCodeName(name string) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeName = name
	go dehydrateToDisk()
}

func GetCodeName() string {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.CodeName
}

func SetCodeState(codeState uint8) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeState = codeState
	go dehydrateToDisk()
}

func GetCodeState() uint8 {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.CodeState
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

// merge new params with existing params, overwriting as necessary
func UpdateCodeParams(params map[string]interface{}) error {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeParams = mergeCodeParams(daemonState.CodeParams, params)
	go dehydrateToDisk()
	return nil
}

// overwrite all params with new params
func SetCodeParams(params map[string]interface{}) error {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeParams = params
	go dehydrateToDisk()
	return nil
}

func GetCodeParams() map[string]interface{} {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.CodeParams
}

func SetSSHAddresses(addrs ContainerAddresses) error {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.SSHAddresses = addrs
	go dehydrateToDisk()
	return nil
}

func GetSSHAddresses() ContainerAddresses {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.SSHAddresses
}

func SetAuthorizationKey(key string) error {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.AuthorizationKey = key
	go dehydrateToDisk()
	return nil
}

func SetCodeKey(key string) error {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.AuthorizationKey = key
	go dehydrateToDisk()
	return nil
}

func SetCodeArguments(args []string) error {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeArguments = args
	go dehydrateToDisk()
	return nil
}

func GetCodeArguments() []string {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.CodeArguments
}

func SetCodeStdout(stdout string) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeStdout = stdout
}

func GetCodeStdout() string {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.CodeStdout
}

func SetCodeStderr(stderr string) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeStderr = stderr
}

func GetCodeStderr() string {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.CodeStderr
}

func AddErrorMessage(err string) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.ErrorMessages = append(daemonState.ErrorMessages, err)
}

func GetErrorMessages() []string {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.ErrorMessages
}

func SetCodePID(pid int) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodePID = pid
}

func GetCodePID() int {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.CodePID
}
