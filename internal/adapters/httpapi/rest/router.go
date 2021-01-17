package rest

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (a *St) router() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/", a.hSave).Methods("POST")
	r.HandleFunc("/clean", a.hClean).Methods("GET")
	r.PathPrefix("/").HandlerFunc(a.hGet).Methods("GET")

	return a.middleware(r)
}
