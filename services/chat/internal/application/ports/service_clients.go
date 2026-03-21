package ports

import (
	"context"

	"github.com/google/uuid"
)

type ProfileBrief struct {
	UserID      uuid.UUID
	DisplayName *string
	AvatarURL   *string
}

type PetBrief struct {
	PetID     uuid.UUID
	Name      string
	AvatarURL *string
}

type ACLClient interface {
	IsActiveMember(ctx context.Context, petID, userID uuid.UUID) (bool, error)
}

type ProfileClient interface {
	BatchGetBrief(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]ProfileBrief, error)
}

type PetClient interface {
	BatchGetBrief(ctx context.Context, petIDs []uuid.UUID) (map[uuid.UUID]PetBrief, error)
}
