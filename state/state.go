package state

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/imdario/mergo"
	"github.com/mrmagooey/hpcaas-common"
)

var stateFile = "/hpcaas/daemon/state.json"

// set defaults
var daemonState = common.DaemonState{}
var stateRWMutex = sync.RWMutex{}

// GetDaemonState return a copy of the daemon state
func GetDaemonState() common.DaemonState {
	return daemonState
}

// SetDaemonState takes a daemon state and overrides the daemons state
func SetDaemonState(newState common.DaemonState) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	mergo.MergeWithOverwrite(daemonState, newState)
	daemonState = newState
	// save to disk
	go dehydrateToDisk()
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
// best-effort attempt, if the file is bad or missing this function will not complain
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
	daemonState.CodeName = &name
	go dehydrateToDisk()
}

// GetCodeName safely sets codeState
func GetCodeName() (name string, exists bool) {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	if daemonState.CodeName != nil {
		return *daemonState.CodeName, true
	}
	return "", false
}

// SetCodeStatus safely sets codeState
func SetCodeStatus(codeStatus common.CodeStatus) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeStatus = &codeStatus
	go dehydrateToDisk()
}

// GetCodeStatus returns codeState
func GetCodeStatus() (common.CodeStatus, bool) {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	if daemonState.CodeStatus != nil {
		return *daemonState.CodeStatus, true
	}
	return common.CodeStatus(0), false
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
	merged := mergeCodeParams(*daemonState.CodeParams, params)
	daemonState.CodeParams = &merged
	go dehydrateToDisk()
	return nil
}

// SetCodeParams overwrite all params with new params
func SetCodeParams(params map[string]string) error {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeParams = &params
	go dehydrateToDisk()
	return nil
}

// GetCodeParams return codeParams
func GetCodeParams() (map[string]string, bool) {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	if daemonState.CodeParams != nil {
		return *daemonState.CodeParams, true
	}
	return nil, false
}

// SetSSHAddresses set ssh addresses
func SetSSHAddresses(addrs common.ContainerAddresses) error {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.SSHAddresses = &addrs
	go dehydrateToDisk()
	return nil
}

// GetSSHAddresses return sshAddresses
func GetSSHAddresses() (common.ContainerAddresses, bool) {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	if daemonState.SSHAddresses != nil {
		return *daemonState.SSHAddresses, true
	}
	return nil, false
}

// SetAuthorizationKey sets auth key
func SetAuthorizationKey(key string) error {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.AuthorizationKey = &key
	go dehydrateToDisk()
	return nil
}

// GetAuthorizationKey gets auth key
func GetAuthorizationKey() (string, bool) {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	if daemonState.AuthorizationKey != nil {
		return *daemonState.AuthorizationKey, true
	}
	return "", false
}

// SetCodeArguments set code arguments
// these the passed to the user executable on startup
func SetCodeArguments(args []string) error {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeArguments = &args
	go dehydrateToDisk()
	return nil
}

// GetCodeArguments return code arguments
func GetCodeArguments() ([]string, bool) {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	if daemonState.CodeArguments != nil {
		return *daemonState.CodeArguments, true
	}
	return nil, false
}

// SetCodeStdout set the stdout of the user code
func SetCodeStdout(stdout string) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeStdout = &stdout
}

// GetCodeStdout get the stdout of the code
func GetCodeStdout() (string, bool) {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	if daemonState.CodeStdout != nil {
		return *daemonState.CodeStdout, true
	}
	return "", false
}

// SetCodeStderr set the stderr of the user code
func SetCodeStderr(stderr string) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeStderr = &stderr
}

// GetCodeStderr get the stderr of the user code
func GetCodeStderr() (string, bool) {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	if daemonState.CodeStderr != nil {
		return *daemonState.CodeStderr, true
	}
	return "", false
}

// SetCodePID set the user code PID
func SetCodePID(pid int) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodePID = &pid
}

// GetCodePID get the user code PID
func GetCodePID() (int, bool) {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	if daemonState.CodePID != nil {
		return *daemonState.CodePID, true
	}
	return 0, false
}

// SetCodeStartedMethod set the user code start method
func SetCodeStartedMethod(method common.StartedStatus) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.CodeStartedStatus = &method
}

// GetCodeStartedMethod get the user code start method
func GetCodeStartedMethod() (common.StartedStatus, bool) {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	if daemonState.CodeStartedStatus != nil {
		return *daemonState.CodeStartedStatus, true
	}
	return common.StartedStatus(0), false
}

// SetSSHPrivateKey set the private key
func SetSSHPrivateKey(priv string) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.SSHPrivateKey = &priv
}

// GetSSHPrivateKey get the ssh private key
func GetSSHPrivateKey() (string, bool) {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	if daemonState.SSHPrivateKey != nil {
		return *daemonState.SSHPrivateKey, true
	}
	return "", false
}

// SetSSHPublicKey set the public key
func SetSSHPublicKey(priv string) {
	stateRWMutex.Lock()
	defer stateRWMutex.Unlock()
	daemonState.SSHPublicKey = &priv
}

// GetSSHPublicKey get the ssh public key
func GetSSHPublicKey() (string, bool) {
	stateRWMutex.RLock()
	defer stateRWMutex.RUnlock()
	if daemonState.SSHPublicKey != nil {
		return *daemonState.SSHPublicKey, true
	}
	return "", false

}
