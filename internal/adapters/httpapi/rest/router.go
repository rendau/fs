package rest

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (a *St) router() http.Handler {
	r := mux.NewRouter()

	r.PathPrefix("/").HandlerFunc(a.hRoot).Methods("GET")

	return a.middleware(r)
}
