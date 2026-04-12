package gatewayauth

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const HeaderUserID = "X-User-ID"

type ctxKey string

const userIDKey ctxKey = "user_id"

func WithUserID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := UserIDFromRequestHeader(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromRequestHeader(r *http.Request) (uuid.UUID, error) {
	raw := r.Header.Get(HeaderUserID)
	return uuid.Parse(raw)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}
