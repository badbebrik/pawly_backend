package handlers

import (
	authuc "auth/internal/application/usecase"
	"auth/internal/transport/http/dto"

	"github.com/google/uuid"
)

func sessionResponse(userID uuid.UUID, accessToken, refreshToken string, expiresIn int) dto.SessionResponse {
	return dto.SessionResponse{
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}
}

func verificationResponse(channel string, codeTTLSeconds, canResendInSeconds int) dto.VerificationResponse {
	return dto.VerificationResponse{
		Channel:            channel,
		CodeTTLSeconds:     codeTTLSeconds,
		CanResendInSeconds: canResendInSeconds,
	}
}

func registerEmailResponse(result *authuc.RegisterEmailResult) dto.RegisterEmailResponse {
	return dto.RegisterEmailResponse{
		UserID: result.UserID,
		Verification: verificationResponse(
			result.Verification.Channel,
			result.Verification.CodeTTLSeconds,
			result.Verification.CanResendInSeconds,
		),
	}
}

func resendEmailVerificationResponse(result *authuc.ResendEmailVerificationResult) dto.ResendEmailVerificationResponse {
	return dto.ResendEmailVerificationResponse{
		UserID: result.UserID,
		Verification: verificationResponse(
			result.Verification.Channel,
			result.Verification.CodeTTLSeconds,
			result.Verification.CanResendInSeconds,
		),
	}
}

func statusResponse() dto.StatusResponse {
	return dto.StatusResponse{Status: "ok"}
}

func passwordResetVerifyResponse(result *authuc.PasswordResetVerifyResult) dto.PasswordResetVerifyResponse {
	return dto.PasswordResetVerifyResponse{ResetToken: result.ResetToken}
}
