package model

type Policy struct {
	PetRead                bool `json:"pet_read"`
	PetEdit                bool `json:"pet_edit"`
	PetStatusChange        bool `json:"pet_status_change"`
	PetDelete              bool `json:"pet_delete"`
	LogRead                bool `json:"log_read"`
	LogCreate              bool `json:"log_create"`
	LogEdit                bool `json:"log_edit"`
	LogDelete              bool `json:"log_delete"`
	LogAttachmentsRead     bool `json:"log_attachments_read"`
	HealthRead             bool `json:"health_read"`
	HealthWrite            bool `json:"health_write"`
	TaskRead               bool `json:"task_read"`
	TaskWrite              bool `json:"task_write"`
	MembersView            bool `json:"members_view"`
	MembersInvite          bool `json:"members_invite"`
	MembersRemove          bool `json:"members_remove"`
	MembersEditPermissions bool `json:"members_edit_permissions"`
}
