package usecase

import (
	"pet/internal/application/ports"
	"pet/internal/domain/model"
	"time"

	"github.com/google/uuid"
)

const ActionPetRead = "pet_read"
const ActionPetWrite = "pet_write"
const MaxPetPhotoSizeBytes int64 = 20 * 1024 * 1024

type (
	UploadInfo                 = ports.UploadInfo
	ACLPolicy                  = ports.ACLPolicy
	ACLRole                    = ports.ACLRole
	ACLMembership              = ports.ACLMembership
	ACLTransferOwnershipResult = ports.ACLTransferOwnershipResult
)

type Pet struct {
	repo ports.PetRepository
	acl  ports.ACLClient
	file ports.FileClient
}

func New(repo ports.PetRepository, acl ports.ACLClient, file ports.FileClient) *Pet {
	return &Pet{repo: repo, acl: acl, file: file}
}

type CreatePetParams struct {
	UserID               uuid.UUID
	Name                 string
	SpeciesID            *uuid.UUID
	CustomSpeciesName    *string
	Sex                  string
	BirthDate            *time.Time
	BreedID              *uuid.UUID
	CustomBreedName      *string
	Colors               []model.Color
	PatternID            *uuid.UUID
	CustomPatternName    *string
	IsNeutered           string
	IsOutdoor            bool
	MicrochipID          *string
	MicrochipInstalledAt *time.Time
}

type ListPetsParams struct {
	UserID          uuid.UUID
	IncludeArchived bool
	Offset          int
	Limit           int
}

type UpdatePetParams struct {
	UserID               uuid.UUID
	PetID                uuid.UUID
	RowVersion           int
	Name                 string
	SpeciesID            *uuid.UUID
	CustomSpeciesName    *string
	Sex                  string
	BirthDate            *time.Time
	BreedID              *uuid.UUID
	CustomBreedName      *string
	Colors               []model.Color
	PatternID            *uuid.UUID
	CustomPatternName    *string
	IsNeutered           string
	IsOutdoor            bool
	MicrochipID          *string
	MicrochipInstalledAt *time.Time
}

type ChangePetStatusParams struct {
	UserID       uuid.UUID
	PetID        uuid.UUID
	RowVersion   int
	Status       string
	MissingSince *time.Time
}

type TransferPetOwnershipParams struct {
	UserID         uuid.UUID
	PetID          uuid.UUID
	RowVersion     int
	TargetMemberID uuid.UUID
}

type InitPetPhotoUploadParams struct {
	UserID            uuid.UUID
	PetID             uuid.UUID
	MimeType          string
	OriginalFilename  string
	ExpectedSizeBytes int64
}

type ConfirmPetPhotoUploadParams struct {
	UserID     uuid.UUID
	PetID      uuid.UUID
	RowVersion int
	FileID     uuid.UUID
	SizeBytes  int64
}

type DeletePetPhotoParams struct {
	UserID     uuid.UUID
	PetID      uuid.UUID
	RowVersion int
}

type PetListItem struct {
	Pet                     model.Pet
	MyAccess                *ACLMembership
	ProfilePhotoDownloadURL *string
}

type PetBrief struct {
	PetID     uuid.UUID
	Name      string
	AvatarURL *string
}
