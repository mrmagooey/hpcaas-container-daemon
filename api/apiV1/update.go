package apiV1

import (
	"encoding/json"
	"net/http"

	common "github.com/mrmagooey/hpcaas-common"
	"github.com/mrmagooey/hpcaas-container-daemon/state"
)

// Update take state in json format and update daemon state
func Update(w http.ResponseWriter, r *http.Request) {
	newState := &common.DaemonState{}
	err := json.NewDecoder(r.Body).Decode(newState)
	if err != nil {
		jsonResponse(w, "fail", map[string]interface{}{
			"message": err.Error(),
		})
	}
	state.SetDaemonState(*newState)
	jsonResponse(w, "success", map[string]interface{}{
		"message": "daemon updated",
	})

}
