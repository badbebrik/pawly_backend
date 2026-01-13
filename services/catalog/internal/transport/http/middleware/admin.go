package middleware

import (
	"net/http"
)

func RequireAdminToken(expected string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expected == "" {
				http.Error(w, "admin token not configured", http.StatusServiceUnavailable)
				return
			}
			got := r.Header.Get("X-Admin-Token")
			if got == "" || got != expected {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
