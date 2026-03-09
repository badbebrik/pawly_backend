package middleware

import (
	"context"
	"github.com/google/uuid"
	"net/http"
)

type ctxKey string

const userIDKey ctxKey = "user_id"

func WithUserID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := r.Header.Get("X-User-ID")
		if raw == "" {
			http.Error(w, "missing user id", http.StatusUnauthorized)
			return
		}

		id, err := uuid.Parse(raw)
		if err != nil {
			http.Error(w, "invalid user id", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
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
