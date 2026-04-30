package model

type Policy struct {
	PetRead      bool `json:"pet_read"`
	PetWrite     bool `json:"pet_write"`
	LogRead      bool `json:"log_read"`
	LogWrite     bool `json:"log_write"`
	HealthRead   bool `json:"health_read"`
	HealthWrite  bool `json:"health_write"`
	MembersRead  bool `json:"members_read"`
	MembersWrite bool `json:"members_write"`
}
