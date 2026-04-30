package dto

type PasswordResetRequest struct {
	Email string `json:"email"`
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
