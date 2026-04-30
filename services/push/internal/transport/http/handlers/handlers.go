package handlers

import pushuc "push/internal/application/usecase"

type Handlers struct {
	useCases *pushuc.Set
}

func New(useCases *pushuc.Set) *Handlers {
	return &Handlers{useCases: useCases}
}
