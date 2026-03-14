package handlers

import authuc "auth/internal/application/usecase"

type AuthHandlers struct {
	uc *authuc.Set
}

func NewAuthHandlers(uc *authuc.Set) *AuthHandlers {
	return &AuthHandlers{uc: uc}
}
