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
