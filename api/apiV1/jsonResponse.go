package apiV1

import (
	"encoding/json"
	"log"
	"net/http"
)

// provide a standardised JSON response back to the client
func jsonResponse(w http.ResponseWriter, status string, data map[string]interface{}) {
	resp := responseJSON{
		Status: status,
		Data:   data,
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		log.Println("Couldn't marshal json response")
		return
	}

	w.Write(respBytes)
}
