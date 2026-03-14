package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeServiceError(w http.ResponseWriter, status int, err error) {
	writeError(w, status, err.Error(), nil)
}

func writeInternalError(w http.ResponseWriter) {
	writeError(w, http.StatusInternalServerError, "internal_error", nil)
}

func writeError(w http.ResponseWriter, status int, code string, details any) {
	writeJSON(w, status, errorResponse{
		Code:    code,
		Message: code,
		Details: details,
	})
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func firstNonEmpty(values ...string) string {
	for i := range values {
		if strings.TrimSpace(values[i]) != "" {
			return values[i]
		}
	}
	return ""
}
