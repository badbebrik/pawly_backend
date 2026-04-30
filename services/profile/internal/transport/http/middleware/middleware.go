package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"pawly/pkg/gatewayauth"
	"pawly/pkg/httpjson"
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
				httpjson.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
				return
			}

			token := r.Header.Get("X-Internal-Token")
			if token == "" || token != expectedToken {
				httpjson.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
