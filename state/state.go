package state

import "sync"
import "encoding/json"
import "io/ioutil"
import "fmt"

var stateFile = "/hpcaas/daemon/state.json"

// ContainerAddresses maps the container id to a string in the format ip<port>
type ContainerAddresses map[int]string

// and what the address:port is for every other container
type extraPort struct {
	InternalPort           int
	Name                   string
	ExternalContainerPorts ContainerAddresses
}

// The set of possible daemon states
const (
	DaemonStartedState uint8 = iota + 1
	DaemonRunningState
	DaemonErrorState
)

// CodeState new type so that we can add our methods
type CodeState int

// The set of user code states
const (
	CodeWaitingState CodeState = iota + 1
	CodeMissingState
	CodeFailedToStartState
	CodeRunningState
	CodeStoppedState
	CodeKilledState
	CodeFailedToKillState
	CodeErrorState
)

// CodeStates A slice containing all codeStates
var codeStates = []string{
	"CodeWaitingState",
	"CodeMissingState",
	"CodeFailedToStartState",
	"CodeRunningState",
	"CodeStoppedState",
	"CodeKilledState",
	"CodeFailedToKillState",
	"CodeErrorState",
}

func (cs CodeState) String() string {
	return codeStates[int(cs)-1]
}

// The set of result states
const (
	ResultWaitingState uint8 = iota + 1
	ResultUploadingState
	ResultsUploadingFinishedState
	ResultErrorState
)

// The set of user code start states
const (
	StartedByDaemonState uint8 = iota + 1
	StartedExternallyState
)

type stateStruct struct {
	CodeParams        map[string]string  `json:"codeParams"`
	SharedFileSystem  bool               `json:"sharedFileSystem"`
	ExtraPorts        []extraPort        `json:"extraPorts"`
	CodeName          string             `json:"codeName"`
	CodeArguments     []string           `json:"codeArguments"`
	CodeState         CodeState          `json:"codeState"`
	DaemonState       uint8              `json:"daemonState"`
	ResultState       uint8              `json:"resultState"`
	SSHAddresses      ContainerAddresses `json:"sshAddresses"`
	WorldRank         int                `json:"worldRank"`
	WorldSize         int                `json:"worldSize"`
	ResultsDirectory  string             `json:"resultsDirectory"`
	ResultsURL        string             `json:"resultsUrl"`
	CodeExitStatus    int                `json:"codeExitStatus"`
	AuthorizationKey  string             `json:"authorizationKey"`
	CodeStdout        string
	CodeStderr        string
	ErrorMessages     []string
	CodePID           int
	CodeStartedMethod uint8
	SSHPrivateKey     string
	SSHPublicKey      string
}

// set defaults
var daemonState = stateStruct{}
var stateRWMutex = sync.RWMutex{}

// InitState resets the state of the daemonState to a set of default values
// used for setting state on daemon startup
func InitState() {
	stateRWMutex.Lock()
	daemonState = stateStruct{
		SharedFileSystem: false,
		CodeName:         "hpc-code",
		CodeState:        CodeWaitingState,
		DaemonState:      DaemonStartedState,
		ResultState:      ResultWaitingState,
		ResultsDirectory: "/hpcaas/results",
	}
	stateRWMutex.Unlock()
}

func init() {
	InitState()
}

// GetStateJSON return current state as json
func GetStateJSON() []byte {
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
	err := ioutil.WriteFile(stateFile, GetStateJSON(), 0777)
	if err != nil {
		// TODO
		fmt.Println("Couldn't write state to disk")
	}
}

// RehydrateFromDisk reads from the state.json file on disk and recreates the internal daemonState of the daemon
// used as a recovery strategy if the daemon has been killed or crashed
// best-effort attempt, if the file is bad or missing this function not complain
func RehydrateFromDisk() {
	file, err := ioutil.ReadFile(stateFile)
	if err != nil {
		// the file doesn't exist or is unreadable
		// could happen if the daemon previously started but didn't manage to write any daemonState
		fmt.Println("Couldn't read state from disk")
	}
	if e := json.Unmarshal(file, &daemonState); e != nil {
		// TODO the file has been corrupted
		fmt.Println("Couldn't read state from disk")
	}
}

