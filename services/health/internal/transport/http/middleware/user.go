package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type ctxKey string

const userIDKey ctxKey = "user_id"

func WithUserID(jwtSecret, jwtIssuer string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, err := extractUserID(r, jwtSecret, jwtIssuer)
			if err != nil {
				logAuthRequest(r, "reject", err.Error())
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			logAuthRequest(r, "accept", "")
			ctx := context.WithValue(r.Context(), userIDKey, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}

func extractUserID(r *http.Request, jwtSecret, jwtIssuer string) (uuid.UUID, error) {
	if raw := strings.TrimSpace(r.Header.Get("X-User-ID")); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return uuid.Nil, errors.New("invalid user id")
		}
		return id, nil
	}

	id, err := extractUserIDFromBearer(r.Header.Get("Authorization"), jwtSecret, jwtIssuer)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func extractUserIDFromBearer(authHeader, jwtSecret, jwtIssuer string) (uuid.UUID, error) {
	if strings.TrimSpace(authHeader) == "" {
		return uuid.Nil, errors.New("missing user id")
	}
	if jwtSecret == "" {
		return uuid.Nil, errors.New("missing user id")
	}

	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return uuid.Nil, errors.New("missing user id")
	}

	tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
	if tokenStr == "" {
		return uuid.Nil, errors.New("missing user id")
	}

	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	token, err := parser.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, errors.New("invalid user id")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid user id")
	}

	if jwtIssuer != "" {
		issuer, err := claims.GetIssuer()
		if err != nil || issuer != jwtIssuer {
			return uuid.Nil, errors.New("invalid user id")
		}
	}

	exp, err := claims.GetExpirationTime()
	if err != nil || exp == nil || time.Now().After(exp.Time) {
		return uuid.Nil, errors.New("invalid user id")
	}

	sub, err := claims.GetSubject()
	if err != nil || strings.TrimSpace(sub) == "" {
		return uuid.Nil, errors.New("invalid user id")
	}

	id, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, errors.New("invalid user id")
	}

	return id, nil
}

func logAuthRequest(r *http.Request, outcome, reason string) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	xUserID := strings.TrimSpace(r.Header.Get("X-User-ID"))
	const bearerPrefix = "Bearer "

	entry := log.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("outcome", outcome).
		Bool("x_user_id_present", xUserID != "").
		Bool("auth_present", authHeader != "").
		Bool("auth_bearer", strings.HasPrefix(authHeader, bearerPrefix)).
		Str("auth_preview", previewAuthHeader(authHeader)).
		Str("kong_request_id", strings.TrimSpace(r.Header.Get("X-Kong-Request-Id")))

	if xUserID != "" {
		entry = entry.Str("x_user_id", xUserID)
	}
	if reason != "" {
		entry = entry.Str("reason", reason)
	}

	entry.Msg("health auth middleware")
}

func previewAuthHeader(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	runes := []rune(authHeader)
	if len(runes) <= 20 {
		return authHeader
	}

	return string(runes[:20]) + "..."
}
