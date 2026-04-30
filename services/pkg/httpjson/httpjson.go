package httpjson

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

const maxBodyBytes = 1 << 20

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Decode(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("invalid_json")
	}

	return nil
}

func Write(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}

	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	Write(w, status, ErrorResponse{
		Code:    code,
		Message: message,
	})
}

func MessageFromCode(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return ""
	}

	return strings.ReplaceAll(code, "_", " ")
}
