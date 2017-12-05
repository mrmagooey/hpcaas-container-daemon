package main

import (
	"log"
	"net/http"

	"github.com/mrmagooey/hpcaas-container-daemon/state"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Before")
		if r.Header.Get("WWW-Authenticate") == state.GetAuthorizationKey() {
			next.ServeHTTP(w, r) // call original routes
		}
		// reject
		http.Error(w, "Bad Authorization header", 401)
	})
}
