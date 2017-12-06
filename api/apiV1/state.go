package apiV1

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mrmagooey/hpcaas-container-daemon/state"
)

// State calls daemon events and triggers various daemon functions
func State(w http.ResponseWriter, r *http.Request) {

	daemonState := state.GetDaemonState()
	stateBytes, err := json.Marshal(&daemonState)
	if err != nil {
		log.Println("Fail to encode state for /state/ request")
		jsonResponse(w, "fail", map[string]interface{}{
			"message": "couldn't encode state",
		})
		return
	}

	data := map[string]interface{}{
		"data": string(stateBytes),
	}

	jsonResponse(w, "success", data)

}
