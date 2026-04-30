package middleware

import (
	"net/http"

	"pawly/pkg/httpjson"
)

func WithInternalToken(expected string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Internal-Token")
			if expected == "" || token != expected {
				httpjson.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
