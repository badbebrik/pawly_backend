package handlers

import (
	"net/http"
	"strings"

	"pawly/pkg/httpjson"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	httpjson.Write(w, status, v)
}

func writeServiceError(w http.ResponseWriter, status int, err error) {
	writeError(w, status, err.Error(), httpjson.MessageFromCode(err.Error()))
}

func writeInternalError(w http.ResponseWriter) {
	writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	httpjson.WriteError(w, status, code, message)
}

func decodeJSON(r *http.Request, dst any) error {
	return httpjson.Decode(r, dst)
}

func firstNonEmpty(values ...string) string {
	for i := range values {
		if strings.TrimSpace(values[i]) != "" {
			return values[i]
		}
	}
	return ""
}
