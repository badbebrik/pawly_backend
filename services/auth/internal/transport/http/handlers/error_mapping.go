package handlers

import (
	authuc "auth/internal/application/usecase"
	"errors"
	"net/http"
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

func registerEmailErrorDetails(resp *authuc.RegisterEmailOutput) any {
	if resp == nil {
		return nil
	}
	return map[string]any{
		"user_id":          resp.UserID,
		"channel":          resp.Verification.Channel,
		"code_ttl_seconds": resp.Verification.CodeTTLSeconds,
		"can_resend_in":    resp.Verification.CanResendInSeconds,
	}
}

func resendEmailVerificationErrorDetails(resp *authuc.ResendEmailVerificationOutput) any {
	if resp == nil {
		return nil
	}
	return map[string]any{
		"user_id":          resp.UserID,
		"channel":          resp.Verification.Channel,
		"code_ttl_seconds": resp.Verification.CodeTTLSeconds,
		"can_resend_in":    resp.Verification.CanResendInSeconds,
	}
}

func writeRegisterEmailError(w http.ResponseWriter, err error, resp *authuc.RegisterEmailOutput) {
	details := any(nil)
	if errors.Is(err, authuc.ErrCannotResendYet) {
		details = registerEmailErrorDetails(resp)
	}

	writeMappedUseCaseError(w, err, details,
		errorRule{err: authuc.ErrEmailAlreadyTaken, status: http.StatusConflict},
		errorRule{err: authuc.ErrUserBlocked, status: http.StatusForbidden},
		errorRule{err: authuc.ErrWeakPassword, status: http.StatusBadRequest},
		errorRule{err: authuc.ErrIncorrectFormat, status: http.StatusBadRequest},
		errorRule{err: authuc.ErrCannotResendYet, status: http.StatusTooManyRequests},
		errorRule{err: authuc.ErrVerificationFailed, status: http.StatusServiceUnavailable},
		errorRule{err: authuc.ErrProfileCreationFailed, status: http.StatusServiceUnavailable},
	)
}

func writeVerifyEmailError(w http.ResponseWriter, err error) {
	writeMappedUseCaseError(w, err, nil,
		errorRule{err: authuc.ErrIncorrectFormat, status: http.StatusBadRequest},
		errorRule{err: authuc.ErrUserNotFound, status: http.StatusNotFound},
		errorRule{err: authuc.ErrUserBlocked, status: http.StatusForbidden},
		errorRule{err: authuc.ErrVerificationCodeInvalid, status: http.StatusUnprocessableEntity},
		errorRule{err: authuc.ErrVerificationCodeExpired, status: http.StatusUnprocessableEntity},
		errorRule{err: authuc.ErrVerificationTooMany, status: http.StatusUnprocessableEntity},
	)
}

func writeResendEmailVerificationError(w http.ResponseWriter, err error, resp *authuc.ResendEmailVerificationOutput) {
	details := any(nil)
	if errors.Is(err, authuc.ErrCannotResendYet) {
		details = resendEmailVerificationErrorDetails(resp)
	}

	writeMappedUseCaseError(w, err, details,
		errorRule{err: authuc.ErrIncorrectFormat, status: http.StatusBadRequest},
		errorRule{err: authuc.ErrUserNotFound, status: http.StatusNotFound},
		errorRule{err: authuc.ErrUserBlocked, status: http.StatusForbidden},
		errorRule{err: authuc.ErrEmailAlreadyVerified, status: http.StatusConflict},
		errorRule{err: authuc.ErrCannotResendYet, status: http.StatusTooManyRequests},
		errorRule{err: authuc.ErrVerificationFailed, status: http.StatusServiceUnavailable},
	)
}

func writeLoginEmailError(w http.ResponseWriter, err error) {
	writeMappedUseCaseError(w, err, nil,
		errorRule{err: authuc.ErrIncorrectFormat, status: http.StatusBadRequest},
		errorRule{err: authuc.ErrInvalidEmailOrPassword, status: http.StatusUnauthorized},
		errorRule{err: authuc.ErrEmailNotVerified, status: http.StatusForbidden},
		errorRule{err: authuc.ErrUserBlocked, status: http.StatusForbidden},
	)
}

func writeLoginOAuthError(w http.ResponseWriter, err error) {
	writeMappedUseCaseError(w, err, nil,
		errorRule{err: authuc.ErrIncorrectFormat, status: http.StatusBadRequest},
		errorRule{err: authuc.ErrOAuthInvalidToken, status: http.StatusUnauthorized},
		errorRule{err: authuc.ErrOAuthProviderUnavailable, status: http.StatusServiceUnavailable},
		errorRule{err: authuc.ErrProfileCreationFailed, status: http.StatusServiceUnavailable},
		errorRule{err: authuc.ErrUserBlocked, status: http.StatusForbidden},
	)
}

func writeLogoutError(w http.ResponseWriter, err error) {
	writeMappedUseCaseError(w, err, nil,
		errorRule{err: authuc.ErrUnauthorized, status: http.StatusUnauthorized},
		errorRule{err: authuc.ErrSessionNotFound, status: http.StatusNotFound},
	)
}

func writeLogoutAllError(w http.ResponseWriter, err error) {
	writeMappedUseCaseError(w, err, nil,
		errorRule{err: authuc.ErrUnauthorized, status: http.StatusUnauthorized},
	)
}

func writeChangePasswordError(w http.ResponseWriter, err error) {
	writeMappedUseCaseError(w, err, nil,
		errorRule{err: authuc.ErrUnauthorized, status: http.StatusUnauthorized},
		errorRule{err: authuc.ErrInvalidEmailOrPassword, status: http.StatusUnauthorized},
		errorRule{err: authuc.ErrIncorrectFormat, status: http.StatusBadRequest},
		errorRule{err: authuc.ErrWeakPassword, status: http.StatusBadRequest},
		errorRule{err: authuc.ErrUserNotFound, status: http.StatusNotFound},
		errorRule{err: authuc.ErrUserBlocked, status: http.StatusForbidden},
	)
}

func writeRefreshError(w http.ResponseWriter, err error) {
	writeMappedUseCaseError(w, err, nil,
		errorRule{err: authuc.ErrIncorrectFormat, status: http.StatusBadRequest},
		errorRule{err: authuc.ErrUnauthorized, status: http.StatusUnauthorized},
		errorRule{err: authuc.ErrRefreshMismatch, status: http.StatusUnauthorized},
		errorRule{err: authuc.ErrSessionExpired, status: http.StatusUnauthorized},
		errorRule{err: authuc.ErrSessionRevoked, status: http.StatusUnauthorized},
		errorRule{err: authuc.ErrSessionNotFound, status: http.StatusNotFound},
	)
}

func writePasswordResetRequestError(w http.ResponseWriter, err error) {
	writeMappedUseCaseError(w, err, nil,
		errorRule{err: authuc.ErrIncorrectFormat, status: http.StatusBadRequest},
		errorRule{err: authuc.ErrCannotResendYet, status: http.StatusTooManyRequests},
		errorRule{err: authuc.ErrVerificationFailed, status: http.StatusServiceUnavailable},
	)
}

func writePasswordResetVerifyError(w http.ResponseWriter, err error) {
	writeMappedUseCaseError(w, err, nil,
		errorRule{err: authuc.ErrIncorrectFormat, status: http.StatusBadRequest},
		errorRule{err: authuc.ErrVerificationCodeInvalid, status: http.StatusUnprocessableEntity},
		errorRule{err: authuc.ErrVerificationCodeExpired, status: http.StatusUnprocessableEntity},
		errorRule{err: authuc.ErrVerificationTooMany, status: http.StatusUnprocessableEntity},
		errorRule{err: authuc.ErrUserBlocked, status: http.StatusForbidden},
	)
}

func writePasswordResetConfirmError(w http.ResponseWriter, err error) {
	writeMappedUseCaseError(w, err, nil,
		errorRule{err: authuc.ErrIncorrectFormat, status: http.StatusBadRequest},
		errorRule{err: authuc.ErrUnauthorized, status: http.StatusUnauthorized},
		errorRule{err: authuc.ErrUserNotFound, status: http.StatusNotFound},
		errorRule{err: authuc.ErrUserBlocked, status: http.StatusForbidden},
	)
}
