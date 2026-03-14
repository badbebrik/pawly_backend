package middleware

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey string

const localeKey ctxKey = "locale"

func WithLocale(fallback string) func(http.Handler) http.Handler {
	if fallback == "" {
		fallback = "ru"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loc := fromRequest(r, fallback)
			ctx := context.WithValue(r.Context(), localeKey, loc)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func LocaleFromCtx(ctx context.Context, fallback string) string {
	if fallback == "" {
		fallback = "ru"
	}
	if v, ok := ctx.Value(localeKey).(string); ok && v != "" {
		return v
	}
	return fallback
}

func fromRequest(r *http.Request, fallback string) string {
	al := strings.ToLower(r.Header.Get("Accept-Language"))
	if al == "" {
		return fallback
	}

	if strings.Contains(al, "ru") {
		return "ru"
	}
	if strings.Contains(al, "en") {
		return "en"
	}

	return fallback
}
