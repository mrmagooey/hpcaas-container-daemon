package container

import "io/ioutil"
import "bytes"
import "fmt"
import "strings"

type configRequest struct {
	Data       map[int]string
	ReturnChan chan error
}

type keyRequest struct {
	Data       string
	ReturnChan chan error
}

var sshConfigChan = make(chan configRequest)
var sshPubChan = make(chan keyRequest)
var sshPrivChan = make(chan keyRequest)

var sshConfigFileLocation = "/root/.ssh/config"
var sshPrivateKeyLocation = "/root/.ssh/private_key"
var sshAuthorizedKeys = "/root/.ssh/authorized_keys"

// writeSSHConfig is intended to be used as a goroutine and writes ssh config to the filesystem
func writeSSHConfig() {
	for {
		req := <-sshConfigChan
		sshAddresses := req.Data
		responseChan := req.ReturnChan
		var buffer bytes.Buffer
		// config file preamble
		buffer.WriteString("Host *\n")
		buffer.WriteString(fmt.Sprintf("IdentityFile %s\n", sshPrivateKeyLocation))
		// this stops the interactive ssh prompt
		buffer.WriteString("StrictHostKeyChecking No\n")
		// add each containers ip
		for containerId, address := range sshAddresses {
			host := fmt.Sprintf("Host container_%d\n", containerId)
			buffer.WriteString(host)
			s := strings.Split(address, ":")
			ipAddr, port := s[0], s[1]
			hostNameString := fmt.Sprintf("    Hostname %s\n", ipAddr)
			buffer.WriteString(hostNameString)
			portString := fmt.Sprintf("    Port %s\n\n", port)
			buffer.WriteString(portString)
		}
		err := ioutil.WriteFile(sshConfigFileLocation, buffer.Bytes(), 0700)
		responseChan <- err
	}
}

// writePublicKey is intended to be used as a goroutine and writes public key information to the filesystem
func writePublicKey() {
	for {
		req := <-sshPubChan
		publicKey := req.Data
		responseChan := req.ReturnChan
		err := ioutil.WriteFile(sshAuthorizedKeys, []byte(publicKey), 0644)
		responseChan <- err
	}
}

// writePrivateKey is intended to be used as a goroutine and writes private key information to the filesystem
func writePrivateKey() {
	for {
		req := <-sshPrivChan
		privateKey := req.Data
		responseChan := req.ReturnChan
		err := ioutil.WriteFile(sshPrivateKeyLocation, []byte(privateKey), 0600)
		responseChan <- err
	}
}

// writesshconfig is the external interface to write ssh config information
func WriteSSHConfig(addrs map[int]string) error {
	var returnChan = make(chan error)
	req := configRequest{addrs, returnChan}
	sshConfigChan <- req
	return <-returnChan
}

// WriteSSHPublicKey is the external interface to write ssh config information
func WriteSSHPublicKey(key string) error {
	var returnChan = make(chan error)
	req := keyRequest{key, returnChan}
	sshPubChan <- req
	return <-returnChan
}

// WriteSSHPrivateKey is the external interface to write ssh config information
func WriteSSHPrivateKey(key string) error {
	var returnChan = make(chan error)
	req := keyRequest{key, returnChan}
	sshPrivChan <- req
	return <-returnChan
}

// starts the write goroutines up
func init() {
	go writePrivateKey()
	go writePublicKey()
	go writeSSHConfig()
}
