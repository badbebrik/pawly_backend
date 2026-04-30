package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}

	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}

	n, err := r.ResponseWriter.Write(data)
	r.bytes += n
	return n, err
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w}

		next.ServeHTTP(rec, req)

		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}
		if !shouldLogRequest(status) {
			return
		}

		logEvent(status).
			Str("request_id", chimw.GetReqID(req.Context())).
			Str("method", req.Method).
			Str("path", req.URL.Path).
			Str("route", routePattern(req)).
			Int("status", status).
			Int("bytes", rec.bytes).
			Dur("duration", time.Since(start)).
			Msg("http request")
	})
}

func shouldLogRequest(status int) bool {
	return status >= http.StatusBadRequest
}

func logEvent(status int) *zerolog.Event {
	switch {
	case status >= http.StatusInternalServerError:
		return log.Error()
	case status >= http.StatusBadRequest:
		return log.Warn()
	default:
		return log.Info()
	}
}

func routePattern(req *http.Request) string {
	routeCtx := chi.RouteContext(req.Context())
	if routeCtx == nil {
		return ""
	}

	pattern := strings.TrimSpace(routeCtx.RoutePattern())
	if pattern == "" {
		return ""
	}
	return pattern
}
