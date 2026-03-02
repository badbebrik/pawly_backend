package model

import (
	"time"

	"github.com/google/uuid"
)

type Breed struct {
	ID        uuid.UUID
	SpeciesID uuid.UUID
	NameRu    string
	NameEn    string
	IsActive  bool
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}
