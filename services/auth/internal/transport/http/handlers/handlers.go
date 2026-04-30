package handlers

import authuc "auth/internal/application/usecase"

type Handlers struct {
	useCases *authuc.Set
}

func New(useCases *authuc.Set) *Handlers {
	return &Handlers{useCases: useCases}
}
