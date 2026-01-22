package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const ctxKeyUserID ctxKey = "user_id"

func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxKeyUserID).(string)
	return v, ok
}

func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if secret == "" {
				writeErr(w, http.StatusInternalServerError, "jwt_secret_missing")
				return
			}

			ah := r.Header.Get("Authorization")
			if ah == "" {
				writeErr(w, http.StatusUnauthorized, "missing_authorization")
				return
			}

			parts := strings.SplitN(ah, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeErr(w, http.StatusUnauthorized, "invalid_authorization")
				return
			}

			tokenStr := parts[1]
			claims := jwt.MapClaims{}
			_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, errors.New("unexpected signing method")
				}
				return []byte(secret), nil
			})
			if err != nil {
				writeErr(w, http.StatusUnauthorized, "invalid_token")
				return
			}

			exp, ok := claims["exp"].(float64)
			if !ok || time.Now().Unix() > int64(exp) {
				writeErr(w, http.StatusUnauthorized, "token_expired")
				return
			}

			typeVal, ok := claims["type"].(string)
			if !ok || typeVal != "access" {
				writeErr(w, http.StatusUnauthorized, "invalid_token_type")
				return
			}

			sub, ok := claims["sub"].(string)
			if !ok || sub == "" {
				writeErr(w, http.StatusUnauthorized, "invalid_subject")
				return
			}

			ctx := context.WithValue(r.Context(), ctxKeyUserID, sub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeErr(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + code + `"}`))
}
