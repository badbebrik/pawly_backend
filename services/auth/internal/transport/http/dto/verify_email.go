package dto

type VerifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}
