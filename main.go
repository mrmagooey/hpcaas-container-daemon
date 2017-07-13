package main

import (
	"github.com/mrmagooey/hpcaas-container-daemon/state"
	"io/ioutil"
	"net/http"
	"os"
)

var STARTUP_FILE = "/hpcaas/daemon/daemon_has_started"
var TLS_CERT_FILE = "/hpcaas/daemon/tls_server.crt"
var TLS_KEY_FILE = "/hpcaas/daemon/tls_server.key"
var AUTHORIZATION = "/hpcaas/daemon/authorization"

// run once at container startup
// pull comm information out of environment variables and save to disk
func setupTLSInfo() {
	tls_public_cert, envErr := os.LookupEnv("TLS_PUBLIC_CERT")
	if envErr == false {
		panic("TLS certificate is missing from environment variables")
	}
	tls_private_key, envErr := os.LookupEnv("TLS_PRIVATE_KEY")
	if envErr == false {
		panic("TLS key is missing from environment variables")
	}
	auth_key, envErr := os.LookupEnv("AUTHORIZATION")
	if envErr == false {
		panic("authorization is missing from environment variables")
	}
	err := ioutil.WriteFile(TLS_CERT_FILE, []byte(tls_public_cert), 0300)
	if err != nil {
		panic("Couldn't save tls server certificate to disk")
	}
	err = ioutil.WriteFile(TLS_KEY_FILE, []byte(tls_private_key), 0300)
	if err != nil {
		panic("Couldn't save tls server key to disk")
	}
	state.SetAuthorizationKey(auth_key)
}

// check if this is the first time that the daemon has started up
// if the daemon was previously running
func startup() {
	if _, err := os.Stat(STARTUP_FILE); err != nil {
		// startup file doesn't exist, this is the first time the daemon has started
		f, err := os.Create(STARTUP_FILE)
		if err != nil {
			panic("Couldn't write startup file")
		}
		err = f.Close()
		if err != nil {
			panic("Couldn't close startup file")
		}
	} else {
		// daemon has already started previously, rehydrate state from disk
		state.RehydrateFromDisk()
	}
	// probably safer to do the tls info again regardless of previous startup
	setupTLSInfo()
}

func main() {
	startup()
	r := register_routes()
	http.ListenAndServeTLS(":443", TLS_CERT_FILE, TLS_KEY_FILE, r)
}
