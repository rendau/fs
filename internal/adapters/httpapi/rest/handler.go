package rest

import (
	"net/http"
)

func (a *St) hRoot(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("hello"))
}
