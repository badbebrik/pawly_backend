package model

type SpeciesItem struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	Version  int    `json:"version"`
}
