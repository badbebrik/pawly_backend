package model

import (
	"time"

	"github.com/google/uuid"
)

type Pattern struct {
	ID        uuid.UUID
	NameRu    string
	NameEn    string
	IconKey   string
	IsActive  bool
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}
