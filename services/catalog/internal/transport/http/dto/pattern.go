package dto

type PatternItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IconKey  string `json:"icon_key"`
	IsActive bool   `json:"is_active"`
	Version  int    `json:"version"`
}

type CreatePatternRequest struct {
	NameRu   string `json:"name_ru"`
	NameEn   string `json:"name_en"`
	IconKey  string `json:"icon_key"`
	IsActive *bool  `json:"is_active"`
}

type UpdatePatternRequest struct {
	NameRu   *string `json:"name_ru"`
	NameEn   *string `json:"name_en"`
	IconKey  *string `json:"icon_key"`
	IsActive *bool   `json:"is_active"`
}
