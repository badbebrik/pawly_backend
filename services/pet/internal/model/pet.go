package model

import (
	"time"

	"github.com/google/uuid"
)

type Breed struct {
	Source          string
	SystemBreedID   *uuid.UUID
	CustomBreedName *string
}

type CoatPattern struct {
	Source                string
	SystemCoatPatternID   *uuid.UUID
	CustomCoatPatternName *string
}

type Color struct {
	PresetID    *uuid.UUID `json:"preset_id"`
	HexOverride *string    `json:"hex_override"`
	Note        *string    `json:"note"`
	SortOrder   int        `json:"sort_order"`
}

type Pet struct {
	ID                   uuid.UUID
	OwnerUserID          uuid.UUID
	RowVersion           int
	Name                 string
	SpeciesID            uuid.UUID
	Sex                  string
	BirthDate            *time.Time
	Breed                Breed
	Colors               []Color
	CoatPattern          CoatPattern
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
