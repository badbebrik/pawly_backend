package model

import (
	"time"

	"github.com/google/uuid"
)

type Species struct {
	ID        uuid.UUID
	Code      string
	Name      string
	IconKey   string
	SortOrder int
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Breed struct {
	ID        uuid.UUID
	SpeciesID uuid.UUID
	Name      string
	SortOrder int
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Pattern struct {
	ID        uuid.UUID
	SpeciesID *uuid.UUID
	Name      string
	IconKey   *string
	SortOrder int
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ColorPreset struct {
	ID        uuid.UUID
	Name      string
	Hex       string
	SortOrder int
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Color struct {
	ID         uuid.UUID
	PetID      uuid.UUID
	SortOrder  int
	PresetID   *uuid.UUID
	CustomName *string
	CustomHex  *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Dictionaries struct {
	Species      []Species
	Breeds       []Breed
	Patterns     []Pattern
	ColorPresets []ColorPreset
}
