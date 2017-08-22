package apiV1

import (
	"encoding/json"
	"errors"
	//	"fmt"
	"github.com/alecthomas/jsonschema"
	"github.com/mrmagooey/hpcaas-container-daemon/container"
	"github.com/mrmagooey/hpcaas-container-daemon/state"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
	"net/http"
	"time"
)

type responseJSON struct {
	Status string                 `json:"status"`
	Data   map[string]interface{} `json:"data"`
}

// generate a schema
// takes a pointer to an initialized copy of the string
func getJSONValidator(schemaStruct interface{}) *gojsonschema.Schema {
	// generate a json schema struct from the original go struct
	jsonSchema := jsonschema.Reflect(schemaStruct)
	// turn the json schema struct into []byte
	jsonSchemaBytes, err := json.Marshal(jsonSchema)
	if err != nil {
		panic(err)
	}
	// turn the jsonSchemaBytes into a schema validator
	schemaLoader := gojsonschema.NewStringLoader(string(jsonSchemaBytes))
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		panic(err.Error())
	}
	return schema
}

// provide a standardised JSON response back to the client
func jsonResponse(w http.ResponseWriter, status string, data map[string]interface{}) {
	resp := responseJSON{
		Status: status,
		Data:   data,
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		panic("Can't marshall the json response")
	}
	w.Write(respBytes)
}

// check that r.body is populated, and that it satisfies schema
func validatePOSTRequest(body []byte, schema *gojsonschema.Schema) (err error) {
	docLoader := gojsonschema.NewStringLoader(string(body))
	if result, err := schema.Validate(docLoader); err != nil && !result.Valid() {
		return errors.New("Couldn't validate json")
	}
	return nil
}

// Heartbeat returns an empty response
func Heartbeat(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	w.Write([]byte(t.Format(time.UnixDate)))
}

type setCodeParamsStruct struct {
	CodeParameters map[string]string
}

// var setCodeParamsSchema = `{
//   "type": "object",
//   "properties": {
//     "codeParameters": {
//       "description": "A list of configuration items that the code will use",
//       "type": "object",
//       "additionalProperties": {
//         "anyOf": [
//           {
//             "type": "string"
//           }
//         ]
//       }
//     }
//   },
//   "required": [
//     "codeParameters"
//   ]
// }`

// SetCodeParams returns a closure that handles http requests
func SetCodeParams() func(w http.ResponseWriter, r *http.Request) {
	schema := getJSONValidator(&setCodeParamsStruct{})
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": err.Error(),
			})
			return
		}
		err = validatePOSTRequest(body, schema)
		if err != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": err.Error(),
			})
			return
		}
		var jsonRequest = &setCodeParamsStruct{}
		json.Unmarshal(body, jsonRequest)
		if err != nil {
			jsonResponse(w, "error", map[string]interface{}{
				"message": err.Error(),
			})
			return
		}
		err = state.SetCodeParams(jsonRequest.CodeParameters)
		if err != nil {
			jsonResponse(w, "error", map[string]interface{}{
				"message": err.Error(),
			})
			return
		}
		// write to disk
		err = container.WriteCodeParams(jsonRequest.CodeParameters)
		if err != nil {
			jsonResponse(w, "error", map[string]interface{}{
				"message": err.Error(),
			})
			return
		}
		jsonResponse(w, "success", map[string]interface{}{
			"message": "parameter accepted",
		})
	}
}

type setCodeNameStruct struct {
	CodeName string `json:"codeName"`
}

// var setCodeNameSchema = `{
//   "type": "object",
//   "properties": {
//     "codeName": {
//       "description": "Name of the code to be executed",
//       "type": "string"
//     }
//   },
//   "required": [
//     "codeName"
//   ]
// }`

