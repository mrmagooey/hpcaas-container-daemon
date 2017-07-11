package apiV1

import "fmt"
import "encoding/json"
import "net/http"
import "errors"

import "github.com/lestrrat/go-jsschema"
import "github.com/lestrrat/go-jsval/builder"
import "github.com/lestrrat/go-jsval"

import "github.com/mrmagooey/hpcaas-container-daemon/container"
import "github.com/mrmagooey/hpcaas-container-daemon/state"

type response_json struct {
	Status string                 `json:"status"`
	Data   map[string]interface{} `json:"data"`
}

// load the json schema at schemaFilename and return a validator back
func getJSONValidator(schemaFilename string) *jsval.JSVal {
	// json schema validation
	s, err := schema.ReadFile(schemaFilename)
	if err != nil {
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
	resp := response_json{}
	resp.Status = status
	resp.Data = data
	resp_bytes, err := json.Marshal(resp)
	if err != nil {
		panic("Can't marshall the json response")
	}
	w.Write(resp_bytes)
}

type parameter_keyval struct {
	Key   string
	Value interface{}
}

func validatePOSTRequest(r *http.Request, v *jsval.JSVal) (jsonStruct map[string]interface{}, err error) {
	if r.Body == nil {
		return nil, errors.New("Bad Request")
	}
	decoder := json.NewDecoder(r.Body)
	var json_request map[string]interface{}
	e := decoder.Decode(&json_request)
	if e != nil {
		return nil, errors.New("Bad JSON content in request")
	}
	if err := v.Validate(json_request); err != nil {
		return nil, errors.New("JSON failed to validate")
	}
	// json_request is now populated and valid
	return json_request, nil
}

func SetCodeConfig() func(w http.ResponseWriter, r *http.Request) {
	v := getJSONValidator(`schemas/setCodeConfig.json`)
	return func(w http.ResponseWriter, r *http.Request) {
		requestJSON, err := validatePOSTRequest(r, v)
		if err != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": err.Error(),
			})
			return
		}
		params := requestJSON["codeParameters"].(map[string]interface{})
		// send to state
		err = state.SetCodeParams(params)
		if err != nil {
			jsonResponse(w, "error", map[string]interface{}{
				"message": "parameters failed to set",
			})
		}
		// write to disk
		err = container.WriteCodeParams()
		if err != nil {
			jsonResponse(w, "error", map[string]interface{}{
				"message": "parameters failed to set",
			})
			return
		}
		jsonResponse(w, "success", map[string]interface{}{
			"message": "parameter accepted",
		})
	}
}

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
				"message": "parameters failed to set",
			})
		}
		jsonResponse(w, "success", map[string]interface{}{
			"message": "parameter accepted",
		})
	}
}

func SetCodeState() func(w http.ResponseWriter, r *http.Request) {
	v := getJSONValidator(`schemas/setCodeState.json`)
	return func(w http.ResponseWriter, r *http.Request) {
		requestJSON, err := validatePOSTRequest(r, v)
		if err != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": err.Error(),
			})
		}
		params := requestJSON["containerParameters"].(map[string]interface{})
		// send to state
		err = state.SetContainerParams(params)
		if err != nil {
			jsonResponse(w, "error", map[string]interface{}{
				"message": "parameters failed to set",
			})
		}
		jsonResponse(w, "success", map[string]interface{}{
			"message": "parameter accepted",
		})
	}
}

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
			} else {
				jsonResponse(w, "success", map[string]interface{}{
					"message": "Code started",
				})
			}
		}
		if commandString == "kill" {
			err = container.KillCode()
			if err != nil {
				jsonResponse(w, "error", map[string]interface{}{
					"message": err.Error(),
				})
				return
			} else {
				jsonResponse(w, "success", map[string]interface{}{
					"message": "Code killed",
				})
				return
			}
		}
	}
}

func SetSSHAddresses() func(w http.ResponseWriter, r *http.Request) {
	v := getJSONValidator(`schemas/setSSHAddresses.json`)
	return func(w http.ResponseWriter, r *http.Request) {
		requestJSON, err := validatePOSTRequest(r, v)
		if err != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": err.Error(),
			})
		}
		// get the ssh addresses
		sshAddresses := requestJSON["sshAddresses"].(map[int]string)
		// update state
		state.SetSSHAddresses(sshAddresses)
		// use the ssh addresses to generate a .ssh/config file
		container.WriteSSHConfig(sshAddresses)
		// the json schema should ensure that these are the only possibilities
	}
}