// SetCodeName safely sets codeName
func SetCodeName(name string) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeName = name
	go dehydrateToDisk()
}

// GetCodeName safely sets codeState
func GetCodeName() string {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.CodeName
}

// SetCodeState safely sets codeState
func SetCodeState(codeState CodeState) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeState = codeState
	go dehydrateToDisk()
}

// GetCodeState returns codeState
func GetCodeState() CodeState {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.CodeState
}

// merge two map[string]interface{}'s
// second argument overrides the first
func mergeCodeParams(original map[string]string, second map[string]string) map[string]string {
	updated := make(map[string]string)
	for k, v := range original {
		updated[k] = v
	}
	for k, v := range second {
		updated[k] = v
	}
	return updated
}

// UpdateCodeParams merge new params with existing params, overwriting as necessary
func UpdateCodeParams(params map[string]string) error {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeParams = mergeCodeParams(daemonState.CodeParams, params)
	go dehydrateToDisk()
	return nil
}

// SetCodeParams overwrite all params with new params
func SetCodeParams(params map[string]string) error {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeParams = params
	go dehydrateToDisk()
	return nil
}

// GetCodeParams return codeParams
func GetCodeParams() map[string]string {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.CodeParams
}

// SetSSHAddresses set ssh addresses
func SetSSHAddresses(addrs ContainerAddresses) error {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.SSHAddresses = addrs
	go dehydrateToDisk()
	return nil
}

// GetSSHAddresses return sshAddresses
func GetSSHAddresses() ContainerAddresses {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.SSHAddresses
}

// SetAuthorizationKey sets auth key
func SetAuthorizationKey(key string) error {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.AuthorizationKey = key
	go dehydrateToDisk()
	return nil
}

// SetCodeArguments set code arguments
// these the passed to the user executable on startup
func SetCodeArguments(args []string) error {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeArguments = args
	go dehydrateToDisk()
	return nil
}

// GetCodeArguments return code arguments
func GetCodeArguments() []string {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.CodeArguments
}

// SetCodeStdout set the stdout of the user code
func SetCodeStdout(stdout string) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeStdout = stdout
}

// GetCodeStdout get the stdout of the code
func GetCodeStdout() string {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.CodeStdout
}

// SetCodeStderr set the stderr of the user code
func SetCodeStderr(stderr string) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeStderr = stderr
}

// GetCodeStderr get the stderr of the user code
func GetCodeStderr() string {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.CodeStderr
}

// AddErrorMessage append an error message to the general error log
func AddErrorMessage(err string) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.ErrorMessages = append(daemonState.ErrorMessages, err)
}

// GetErrorMessages returns the error log
func GetErrorMessages() []string {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.ErrorMessages
}

// SetCodePID set the user code PID
func SetCodePID(pid int) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodePID = pid
}

// GetCodePID get the user code PID
func GetCodePID() int {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.CodePID
}

// SetCodeStartedMethod set the user code start method
func SetCodeStartedMethod(method uint8) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeStartedMethod = method
}

// GetCodeStartedMethod get the user code start method
func GetCodeStartedMethod() uint8 {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.CodeStartedMethod
}

// SetSSHPrivateKey set the private key
func SetSSHPrivateKey(priv string) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.SSHPrivateKey = priv
}

// GetSSHPrivateKey get the ssh private key
func GetSSHPrivateKey() string {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.SSHPrivateKey
}

// SetSSHPublicKey set the public key
func SetSSHPublicKey(priv string) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.SSHPublicKey = priv
}

// GetSSHPublicKey get the ssh public key
func GetSSHPublicKey() string {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	return daemonState.SSHPublicKey
}
