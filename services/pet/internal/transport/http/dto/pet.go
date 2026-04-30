package dto

import "github.com/google/uuid"

type CreatePetRequest struct {
	Name                 string         `json:"name"`
	SpeciesID            *string        `json:"species_id"`
	CustomSpeciesName    *string        `json:"custom_species_name"`
	Sex                  string         `json:"sex"`
	BirthDate            *string        `json:"birth_date"`
	BreedID              *string        `json:"breed_id"`
	CustomBreedName      *string        `json:"custom_breed_name"`
	Colors               []ColorRequest `json:"colors"`
	PatternID            *string        `json:"pattern_id"`
	CustomPatternName    *string        `json:"custom_pattern_name"`
	IsNeutered           string         `json:"is_neutered"`
	IsOutdoor            bool           `json:"is_outdoor"`
	MicrochipID          *string        `json:"microchip_id"`
	MicrochipInstalledAt *string        `json:"microchip_installed_at"`
}

type UpdatePetRequest struct {
	RowVersion           int            `json:"row_version"`
	Name                 string         `json:"name"`
	SpeciesID            *string        `json:"species_id"`
	CustomSpeciesName    *string        `json:"custom_species_name"`
	Sex                  string         `json:"sex"`
	BirthDate            *string        `json:"birth_date"`
	BreedID              *string        `json:"breed_id"`
	CustomBreedName      *string        `json:"custom_breed_name"`
	Colors               []ColorRequest `json:"colors"`
	PatternID            *string        `json:"pattern_id"`
	CustomPatternName    *string        `json:"custom_pattern_name"`
	IsNeutered           string         `json:"is_neutered"`
	IsOutdoor            bool           `json:"is_outdoor"`
	MicrochipID          *string        `json:"microchip_id"`
	MicrochipInstalledAt *string        `json:"microchip_installed_at"`
}

type ColorRequest struct {
	PresetID   *string `json:"preset_id"`
	CustomName *string `json:"custom_name"`
	CustomHex  *string `json:"custom_hex"`
	SortOrder  int     `json:"sort_order"`
}

type ChangePetStatusRequest struct {
	RowVersion   int     `json:"row_version"`
	Status       string  `json:"status"`
	MissingSince *string `json:"missing_since"`
}

type TransferOwnershipRequest struct {
	RowVersion     int    `json:"row_version"`
	TargetMemberID string `json:"target_member_id"`
}

type InitPetPhotoUploadRequest struct {
	MimeType          string `json:"mime_type"`
	OriginalFilename  string `json:"original_filename"`
	ExpectedSizeBytes int64  `json:"expected_size_bytes"`
}

type ConfirmPetPhotoUploadRequest struct {
	RowVersion int    `json:"row_version"`
	FileID     string `json:"file_id"`
	SizeBytes  int64  `json:"size_bytes"`
}

type DeletePetPhotoRequest struct {
	RowVersion int `json:"row_version"`
}

type ColorResponse struct {
	ID         *uuid.UUID `json:"id"`
	PresetID   *uuid.UUID `json:"preset_id"`
	CustomName *string    `json:"custom_name"`
	CustomHex  *string    `json:"custom_hex"`
	SortOrder  int        `json:"sort_order"`
}

type ACLRoleResponse struct {
	ID              *uuid.UUID        `json:"id"`
	Kind            string            `json:"kind"`
	PetID           *uuid.UUID        `json:"pet_id"`
	Code            *string           `json:"code"`
	Title           string            `json:"title"`
	Policy          ACLPolicyResponse `json:"policy"`
	CreatedByUserID *uuid.UUID        `json:"created_by_user_id"`
}

type ACLPolicyResponse struct {
	PetRead      bool `json:"pet_read"`
	PetWrite     bool `json:"pet_write"`
	LogRead      bool `json:"log_read"`
	LogWrite     bool `json:"log_write"`
	HealthRead   bool `json:"health_read"`
	HealthWrite  bool `json:"health_write"`
	MembersRead  bool `json:"members_read"`
	MembersWrite bool `json:"members_write"`
}

type ACLMembershipResponse struct {
	MemberID       uuid.UUID         `json:"member_id"`
	Status         string            `json:"status"`
	IsPrimaryOwner bool              `json:"is_primary_owner"`
	Role           ACLRoleResponse   `json:"role"`
	Policy         ACLPolicyResponse `json:"policy"`
}

type PetResponse struct {
	ID                      uuid.UUID       `json:"id"`
	OwnerUserID             uuid.UUID       `json:"owner_user_id"`
	RowVersion              int             `json:"row_version"`
	Name                    string          `json:"name"`
	SpeciesID               *uuid.UUID      `json:"species_id"`
	CustomSpeciesName       *string         `json:"custom_species_name"`
	Sex                     string          `json:"sex"`
	BirthDate               *string         `json:"birth_date"`
	BreedID                 *uuid.UUID      `json:"breed_id"`
	CustomBreedName         *string         `json:"custom_breed_name"`
	Colors                  []ColorResponse `json:"colors"`
	PatternID               *uuid.UUID      `json:"pattern_id"`
	CustomPatternName       *string         `json:"custom_pattern_name"`
	IsNeutered              string          `json:"is_neutered"`
	IsOutdoor               bool            `json:"is_outdoor"`
	ProfilePhotoFileID      *uuid.UUID      `json:"profile_photo_file_id"`
	ProfilePhotoDownloadURL *string         `json:"profile_photo_download_url"`
	MicrochipID             *string         `json:"microchip_id"`
	MicrochipInstalledAt    *string         `json:"microchip_installed_at"`
	Status                  string          `json:"status"`
	MissingSince            *string         `json:"missing_since"`
	ArchivedAt              *string         `json:"archived_at"`
	CreatedAt               string          `json:"created_at"`
	UpdatedAt               string          `json:"updated_at"`
}

type PetEnvelopeResponse struct {
	Pet PetResponse `json:"pet"`
}

type PetListItemResponse struct {
	Pet      PetResponse            `json:"pet"`
	MyAccess *ACLMembershipResponse `json:"my_access"`
}

type ListPetsResponse struct {
	Items  []PetListItemResponse `json:"items"`
	Total  int                   `json:"total"`
	Offset int                   `json:"offset"`
	Limit  int                   `json:"limit"`
}

type UploadInfoResponse struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	ExpiresAt string            `json:"expires_at"`
}

type InitPetPhotoUploadResponse struct {
	FileID uuid.UUID          `json:"file_id"`
	Upload UploadInfoResponse `json:"upload"`
}
