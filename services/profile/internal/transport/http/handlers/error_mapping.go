package handlers

import (
	"errors"
	"net/http"
	"profile/internal/application/ports"
	profileuc "profile/internal/application/usecase"
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

func writeProfileQueryError(w http.ResponseWriter, err error) {
	writeMappedError(w, err,
		errorRule{err: profileuc.ErrInvalidInput, status: http.StatusBadRequest},
		errorRule{err: ports.ErrNotFound, status: http.StatusNotFound},
	)
}

func writeUpdateProfileInfoError(w http.ResponseWriter, err error) {
	writeMappedError(w, err,
		errorRule{err: profileuc.ErrInvalidInput, status: http.StatusBadRequest},
		errorRule{err: ports.ErrNotFound, status: http.StatusNotFound},
	)
}

func writeUpdatePreferencesError(w http.ResponseWriter, err error) {
	writeMappedError(w, err,
		errorRule{err: profileuc.ErrInvalidInput, status: http.StatusBadRequest},
		errorRule{err: profileuc.ErrInvalidLocale, status: http.StatusBadRequest},
		errorRule{err: profileuc.ErrInvalidTimezone, status: http.StatusBadRequest},
		errorRule{err: ports.ErrNotFound, status: http.StatusNotFound},
	)
}

func writeAvatarError(w http.ResponseWriter, err error) {
	writeMappedError(w, err,
		errorRule{err: profileuc.ErrInvalidInput, status: http.StatusBadRequest},
		errorRule{err: profileuc.ErrAvatarUpload, status: http.StatusBadRequest},
		errorRule{err: ports.ErrNotFound, status: http.StatusNotFound},
	)
}

func writeBatchProfilesBriefError(w http.ResponseWriter, err error) {
	writeMappedError(w, err,
		errorRule{err: profileuc.ErrInvalidInput, status: http.StatusBadRequest},
	)
}
