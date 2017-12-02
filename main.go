package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mrmagooey/hpcaas-container-daemon/state"
)

// BaseDaemonDir the root of the daemon working dir
var BaseDaemonDir = "/hpcaas/daemon"

var startupFile = filepath.Join(BaseDaemonDir, "daemon_has_started")
var tlsCertFile = filepath.Join(BaseDaemonDir, "tls_server.crt")
var tlsKeyFile = filepath.Join(BaseDaemonDir, "tls_server.key")

var daemonPublicCertEnvVar = "TLS_PUBLIC_CERT"
var daemonPrivateKeyEnvVar = "TLS_PRIVATE_KEY"
var daemonAuthEnvVar = "AUTHORIZATION"

var logFileLocation = filepath.Join(BaseDaemonDir, "log.txt")

func init() {
	f, err := os.OpenFile(logFileLocation, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		fmt.Println("Failure to open log file")
	}
	log.SetOutput(f)
}

// run once at container startup
// pull comm information out of environment variables and save to disk
func setupTLSInfo() {
	tlsPublicCert, envErr := os.LookupEnv(daemonPublicCertEnvVar)
	if !envErr {
		log.Panicf("TLS certificate is missing from environment variables")
	}
	tlsPrivateKey, envErr := os.LookupEnv(daemonPrivateKeyEnvVar)
	if !envErr {
		log.Panicf("TLS key is missing from environment variables")
	}
	authKey, envErr := os.LookupEnv(daemonAuthEnvVar)
	if !envErr {
		log.Panicf("authorization is missing from environment variables")
	}
	//
	err := ioutil.WriteFile(tlsCertFile, []byte(tlsPublicCert), 0300)
	if err != nil {
		log.Panicf("Couldn't save tls server certificate to disk")
	}
	err = ioutil.WriteFile(tlsKeyFile, []byte(tlsPrivateKey), 0300)
	if err != nil {
		log.Panicf("Couldn't save tls server key to disk")
	}

	state.SetAuthorizationKey(authKey)
}

// check if this is the first time that the daemon has started up
// if the daemon was previously running
func daemonStartup() {
	if _, err := os.Stat(startupFile); err != nil {
		// startup file doesn't exist, this is the first time the daemon has started
		f, err := os.Create(startupFile)
		if err != nil {
			log.Panicf("Couldn't write startup file")
		}
		err = f.Close()
		if err != nil {
			log.Panicf("Couldn't close startup file")
		}
	} else {
		// daemon has already started previously, rehydrate state from disk
		state.RehydrateFromDisk()
	}
}

// setup the tls information for the server
func setupServer() *http.Server {
	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		}}
	routes := registerRoutes()
	server := &http.Server{
		Addr:      ":443",
		TLSConfig: tlsConfig,
		Handler:   routes,
	}
	return server
}

func main() {
	daemonStartup()
	log.Println("daemonStartup")
	setupTLSInfo()
	log.Println("TLS info retrieved")
	server := setupServer()
	log.Println("server setup")
	err := server.ListenAndServeTLS(tlsCertFile, tlsKeyFile)
	log.Println(err)
	log.Println("daemon finished")
}
