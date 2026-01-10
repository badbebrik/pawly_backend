package authz

import (
	"errors"
	"net/http"
	"strings"
)

var ErrMissingAuthHeader = errors.New("missing_authorization_header")
var ErrInvalidAuthHeader = errors.New("invalid_authorization_header")

func BearerToken(r *http.Request) (string, error) {
	h := strings.TrimSpace(r.Header.Get("Authorization"))
	if h == "" {
		return "", ErrMissingAuthHeader
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" || strings.TrimSpace(parts[1]) == "" {
		return "", ErrInvalidAuthHeader
	}
	return strings.TrimSpace(parts[1]), nil
}
