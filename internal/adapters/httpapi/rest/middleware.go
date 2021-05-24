package rest

import (
	"net/http"

	"github.com/rs/cors"
)

func (a *St) middleware(h http.Handler) http.Handler {
	h = cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool { return true },
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodConnect,
			http.MethodOptions,
			http.MethodTrace,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           604800,
	}).Handler(h)

	h = a.mwRecovery(h)

	return h
}

func (a *St) mwRecovery(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				a.uLogErrorRequest(r, err, "Panic in http handler")

				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}
