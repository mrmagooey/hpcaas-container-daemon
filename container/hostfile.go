package container

import "io/ioutil"
import "sync"
import "github.com/mrmagooey/hpcaas-container-daemon/state"
import "bytes"
import "fmt"

var hostFilePath = "/hpcaas/runtime/hostfile"

var hostFileMutex = sync.Mutex{}

// WriteHostFile write the MPI style hostfile to disk
// pull information from the state
func WriteHostFile() error {
	hostFileMutex.Lock()
	defer hostFileMutex.Unlock()
	addrs := state.GetSSHAddresses()
	var buf bytes.Buffer
	for id := range addrs {
		hostfileEntry := fmt.Sprintf("%s slots 1\n", generateContainerName(id))
		buf.WriteString(hostfileEntry)
	}
	err := ioutil.WriteFile(hostFilePath, buf.Bytes(), 0x777)
	return err
}
