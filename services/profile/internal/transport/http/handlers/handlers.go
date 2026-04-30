package handlers

import (
	profileuc "profile/internal/application/usecase"
)

type Handlers struct {
	useCases *profileuc.Set
}

func New(useCases *profileuc.Set) *Handlers {
	return &Handlers{useCases: useCases}
}
