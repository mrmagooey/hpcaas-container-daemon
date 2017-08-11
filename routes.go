package main

import "github.com/mrmagooey/hpcaas-container-daemon/api/apiV1"
import "github.com/gorilla/mux"

func registerRoutes() *mux.Router {
	r := mux.NewRouter()
	v1_subroute := r.PathPrefix("/v1").Subrouter()
	v1_subroute.Schemes("https")
	v1_subroute.Methods("POST").Path("/code-parameters/").HandlerFunc(apiV1.SetCodeParams())
	v1_subroute.Methods("POST").Path("/code-name/").HandlerFunc(apiV1.SetCodeName())
	v1_subroute.Methods("POST").Path("/command/").HandlerFunc(apiV1.Command())
	return r
}
