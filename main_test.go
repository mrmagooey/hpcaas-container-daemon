package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/mrmagooey/hpcaas-container-daemon/state"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"testing"
	"time"
)

// https://golang.org/src/crypto/tls/generate_cert.go
func generateCACertAndKey() (certBytes []byte, keyBytes []byte, err error) {
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	notBefore := time.Now()
	notAfter := notBefore.AddDate(10, 0, 0)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, err
	}
	caTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"HPCAAS"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	// caTemplate.DNSNames = append(caTemplate.DNSNames, "no-domain-here.example.com")
	caTemplate.IsCA = true
	caTemplate.KeyUsage |= x509.KeyUsageCertSign

	caCertBytes, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, nil, err
	}
	caPemBlock := &pem.Block{Type: "CERTIFICATE", Bytes: caCertBytes}
	privPemBlock := &pem.Block{Type: "RSA PRIVATE KEYS", Bytes: x509.MarshalPKCS1PrivateKey(caPrivateKey)}
	certBuf := &bytes.Buffer{}
	keyBuf := &bytes.Buffer{}
	err = pem.Encode(certBuf, caPemBlock)
	if err != nil {
		return nil, nil, err
	}

	err = pem.Encode(keyBuf, privPemBlock)
	if err != nil {
		return nil, nil, err
	}

	return certBuf.Bytes(), keyBuf.Bytes(), nil
}

func generateNewCert(caCertBytes []byte, caKeyBytes []byte) ([]byte, []byte, error) {
	// regenerate ca info from byte buffers, pretending that we've loaded files
	caCertPemBlock, _ := pem.Decode(caCertBytes)
	if caCertPemBlock == nil {
		return nil, nil, errors.New("failed to decode caCertBytes")
	}

	caCert, err := x509.ParseCertificate(caCertPemBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	caKeyPemBlock, _ := pem.Decode(caKeyBytes)
	if caKeyPemBlock == nil {
		return nil, nil, errors.New("failed to decode caKeyBytes ")
	}

	caKey, err := x509.ParsePKCS1PrivateKey(caKeyPemBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	// generate new cert
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	notBefore := time.Now()
	notAfter := notBefore.AddDate(10, 0, 0)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"HPCAAS"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	template.DNSNames = append(template.DNSNames, "no-domain-here.example.com")
	template.IsCA = false

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, &privateKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	pemBlock := &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}
	privPemBlock := &pem.Block{Type: "PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}

	certBuf := &bytes.Buffer{}
	keyBuf := &bytes.Buffer{}

	pem.Encode(certBuf, pemBlock)
	pem.Encode(keyBuf, privPemBlock)

	return certBuf.Bytes(), keyBuf.Bytes(), nil
}

func generateAuthKey() string {
	numberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	authInt, _ := rand.Int(rand.Reader, numberLimit)
	return fmt.Sprintf("%x", authInt)
}

// generate the better part of a x509 cert signing chain
// test that the daemon can pick up a correctly signed x509 certificate
// from the environment variables
func TestTLSInformation(t *testing.T) {
	assert := assert.New(t)
	state.InitState()
	// generate tls information
	caCertBytes, caKeyBytes, err := generateCACertAndKey()
	if err != nil {
		t.Error(err)
		return
	}
	daemonCertBytes, daemonKeyBytes, err := generateNewCert(caCertBytes, caKeyBytes)
	if err != nil {
		t.Error(err)
		return
	}
	// generate auth key
	authKey := generateAuthKey()
	// set the env variables
	os.Setenv(daemonPublicCertEnvVar, string(daemonCertBytes))
	os.Setenv(daemonPrivateKeyEnvVar, string(daemonKeyBytes))
	os.Setenv(daemonAuthEnvVar, authKey)
	// setup tls info
	setupTLSInfo()
	// check that the files have been correctly written to where they need to be
	loadedCert, err := ioutil.ReadFile(tlsCertFile)
	assert.NoError(err)
	loadedKey, err := ioutil.ReadFile(tlsKeyFile)
	assert.NoError(err)

	certPemBlock, _ := pem.Decode(loadedCert)
	_, err = x509.ParseCertificate(certPemBlock.Bytes)
	assert.NoError(err)
	keyPemBlock, _ := pem.Decode(loadedKey)
	_, err = x509.ParsePKCS1PrivateKey(keyPemBlock.Bytes)
	assert.NoError(err)

}

// check that we can connect to the server, and verify that the tls cert it offers
// is in our cert pool (the "CA" cert)
func TestTLSServerStartup(t *testing.T) {
	assert := assert.New(t)
	state.InitState()
	// generate tls information
	caCertBytes, caKeyBytes, err := generateCACertAndKey()
	if err != nil {
		t.Error(err)
		return
	}
	daemonCertBytes, daemonKeyBytes, err := generateNewCert(caCertBytes, caKeyBytes)
	if err != nil {
		t.Error(err)
		return
	}
	// generate auth key
	authKey := generateAuthKey()
	// set the env variables
	os.Setenv(daemonPublicCertEnvVar, string(daemonCertBytes))
	os.Setenv(daemonPrivateKeyEnvVar, string(daemonKeyBytes))
	os.Setenv(daemonAuthEnvVar, authKey)
	// setup tls info
	setupTLSInfo()
	// check that the files have been correctly written to where they need to be
	server := setupServer()
	go func() {
		e := server.ListenAndServeTLS(tlsCertFile, tlsKeyFile)
		fmt.Println(e)
	}()
	// wait for goroutine to startup the server
	time.Sleep(50 * time.Millisecond)
	// setup client to make requests
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(caCertBytes)
	if !ok {
		t.Error("couldn't append pem to root cert pool")
		return
	}

	var tr http.RoundTripper = &http.Transport{
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := http.Client{Transport: tr}
	res, err := client.Get("https://127.0.0.1:443/v1/heartbeat/")
	if err != nil {
		t.Error(err)
		return
	}
	// get the certs that the server has presented
	certs := res.TLS.PeerCertificates
	opts := x509.VerifyOptions{
		Intermediates: x509.NewCertPool(),
		Roots:         roots,
	}
	for i, cert := range certs {
		if i == 0 {
			continue
		}
		opts.Intermediates.AddCert(cert)
	}
	_, err = certs[0].Verify(opts)
	if err != nil {
		t.Error(err)
		return
	}
	// check that the response body is correct
	assert.Equal(res.StatusCode, http.StatusOK)
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = time.Parse(time.UnixDate, string(bodyBytes))
	assert.NoError(err)
	server.Close()
}
