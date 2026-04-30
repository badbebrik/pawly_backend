package ports

import (
	"context"

	"github.com/google/uuid"
)

type PetBrief struct {
	PetID            uuid.UUID
	Name             string
	PhotoDownloadURL *string
}

type PetClient interface {
	BatchGetBrief(ctx context.Context, petIDs []uuid.UUID) ([]PetBrief, []uuid.UUID, error)
}