// SetCodeName closure returning http handler that sets the code name
func SetCodeName() func(w http.ResponseWriter, r *http.Request) {
	schema := getJSONValidator(&setCodeNameStruct{})
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": err.Error(),
			})
			return
		}
		err = validatePOSTRequest(body, schema)
		if err != nil {
			jsonResponse(w, "error", map[string]interface{}{
				"message": err.Error(),
			})
		}
		var responseStruct = &setCodeNameStruct{}
		json.Unmarshal(body, responseStruct)
		// send to state
		state.SetCodeName(responseStruct.CodeName)
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

type setCodeStateStruct struct {
	CodeState float64 `json:"codeState"`
}

// var setCodeStateSchema = bytes.NewBufferString(`{
//   "type": "object",
//   "properties": {
//     "codeState": {
//       "description": "The state that the code will be set to",
//       "type": "number"
//     }
//   },
//   "required": [
//     "codeState"
//   ]
// }`)

// SetCodeState closure returning http handler that sets the code state
func SetCodeState() func(w http.ResponseWriter, r *http.Request) {
	schema := getJSONValidator(&setCodeStateStruct{})
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": err.Error(),
			})
			return
		}
		err = validatePOSTRequest(body, schema)
		if err != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": err.Error(),
			})
		}
		var responseStruct = &setCodeStateStruct{}
		json.Unmarshal(body, responseStruct)
		// send to state
		state.SetCodeState(uint8(responseStruct.CodeState))
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

type commandSchemaStruct struct {
	Command string `json:"command"`
}

// var commandSchema = bytes.NewBufferString(`{
// 	"type": "object",
// 	"properties": {
// 		"command": {
// 			"description": "Command",
// 			"type": "string",
// 			"enum": ["start", "kill"]
// 		}
// 	},
// 	"required": [
// 		"codeState"
// 	]
// }`)

// Command closure that returning http handler that gives a command to the daemon
// this is responsible for starting and killing code
func Command() func(w http.ResponseWriter, r *http.Request) {
	schema := getJSONValidator(&commandSchemaStruct{})
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": err.Error(),
			})
			return
		}
		err = validatePOSTRequest(body, schema)
		if err != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": err.Error(),
			})
			return
		}
		var responseStruct = &commandSchemaStruct{}
		json.Unmarshal(body, responseStruct)
		if err != nil {
			jsonResponse(w, "error", map[string]interface{}{
				"message": err.Error(),
			})
		}
		// the json schema should ensure that these are the only possibilities
		if responseStruct.Command == "start" {
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
		} else if responseStruct.Command == "kill" {
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
		} else {
			jsonResponse(w, "error", map[string]interface{}{
				"message": "need either kill or start",
			})
			return
		}
	}
}

type setSSHAddressesSchemaStruct struct {
	SSHAddresses map[int]string `json:"sshAddresses"`
}

// var setSSHAddressesSchema = bytes.NewBufferString(`{
//   "type": "object",
//   "properties": {
//     "sshAddresses": {
//       "description": "Keys are container ids and values are the <ip>:<port> strings",
//       "type": "object",
//       "patternProperties": {
//         "^\\d+$": {"type": "string"}
//       },
//       "additionalProperties": false
//     }
//   },
//   "required": [
//     "sshAddresses"
//   ]
// }`)

// SetSSHAddresses closure that returns http handler responsible for setting ssh addresses of other containers
func SetSSHAddresses() func(w http.ResponseWriter, r *http.Request) {
	schema := getJSONValidator(&setSSHAddressesSchemaStruct{})
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": err.Error(),
			})
			return
		}
		err = validatePOSTRequest(body, schema)
		if err != nil {
			jsonResponse(w, "fail", map[string]interface{}{
				"message": err.Error(),
			})
			return
		}

		var responseStruct = &setSSHAddressesSchemaStruct{}
		json.Unmarshal(body, responseStruct)

		// // get the ssh addresses
		// addrs := requestJSON["sshAddresses"].(map[string]interface{})
		// sshAddresses := make(map[int]string)
		// for key, val := range addrs {
		// 	intKey, _ := strconv.Atoi(key)
		// 	sshAddresses[intKey] = val.(string)
		// }
		// // update state
		stateErr := state.SetSSHAddresses(responseStruct.SSHAddresses)
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
