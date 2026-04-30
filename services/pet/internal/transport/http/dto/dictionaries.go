package dto

import "github.com/google/uuid"

type SpeciesResponse struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	IconKey   string    `json:"icon_key"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
}

type BreedResponse struct {
	ID        uuid.UUID `json:"id"`
	SpeciesID uuid.UUID `json:"species_id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
}

type PatternResponse struct {
	ID        uuid.UUID  `json:"id"`
	SpeciesID *uuid.UUID `json:"species_id"`
	Name      string     `json:"name"`
	IconKey   *string    `json:"icon_key"`
	SortOrder int        `json:"sort_order"`
	IsActive  bool       `json:"is_active"`
}

type ColorPresetResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Hex       string    `json:"hex"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
}

type DictionariesResponse struct {
	Species      []SpeciesResponse     `json:"species"`
	Breeds       []BreedResponse       `json:"breeds"`
	Patterns     []PatternResponse     `json:"patterns"`
	ColorPresets []ColorPresetResponse `json:"color_presets"`
}
