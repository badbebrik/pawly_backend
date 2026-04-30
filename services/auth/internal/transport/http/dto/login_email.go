package dto

type LoginEmailRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
