package handlers

import (
	"errors"
	"net/http"
	profileuc "profile/internal/application/usecase"
	"profile/internal/repository"
)

type errorRule struct {
	err    error
	status int
}

func writeMappedUseCaseError(w http.ResponseWriter, err error, details any, rules ...errorRule) {
	for _, rule := range rules {
		if errors.Is(err, rule.err) {
			if details != nil {
				writeError(w, rule.status, err.Error(), details)
				return
			}
			writeServiceError(w, rule.status, err)
			return
		}
	}
	writeInternalError(w)
}

func writeProfileQueryError(w http.ResponseWriter, err error) {
	writeMappedUseCaseError(w, err, nil,
		errorRule{err: profileuc.ErrInvalidInput, status: http.StatusBadRequest},
		errorRule{err: repository.ErrNotFound, status: http.StatusNotFound},
	)
}

func writeUpdateProfileInfoError(w http.ResponseWriter, err error) {
	writeMappedUseCaseError(w, err, nil,
		errorRule{err: profileuc.ErrInvalidInput, status: http.StatusBadRequest},
		errorRule{err: repository.ErrNotFound, status: http.StatusNotFound},
	)
}

func writeUpdatePreferencesError(w http.ResponseWriter, err error) {
	writeMappedUseCaseError(w, err, nil,
		errorRule{err: profileuc.ErrInvalidInput, status: http.StatusBadRequest},
		errorRule{err: profileuc.ErrInvalidLocale, status: http.StatusBadRequest},
		errorRule{err: profileuc.ErrInvalidTimezone, status: http.StatusBadRequest},
		errorRule{err: repository.ErrNotFound, status: http.StatusNotFound},
	)
}

func writeAvatarError(w http.ResponseWriter, err error) {
	writeMappedUseCaseError(w, err, nil,
		errorRule{err: profileuc.ErrInvalidInput, status: http.StatusBadRequest},
		errorRule{err: profileuc.ErrAvatarUpload, status: http.StatusBadRequest},
		errorRule{err: repository.ErrNotFound, status: http.StatusNotFound},
	)
}

func writeBatchProfilesBriefError(w http.ResponseWriter, err error) {
	writeMappedUseCaseError(w, err, nil,
		errorRule{err: profileuc.ErrInvalidInput, status: http.StatusBadRequest},
	)
}
