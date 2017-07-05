package main

import "github.com/mrmagooey/hpcaas-container-daemon/api/apiV1"
import "github.com/gorilla/mux"

func register_routes() *mux.Router {
	r := mux.NewRouter()
	v1_subroute := r.PathPrefix("/v1").Subrouter()
	v1_subroute.Schemes("https")
	v1_subroute.Methods("POST").Path("/code-configuration/").HandlerFunc(apiV1.SetCodeConfig())
	v1_subroute.Methods("POST").Path("/code-name/").HandlerFunc(apiV1.SetCodeName())
	v1_subroute.Methods("POST").Path("/command/").HandlerFunc(apiV1.Command())
	return r
}
