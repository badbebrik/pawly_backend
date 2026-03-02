package dto

type ColorItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Hex      string `json:"hex"`
	IsActive bool   `json:"is_active"`
	Version  int    `json:"version"`
}

type CreateColorRequest struct {
	NameRu   string `json:"name_ru"`
	NameEn   string `json:"name_en"`
	Hex      string `json:"hex"`
	IsActive *bool  `json:"is_active"`
}

type UpdateColorRequest struct {
	NameRu   *string `json:"name_ru"`
	NameEn   *string `json:"name_en"`
	Hex      *string `json:"hex"`
	IsActive *bool   `json:"is_active"`
}
