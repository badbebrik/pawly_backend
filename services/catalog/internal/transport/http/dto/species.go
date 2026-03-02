package dto

type CreateSpeciesRequest struct {
	NameRu   string `json:"name_ru"`
	NameEn   string `json:"name_en"`
	IsActive *bool  `json:"is_active"`
}

type UpdateSpeciesRequest struct {
	NameRu   *string `json:"name_ru"`
	NameEn   *string `json:"name_en"`
	IsActive *bool   `json:"is_active"`
}

type SpeciesItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	Version  int    `json:"version"`
}
