package container

import "encoding/json"
import "io/ioutil"
import "bytes"
import "fmt"
import "sync"

var parameterJSONPath = "/hpcaas/runtime/parameters.json"
var parameterPath = "/hpcaas/runtime/parameters"

type codeParamRequest struct {
	Data       map[string]string
	ReturnChan chan error
}

// var writeCodeParamsChan = make(chan codeParamRequest)
var writeCodeMut = sync.Mutex{}

// WriteCodeParams write the params to disk
// does so by sending the values over the channel to the
// disk writing goroutine
func WriteCodeParams(params map[string]string) error {
	// write json
	newJSON, err := json.Marshal(params)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(parameterJSONPath, newJSON, 0777)
	if err != nil {
		return err
	}
	// write newline separated file
	var buffer bytes.Buffer
	for k, v := range params {
		envLine := fmt.Sprintf("%s=%v\n", k, v)
		buffer.WriteString(envLine)
	}
	err = ioutil.WriteFile(parameterPath, buffer.Bytes(), 0777)
	return err
}
