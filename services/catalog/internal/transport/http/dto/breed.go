package dto

type BreedItem struct {
	ID        string `json:"id"`
	SpeciesID string `json:"species_id"`
	Name      string `json:"name"`
	IsActive  bool   `json:"is_active"`
	Version   int    `json:"version"`
}
