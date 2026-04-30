package handlers

import (
	"acl/internal/application/ports"
	acluc "acl/internal/application/usecase"
	"errors"
	"net/http"
)

type errorRule struct {
	err    error
	status int
}

func writeMappedError(w http.ResponseWriter, err error, rules ...errorRule) {
	for _, rule := range rules {
		if errors.Is(err, rule.err) {
			writeServiceError(w, rule.status, err)
			return
		}
	}

	writeInternalError(w)
}

func writeACLError(w http.ResponseWriter, err error) {
	writeMappedError(w, err,
		errorRule{err: acluc.ErrInvalidInput, status: http.StatusBadRequest},
		errorRule{err: acluc.ErrForbidden, status: http.StatusForbidden},
		errorRule{err: acluc.ErrNotFound, status: http.StatusNotFound},
		errorRule{err: acluc.ErrConflict, status: http.StatusConflict},
		errorRule{err: ports.ErrForbidden, status: http.StatusForbidden},
		errorRule{err: ports.ErrNotFound, status: http.StatusNotFound},
		errorRule{err: ports.ErrConflict, status: http.StatusConflict},
	)
}
