package main

import (
	"io/ioutil"
	"net/http"
	"os"
)

// run once at container startup
// pull tls information out of environment variables
func setup_tls_info() bool {
	tls_public_cert, setup_err := os.LookupEnv("TLS_PUBLIC_CERT")
	tls_private_key, setup_err := os.LookupEnv("TLS_PRIVATE_KEY")
	auth_key, setup_err := os.LookupEnv("AUTHORIZATION")
	write_err := ioutil.WriteFile("/hpcaas/daemon/tls_server.crt", []byte(tls_public_cert), 0300)
	write_err = ioutil.WriteFile("/hpcaas/daemon/tls_server.key", []byte(tls_private_key), 0300)
	write_err = ioutil.WriteFile("/hpcaas/daemon/AUTHORIZATION", []byte(auth_key), 0300)
	if write_err != nil {
		setup_err = true
	}
	return setup_err
}

func main() {
	// setup http endpoint
	r := register_routes()
	http.Handle("/", r)
}
