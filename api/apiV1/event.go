package apiV1

import (
	"io/ioutil"
	"net/http"
	"strconv"

	common "github.com/mrmagooey/hpcaas-common"
	"github.com/mrmagooey/hpcaas-container-daemon/container"
)

// Event calls daemon events and triggers various daemon functions
func Event(w http.ResponseWriter, r *http.Request) {
	// get the DaemonEvent id
	respBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		jsonResponse(w, "fail", map[string]interface{}{
			"message": err.Error(),
		})
	}
	eventInt, err := strconv.Atoi(string(respBytes))
	if err != nil {
		jsonResponse(w, "fail", map[string]interface{}{
			"message": err.Error(),
		})
	}
	demonEvent := common.DaemonEvent(eventInt)
	// find the correct action to take
	switch demonEvent {
	case common.DaemonEventRunCode:
		container.ExecuteCode()
	}

	jsonResponse(w, "success", map[string]interface{}{
		"message": "daemon event received",
	})
}
