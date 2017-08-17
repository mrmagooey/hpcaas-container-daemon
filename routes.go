package main

import "github.com/mrmagooey/hpcaas-container-daemon/api/apiV1"
import "github.com/gorilla/mux"

func registerRoutes() *mux.Router {
	r := mux.NewRouter()
	version1Subroute := r.PathPrefix("/v1").Subrouter()
	version1Subroute.Schemes("https")
	version1Subroute.Methods("GET").Path("/heartbeat/").HandlerFunc(apiV1.Heartbeat)
	version1Subroute.Methods("POST").Path("/code-parameters/").HandlerFunc(apiV1.SetCodeParams())
	version1Subroute.Methods("POST").Path("/code-name/").HandlerFunc(apiV1.SetCodeName())
	version1Subroute.Methods("POST").Path("/command/").HandlerFunc(apiV1.Command())
	return r
}
