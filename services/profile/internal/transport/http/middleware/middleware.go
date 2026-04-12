package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"pawly/pkg/gatewayauth"
)

func WithUserID(next http.Handler) http.Handler {
	return gatewayauth.WithUserID(next)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	return gatewayauth.UserIDFromContext(ctx)
}

func WithInternalToken(expectedToken string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expectedToken == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			token := r.Header.Get("X-Internal-Token")
			if token == "" || token != expectedToken {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
