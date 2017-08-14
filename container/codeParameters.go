package container

import "encoding/json"
import "io/ioutil"
import "bytes"
import "fmt"

var parameterJSONPath = "/hpcaas/runtime/parameters.json"
var parameterPath = "/hpcaas/runtime/parameters"

type codeParamRequest struct {
	Data       map[string]string
	ReturnChan chan error
}

var writeCodeParamsChan = make(chan codeParamRequest)

// write code param state to disk
func writeCodeParams() {
	for {
		req := <-writeCodeParamsChan
		params := req.Data
		returnChan := req.ReturnChan
		// write json
		newJSON, err := json.Marshal(params)
		if err != nil {
			returnChan <- err
			return
		}
		err = ioutil.WriteFile(parameterJSONPath, newJSON, 0777)
		if err != nil {
			returnChan <- err
			return
		}
		// write newline separated file
		var buffer bytes.Buffer
		for k, v := range params {
			envLine := fmt.Sprintf("%s=%v\n", k, v)
			buffer.WriteString(envLine)
		}
		err = ioutil.WriteFile(parameterPath, buffer.Bytes(), 0777)
		returnChan <- err
	}
}

func WriteCodeParams(params map[string]string) error {
	var returnChan = make(chan error)
	req := codeParamRequest{params, returnChan}
	writeCodeParamsChan <- req
	return <-returnChan
}

func init() {
	go writeCodeParams()
}
