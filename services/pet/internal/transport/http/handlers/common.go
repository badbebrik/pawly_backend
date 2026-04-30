package handlers

import (
	"errors"
	"net/http"
	"pawly/pkg/httpjson"
	"pet/internal/application/usecase"
)

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, usecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid input")
	case errors.Is(err, usecase.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden", "forbidden")
	case errors.Is(err, usecase.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "not found")
	case errors.Is(err, usecase.ErrConflict):
		writeError(w, http.StatusConflict, "conflict", "conflict")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal error")
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	httpjson.WriteError(w, status, code, message)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	httpjson.Write(w, status, body)
}

func decodeJSON(r *http.Request, dst any) error {
	return httpjson.Decode(r, dst)
}
