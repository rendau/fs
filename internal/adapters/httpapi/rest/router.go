package rest

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (a *St) router() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/", a.hSaveFile).Methods("POST")
	r.HandleFunc("/", a.hSaveFile).Methods("GET")

	return a.middleware(r)
}
