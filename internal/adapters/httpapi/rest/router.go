/*
Package rest FS API.

### HTTP File server
This service handles file upload and download operations.

    Schemes: http, https
    Version: 1.0.0

    Consumes:
    - application/json

    Produces:
    - application/json

swagger:meta
*/
package rest

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (a *St) router() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/doc", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "doc/")
		w.WriteHeader(http.StatusMovedPermanently)
	})
	r.PathPrefix("/doc/").Handler(a.mwNoCache(http.StripPrefix("/doc/", http.FileServer(http.Dir("./doc/")))))

	r.HandleFunc("/", a.hSave).Methods("POST")
	r.HandleFunc("/clean", a.hClean).Methods("GET")
	r.PathPrefix("/").HandlerFunc(a.hGet).Methods("GET")

	return a.middleware(r)
}
