package model

import (
	"time"

	"github.com/google/uuid"
)

type Pet struct {
	ID                   uuid.UUID
	OwnerUserID          uuid.UUID
	RowVersion           int
	Name                 string
	SpeciesID            uuid.UUID
	Sex                  string
	BirthDate            *time.Time
	BreedID              *uuid.UUID
	CustomBreedName      *string
	PatternID            *uuid.UUID
	CustomPatternName    *string
	Colors               []Color
	IsNeutered           string
	IsOutdoor            bool
	ProfilePhotoFileID   *uuid.UUID
	MicrochipID          *string
	MicrochipInstalledAt *time.Time
	Status               string
	MissingSince         *time.Time
	ArchivedAt           *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
