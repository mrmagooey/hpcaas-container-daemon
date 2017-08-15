package apiV1

import (
	"bytes"
	"fmt"
	"github.com/mrmagooey/hpcaas-container-daemon/state"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func init() {
	fmt.Println("")
}

func TestSetCodeParametersV1_POST(t *testing.T) {
	state.InitState()
	assert := assert.New(t)
	var jsonStr = []byte(`{"codeParameters":{"foo":"bar", "hello":"value", "myParam": "1"}}`)
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SetCodeParams())
	handler.ServeHTTP(rr, req)
	assert.Equal(rr.Code, http.StatusOK)
	// Check the response body is what we expect.
	expected := `{"status":"success","data":{"message":"parameter accepted"}}`
	assert.JSONEq(expected, rr.Body.String())
	// check that the internal state has been updated
	assert.Equal("bar", state.GetCodeParams()["foo"])
	assert.Equal("value", state.GetCodeParams()["hello"])
	assert.Equal("1", state.GetCodeParams()["myParam"])
}

func TestSetCodeName(t *testing.T) {
	state.InitState()
	assert := assert.New(t)
	var jsonStr = []byte(`{"codeName": "blah"}`)
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SetCodeName())
	handler.ServeHTTP(rr, req)
	assert.Equal(rr.Code, http.StatusOK)
	// Check the response body is what we expect.
	expected := `{"status":"success","data":{"message":"name accepted"}}`
	assert.JSONEq(expected, rr.Body.String())
	// check that the internal state has been updated
	assert.Equal("blah", state.GetCodeName())
}

func TestSetCodeState(t *testing.T) {
	state.InitState()
	assert := assert.New(t)
	var jsonBytes = []byte(`{"codeState": 1}`)
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(jsonBytes))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SetCodeState())
	handler.ServeHTTP(rr, req)
	assert.Equal(rr.Code, http.StatusOK)
	// Check the response body is what we expect.
	expected := `{"status":"success","data":{"message":"state accepted"}}`
	assert.JSONEq(expected, rr.Body.String())
	// check that the internal state has been updated
	assert.Equal(uint8(1), state.GetCodeState())
}

func TestCommand(t *testing.T) {
	assert := assert.New(t)
	state.InitState()
	if err := os.Symlink("/bin/sleep", "/hpcaas/code/sleep"); err != nil {
		t.Error(err)
	}
	defer os.Remove("/hpcaas/code/sleep")
	state.SetCodeName("sleep")
	state.SetCodeArguments([]string{"10"})
	var jsonBytes = []byte(`{"command": "start"}`)
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(jsonBytes))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Command())
	handler.ServeHTTP(rr, req)
	assert.Equal(rr.Code, http.StatusOK)
	// Check the response body is what we expect.
	expected := `{"status":"success","data":{"message":"code started"}}`
	assert.JSONEq(expected, rr.Body.String())
	// check that the internal state has been updated
	assert.Equal(state.CodeRunningState, state.GetCodeState())

	// kill the code
	jsonBytes = []byte(`{"command": "kill"}`)
	req, err = http.NewRequest("POST", "/", bytes.NewBuffer(jsonBytes))
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(Command())
	handler.ServeHTTP(rr, req)
	assert.Equal(rr.Code, http.StatusOK)
	// Check the response body is what we expect.
	expected = `{"status":"success","data":{"message":"code killed"}}`
	assert.JSONEq(expected, rr.Body.String())
	// check that the internal state has been updated
	assert.Equal(state.CodeKilledState, state.GetCodeState())
}

func TestSetSSHAddrs(t *testing.T) {
	state.InitState()
	assert := assert.New(t)
	var jsonBytes = []byte(`{"sshAddresses": {"1":"127.0.0.1:8230", "2": "127.0.0.1:9809"}}`)
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(jsonBytes))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SetSSHAddresses())
	handler.ServeHTTP(rr, req)
	assert.Equal(rr.Code, http.StatusOK)
	// Check the response body is what we expect.
	expected := `{"status":"success","data":{"message":"ssh addresses updated"}}`
	assert.JSONEq(expected, rr.Body.String())
	// check that the internal state has been updated
	assert.Equal(state.ContainerAddresses{
		1: "127.0.0.1:8230",
		2: "127.0.0.1:9809",
	}, state.GetSSHAddresses())
}
