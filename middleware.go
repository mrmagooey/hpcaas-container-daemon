package main

import (
	"log"
	"net/http"

	"github.com/mrmagooey/hpcaas-container-daemon/state"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Before")
		if authKey, ok := state.GetAuthorizationKey(); ok {
			if r.Header.Get("WWW-Authenticate") == authKey {
				next.ServeHTTP(w, r) // call original routes
				return
			}
		}
		// reject
		http.Error(w, "Bad Authorization header", 401)
		return
	})
}
