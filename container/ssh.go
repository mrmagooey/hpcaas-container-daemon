package container

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/mrmagooey/hpcaas-container-daemon/state"
)

var sshConfigFileLocation = "/root/.ssh/config"
var sshPrivateKeyLocation = "/root/.ssh/private_key"
var sshAuthorizedKeys = "/root/.ssh/authorized_keys"

var writeConfigMut = sync.Mutex{}

// WriteSSHConfig is intended to be used as a goroutine and writes ssh config to the filesystem
func WriteSSHConfig() error {
	writeConfigMut.Lock()
	defer writeConfigMut.Unlock()
	sshAddresses, ok := state.GetSSHAddresses()
	if !ok {
		return errors.New("No SSH Addresses, cannot write ssh config")
	}
	var buffer bytes.Buffer
	// config file preamble
	buffer.WriteString("Host *\n")
	buffer.WriteString(fmt.Sprintf("IdentityFile %s\n", sshPrivateKeyLocation))
	// this stops the interactive ssh prompt
	buffer.WriteString("StrictHostKeyChecking No\n")
	// add each containers ip
	for containerID, address := range sshAddresses {
		host := fmt.Sprintf("Host %s\n", generateContainerName(containerID))
		buffer.WriteString(host)
		s := strings.Split(address, ":")
		ipAddr, port := s[0], s[1]
		hostNameString := fmt.Sprintf("    Hostname %s\n", ipAddr)
		buffer.WriteString(hostNameString)
		portString := fmt.Sprintf("    Port %s\n\n", port)
		buffer.WriteString(portString)
	}
	err := ioutil.WriteFile(sshConfigFileLocation, buffer.Bytes(), 0700)
	return err
}

var writePubKeyMut = sync.Mutex{}

// WritePublicKey writes pub key to authorized keys
func WritePublicKey() error {
	writePubKeyMut.Lock()
	defer writePubKeyMut.Unlock()
	publicKey, ok := state.GetSSHPublicKey()
	if !ok {
		return errors.New("No public key in state")
	}
	err := ioutil.WriteFile(sshAuthorizedKeys, []byte(publicKey), 0644)
	return err
}

var writePrivKeyMut = sync.Mutex{}

// WritePrivateKey writes private key to disk
func WritePrivateKey() error {
	writePrivKeyMut.Lock()
	defer writePrivKeyMut.Unlock()
	privateKey, ok := state.GetSSHPrivateKey()
	if !ok {
		return errors.New("No private key in state")
	}
	err := ioutil.WriteFile(sshPrivateKeyLocation, []byte(privateKey), 0600)
	return err
}
