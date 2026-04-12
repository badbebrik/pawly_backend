package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"pawly/pkg/gatewayauth"
)

func WithUserID(jwtSecret, jwtIssuer string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return gatewayauth.WithUserID(next)
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	return gatewayauth.UserIDFromContext(ctx)
}
