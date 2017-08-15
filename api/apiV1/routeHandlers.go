package apiV1

import "fmt"
import "encoding/json"
import "net/http"
import "errors"
import "strconv"

import "github.com/lestrrat/go-jsschema"
import "github.com/lestrrat/go-jsval/builder"
import "github.com/lestrrat/go-jsval"

import "github.com/mrmagooey/hpcaas-container-daemon/container"
import "github.com/mrmagooey/hpcaas-container-daemon/state"

type responseJSON struct {
	Status string                 `json:"status"`
	Data   map[string]interface{} `json:"data"`
}

// load the json schema at schemaFilename and return a validator back
func getJSONValidator(schemaFilename string) *jsval.JSVal {
	// json schema validation
	s, err := schema.ReadFile(schemaFilename)
	if err != nil {
		fmt.Println(err)
		panic(fmt.Sprintf("Failed to open %s", schemaFilename))
	}
	b := builder.New()
	v, err := b.Build(s)
	if err != nil {
		panic(fmt.Sprintf("Failed to build validator for %s", schemaFilename))
	}
	return v
}

// provide a standardised JSON response back to the client
func jsonResponse(w http.ResponseWriter, status string, data map[string]interface{}) {
	resp := responseJSON{}
	resp.Status = status
	resp.Data = data
	respBytes, err := json.Marshal(resp)
	if err != nil {
		panic("Can't marshall the json response")
	}
	w.Write(respBytes)
}

type parameterKeyval struct {
	Key   string
	Value interface{}
}

func validatePOSTRequest(r *http.Request, v *jsval.JSVal) (jsonStruct map[string]interface{}, err error) {
	if r.Body == nil {
		return nil, errors.New("Bad Request")
	}
	decoder := json.NewDecoder(r.Body)
	var jsonRequest map[string]interface{}
	e := decoder.Decode(&jsonRequest)
	if e != nil {
		return nil, e
	}
	if err := v.Validate(jsonRequest); err != nil {
		return nil, err
	}
	// json_request is now populated and valid
	return jsonRequest, nil
}

// SetCodeParams returns a closure that handles http requests
func SetCodeParams() func(w http.ResponseWriter, r *http.Request) {
	v := getJSONValidator(`schemas/setCodeParams.json`)
	return func(w http.ResponseWriter, r *http.Request) {
		requestJSON, err := validatePOSTRequest(r, v)
		if err != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": err.Error(),
			})
			return
		}
		// extract parameters and cast
		codeParameters := requestJSON["codeParameters"].(map[string]interface{})
		params := make(map[string]string)
		// concretize values to strings
		for key, value := range codeParameters {
			params[key] = value.(string)
		}
		err = state.SetCodeParams(params)
		if err != nil {
			jsonResponse(w, "error", map[string]interface{}{
				"message": fmt.Sprintf("state failed to set: %s", err.Error()),
			})
			return
		}
		// write to disk
		err = container.WriteCodeParams(params)
		if err != nil {
			jsonResponse(w, "error", map[string]interface{}{
				"message": fmt.Sprintf("parameters failed to write to disk: %s", err.Error()),
			})
			return
		}
		jsonResponse(w, "success", map[string]interface{}{
			"message": "parameter accepted",
		})
	}
}

// SetCodeName closure returning http handler that sets the code name
func SetCodeName() func(w http.ResponseWriter, r *http.Request) {
	v := getJSONValidator(`schemas/setCodeName.json`)
	return func(w http.ResponseWriter, r *http.Request) {
		requestJSON, err := validatePOSTRequest(r, v)
		if err != nil {
			jsonResponse(w, "error", map[string]interface{}{
				"message": err.Error(),
			})
		}
		// send to state
		state.SetCodeName(requestJSON["codeName"].(string))
		if err != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": "name failed to set",
			})
			return
		}
		jsonResponse(w, "success", map[string]interface{}{
			"message": "name accepted",
		})
	}
}

// SetCodeState closure returning http handler that sets the code state
func SetCodeState() func(w http.ResponseWriter, r *http.Request) {
	v := getJSONValidator(`schemas/setCodeState.json`)
	return func(w http.ResponseWriter, r *http.Request) {
		requestJSON, err := validatePOSTRequest(r, v)
		if err != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": err.Error(),
			})
		}
		codeState := uint8(requestJSON["codeState"].(float64))
		// send to state
		state.SetCodeState(codeState)
		if err != nil {
			jsonResponse(w, "error", map[string]interface{}{
				"message": "state failed to set",
			})
			return
		}
		jsonResponse(w, "success", map[string]interface{}{
			"message": "state accepted",
		})
	}
}

// Command closure that returning http handler that gives a command to the daemon
// this is responsible for starting and killing code
func Command() func(w http.ResponseWriter, r *http.Request) {
	v := getJSONValidator(`schemas/command.json`)
	return func(w http.ResponseWriter, r *http.Request) {
		requestJSON, err := validatePOSTRequest(r, v)
		if err != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": err.Error(),
			})
			return
		}
		commandString := requestJSON["command"].(string)
		// the json schema should ensure that these are the only possibilities
		if commandString == "start" {
			err = container.ExecuteCode()
			if err != nil {
				jsonResponse(w, "error", map[string]interface{}{
					"message": err.Error(),
				})
				return
			}
			jsonResponse(w, "success", map[string]interface{}{
				"message": "code started",
			})
			return
		}
		if commandString == "kill" {
			err = container.KillCode()
			if err != nil {
				jsonResponse(w, "error", map[string]interface{}{
					"message": err.Error(),
				})
				return
			}
			jsonResponse(w, "success", map[string]interface{}{
				"message": "code killed",
			})
			return
		}
	}
}

// SetSSHAddresses closure that returns http handler responsible for setting ssh addresses of other containers
func SetSSHAddresses() func(w http.ResponseWriter, r *http.Request) {
	v := getJSONValidator(`schemas/setSSHAddresses.json`)
	return func(w http.ResponseWriter, r *http.Request) {
		requestJSON, err := validatePOSTRequest(r, v)
		if err != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": err.Error(),
			})
			return
		}
		// get the ssh addresses
		addrs := requestJSON["sshAddresses"].(map[string]interface{})
		sshAddresses := make(map[int]string)
		for key, val := range addrs {
			intKey, _ := strconv.Atoi(key)
			sshAddresses[intKey] = val.(string)
		}
		// update state
		stateErr := state.SetSSHAddresses(sshAddresses)
		if stateErr != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": err.Error(),
			})
			return
		}
		// use the ssh addresses to generate a .ssh/config file
		writeErr := container.WriteSSHConfig()
		// the json schema should ensure that these are the only possibilities
		if writeErr != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": err.Error(),
			})
			return
		}
		jsonResponse(w, "success", map[string]interface{}{
			"message": "ssh addresses updated",
		})
	}
}
