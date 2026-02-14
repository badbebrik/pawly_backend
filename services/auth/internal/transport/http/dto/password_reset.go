package dto

type PasswordResetRequestRequest struct {
	Email string `json:"email"`
}

type PasswordResetRequestResponse struct {
	Status string `json:"status"`
}

type PasswordResetVerifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type PasswordResetVerifyResponse struct {
	ResetToken string `json:"reset_token"`
}

type PasswordResetConfirmRequest struct {
	ResetToken  string `json:"reset_token"`
	NewPassword string `json:"new_password"`
}

type PasswordResetConfirmResponse struct {
	Status string `json:"status"`
}
