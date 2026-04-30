package ports

import (
	"context"

	"github.com/google/uuid"
)

type ProfileBrief struct {
	UserID            uuid.UUID
	FirstName         *string
	LastName          *string
	DisplayName       *string
	AvatarDownloadURL *string
}

type ProfileClient interface {
	BatchProfilesBrief(ctx context.Context, userIDs []uuid.UUID) ([]ProfileBrief, []uuid.UUID, error)
}
