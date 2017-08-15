package container

import "io/ioutil"
import "github.com/stretchr/testify/assert"
import "testing"
import "crypto/rsa"
import "crypto/x509"
import "crypto/rand"
import "encoding/pem"
import "bytes"
import "os"
import "golang.org/x/crypto/ssh"
import "github.com/mrmagooey/hpcaas-container-daemon/state"

func TestWriteSSHConfig(t *testing.T) {
	assert := assert.New(t)
	state.InitState()
	testAddrs := state.ContainerAddresses{
		1: "127.0.0.1:8000",
		2: "127.0.0.1:8001",
		3: "127.0.0.1:8002",
		4: "127.0.0.1:8003",
		5: "127.0.0.1:8004",
	}
	state.SetSSHAddresses(testAddrs)
	err := WriteSSHConfig()
	assert.NoError(err)
	conf, err := ioutil.ReadFile("/root/.ssh/config")
	confString := string(conf)
	assert.Contains(confString, "Host *\n")
	assert.Contains(confString, "Host container_1\n    Hostname 127.0.0.1\n    Port 8000")
}

func TestWriteSSHPrivateKey(t *testing.T) {
	assert := assert.New(t)
	state.InitState()
	// generate a valid private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Error(err)
	}
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	privateKeyBuf := bytes.Buffer{}
	if err := pem.Encode(&privateKeyBuf, privateKeyPEM); err != nil {
		t.Error(err)
	}
	state.SetSSHPrivateKey(privateKeyBuf.String())
	WritePrivateKey()
	conf, err := ioutil.ReadFile("/root/.ssh/private_key")
	assert.Equal(privateKeyBuf.String(), string(conf))
	os.Remove("/root/.ssh/private_key")
}

func TestWriteSSHPublicKey(t *testing.T) {
	assert := assert.New(t)
	state.InitState()
	// generate a valid private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Error(err)
	}
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	privateKeyBuf := bytes.Buffer{}
	if err := pem.Encode(&privateKeyBuf, privateKeyPEM); err != nil {
		t.Error(err)
	}

	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	state.SetSSHPublicKey(string(ssh.MarshalAuthorizedKey(publicKey)))
	WritePublicKey()
	priv, err := ioutil.ReadFile(sshAuthorizedKeys)
	assert.Equal(string(priv), string(ssh.MarshalAuthorizedKey(publicKey)))
	os.Remove(sshAuthorizedKeys)
}

// func MakeSSHKeyPair() (string, string) {
// 	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
// 	if err != nil {
// 		return err
// 	}
// 	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
// 	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
// 		return err
// 	}
// 	// generate and write public key
// 	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
// 	if err != nil {
// 		return err
// 	}
// 	return ioutil.WriteFile(pubKeyPath, ssh.MarshalAuthorizedKey(pub), 0655)
// }
