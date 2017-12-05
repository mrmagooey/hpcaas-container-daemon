package main

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mrmagooey/hpcaas-container-daemon/api/apiV1"
)

func registerRoutes() *mux.Router {
	r := mux.NewRouter()

	r.Methods("GET").Path("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		w.Write([]byte("hpcaas-daemon: " + t.Format(time.UnixDate)))
	})

	version1Subroute := r.PathPrefix("/v1").Subrouter()
	version1Subroute.Schemes("https")
	// version1Subroute.Methods("GET").Path("/heartbeat/").HandlerFunc(apiV1.Heartbeat)
	// version1Subroute.Methods("POST").Path("/code-parameters/").HandlerFunc(apiV1.SetCodeParams())
	// version1Subroute.Methods("POST").Path("/code-name/").HandlerFunc(apiV1.SetCodeName())
	// version1Subroute.Methods("POST").Path("/command/").HandlerFunc(apiV1.Command())

	// update any state variable
	version1Subroute.Methods("POST").Path("/update/").HandlerFunc(apiV1.Update)
	// get the current state of the daemon
	version1Subroute.Methods("GET").Path("/states/").HandlerFunc(apiV1.State)
	// send an event
	version1Subroute.Methods("POST").Path("/event/").HandlerFunc(apiV1.Event)

	return r
}
