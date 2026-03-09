package service

import (
	"context"
	"pet/internal/model"
	"pet/internal/repository"
	"strings"
	"time"

	"github.com/google/uuid"
)

const ActionPetRead = "pet_read"
const ActionPetEdit = "pet_edit"
const ActionPetStatusChange = "pet_status_change"
const MaxPetPhotoSizeBytes int64 = 20 * 1024 * 1024

type ACLClient interface {
	Check(ctx context.Context, petID, userID uuid.UUID, action string) (bool, error)
	ListPetsForUser(ctx context.Context, userID uuid.UUID) ([]ACLMembership, error)
	CreateOwnerMembership(ctx context.Context, petID, userID uuid.UUID) (uuid.UUID, error)
}

type FileClient interface {
	InitUpload(ctx context.Context, mimeType string, expectedSize int64) (uuid.UUID, UploadInfo, error)
	ConfirmUpload(ctx context.Context, fileID uuid.UUID, sizeBytes int64) error
	GetDownloadURL(ctx context.Context, fileID uuid.UUID) (string, time.Time, error)
	BatchGetDownloadURLs(ctx context.Context, fileIDs []uuid.UUID) (map[uuid.UUID]string, error)
	LinkPetAvatar(ctx context.Context, fileID, petID uuid.UUID) error
}

type PetService struct {
	repo repository.PetRepository
	acl  ACLClient
	file FileClient
}

func New(repo repository.PetRepository, acl ACLClient, file FileClient) *PetService {
	return &PetService{repo: repo, acl: acl, file: file}
}

type CreatePetParams struct {
	UserID               uuid.UUID
	Name                 string
	SpeciesID            uuid.UUID
	Sex                  string
	BirthDate            *time.Time
	Breed                model.Breed
	Colors               []model.Color
	CoatPattern          model.CoatPattern
	IsNeutered           string
	IsOutdoor            bool
	ProfilePhotoFileID   *uuid.UUID
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
	SpeciesID            uuid.UUID
	Sex                  string
	BirthDate            *time.Time
	Breed                model.Breed
	Colors               []model.Color
	CoatPattern          model.CoatPattern
	IsNeutered           string
	IsOutdoor            bool
	ProfilePhotoFileID   *uuid.UUID
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

type UploadInfo struct {
	Method    string
	URL       string
	Headers   map[string]string
	ExpiresAt time.Time
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

type ACLPolicy struct {
	PetRead                bool
	PetEdit                bool
	PetStatusChange        bool
	PetDelete              bool
	LogRead                bool
	LogCreate              bool
	LogEdit                bool
	LogDelete              bool
	LogAttachmentsRead     bool
	HealthRead             bool
	HealthWrite            bool
	TaskRead               bool
	TaskWrite              bool
	MembersView            bool
	MembersInvite          bool
	MembersRemove          bool
	MembersEditPermissions bool
}

type ACLRole struct {
	ID              uuid.UUID
	Kind            string
	PetID           *uuid.UUID
	Code            *string
	Title           string
	CreatedByUserID *uuid.UUID
}

type ACLMembership struct {
	PetID          uuid.UUID
	MemberID       uuid.UUID
	Status         string
	IsPrimaryOwner bool
	Role           ACLRole
	Policy         ACLPolicy
}

type PetListItem struct {
	Pet                     model.Pet
	MyAccess                *ACLMembership
	ProfilePhotoDownloadURL *string
}

func (s *PetService) CreatePet(ctx context.Context, p CreatePetParams) (*model.Pet, error) {
	if p.UserID == uuid.Nil || p.SpeciesID == uuid.Nil || strings.TrimSpace(p.Name) == "" {
		return nil, ErrInvalidInput
	}

	pet := model.Pet{
		ID:                   uuid.New(),
		OwnerUserID:          p.UserID,
		Name:                 strings.TrimSpace(p.Name),
		SpeciesID:            p.SpeciesID,
		Sex:                  p.Sex,
		BirthDate:            p.BirthDate,
		Breed:                p.Breed,
		Colors:               p.Colors,
		CoatPattern:          p.CoatPattern,
		IsNeutered:           p.IsNeutered,
		IsOutdoor:            p.IsOutdoor,
		ProfilePhotoFileID:   p.ProfilePhotoFileID,
		MicrochipID:          p.MicrochipID,
		MicrochipInstalledAt: p.MicrochipInstalledAt,
		Status:               "ACTIVE",
	}

	created, err := s.repo.Create(ctx, repository.CreatePetInput{Pet: pet})
	if err != nil {
		if err == repository.ErrConflict {
			return nil, ErrConflict
		}
		return nil, err
	}

	if _, err := s.acl.CreateOwnerMembership(ctx, created.ID, p.UserID); err != nil {
		_ = s.repo.DeleteByID(ctx, created.ID)
		if err == ErrConflict {
			return nil, ErrConflict
		}
		return nil, err
	}

	return created, nil
}

func (s *PetService) ListPets(ctx context.Context, p ListPetsParams) ([]PetListItem, int, error) {
	if p.UserID == uuid.Nil {
		return nil, 0, ErrInvalidInput
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	if p.Limit <= 0 {
		p.Limit = 50
	}
	if p.Limit > 100 {
		p.Limit = 100
	}

	memberships, err := s.acl.ListPetsForUser(ctx, p.UserID)
	if err != nil {
		return nil, 0, err
	}
	petIDs := make([]uuid.UUID, 0, len(memberships))
	accessByPet := make(map[uuid.UUID]ACLMembership, len(memberships))
	for i := range memberships {
		member := memberships[i]
		petIDs = append(petIDs, member.PetID)
		if member.MemberID != uuid.Nil {
			accessByPet[member.PetID] = member
		}
	}

	items, total, err := s.repo.ListByIDs(ctx, petIDs, p.IncludeArchived, p.Offset, p.Limit)
	if err != nil {
		return nil, 0, err
	}

	photoIDs := make([]uuid.UUID, 0, len(items))
	for i := range items {
		if items[i].ProfilePhotoFileID != nil {
			photoIDs = append(photoIDs, *items[i].ProfilePhotoFileID)
		}
	}

	photoURLs := make(map[uuid.UUID]string, len(photoIDs))
	if len(photoIDs) > 0 {
		if urls, err := s.file.BatchGetDownloadURLs(ctx, photoIDs); err == nil {
			photoURLs = urls
		}
	}

	out := make([]PetListItem, 0, len(items))
	for i := range items {
		item := PetListItem{
			Pet:      items[i],
			MyAccess: nil,
		}
		if access, ok := accessByPet[items[i].ID]; ok {
			accessCopy := access
			item.MyAccess = &accessCopy
		}
		if items[i].ProfilePhotoFileID != nil {
			if url, ok := photoURLs[*items[i].ProfilePhotoFileID]; ok && strings.TrimSpace(url) != "" {
				urlCopy := url
				item.ProfilePhotoDownloadURL = &urlCopy
			}
		}
		out = append(out, item)
	}

	return out, total, nil
}

func (s *PetService) ResolveProfilePhotoDownloadURL(ctx context.Context, fileID *uuid.UUID) *string {
	if fileID == nil {
		return nil
	}
	url, _, err := s.file.GetDownloadURL(ctx, *fileID)
	if err != nil || strings.TrimSpace(url) == "" {
		return nil
	}
	return &url
}

func (s *PetService) GetPet(ctx context.Context, userID, petID uuid.UUID) (*model.Pet, error) {
	if userID == uuid.Nil || petID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, petID, userID, ActionPetRead)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrForbidden
		}
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	pet, err := s.repo.GetByID(ctx, petID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return pet, nil
}

func (s *PetService) UpdatePet(ctx context.Context, p UpdatePetParams) (*model.Pet, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.RowVersion <= 0 || p.SpeciesID == uuid.Nil || strings.TrimSpace(p.Name) == "" {
		return nil, ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionPetEdit)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrForbidden
		}
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	current, err := s.repo.GetByID(ctx, p.PetID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if current.Status == "ARCHIVED" {
		return nil, ErrConflict
	}

	updated, err := s.repo.Update(ctx, p.PetID, p.RowVersion, model.Pet{
		Name:                 strings.TrimSpace(p.Name),
		SpeciesID:            p.SpeciesID,
		Sex:                  p.Sex,
		BirthDate:            p.BirthDate,
		Breed:                p.Breed,
		Colors:               p.Colors,
		CoatPattern:          p.CoatPattern,
		IsNeutered:           p.IsNeutered,
		IsOutdoor:            p.IsOutdoor,
		ProfilePhotoFileID:   p.ProfilePhotoFileID,
		MicrochipID:          p.MicrochipID,
		MicrochipInstalledAt: p.MicrochipInstalledAt,
	})
	if err != nil {
		switch err {
		case repository.ErrNotFound:
			return nil, ErrNotFound
		case repository.ErrConflict:
			return nil, ErrConflict
		default:
			return nil, err
		}
	}
	return updated, nil
}

func (s *PetService) ChangePetStatus(ctx context.Context, p ChangePetStatusParams) (*model.Pet, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.RowVersion <= 0 {
		return nil, ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionPetStatusChange)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrForbidden
		}
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	status := strings.ToUpper(strings.TrimSpace(p.Status))
	var (
		missingSince *time.Time
		archivedAt   *time.Time
	)
	switch status {
	case "ACTIVE":
	case "MISSING":
		if p.MissingSince == nil {
			return nil, ErrInvalidInput
		}
		missingSince = p.MissingSince
	case "ARCHIVED":
		now := time.Now().UTC()
		archivedAt = &now
	default:
		return nil, ErrInvalidInput
	}

	pet, err := s.repo.UpdateStatus(ctx, p.PetID, p.RowVersion, status, missingSince, archivedAt)
	if err != nil {
		switch err {
		case repository.ErrNotFound:
			return nil, ErrNotFound
		case repository.ErrConflict:
			return nil, ErrConflict
		default:
			return nil, err
		}
	}
	return pet, nil
}

func (s *PetService) InitPetPhotoUpload(ctx context.Context, p InitPetPhotoUploadParams) (uuid.UUID, UploadInfo, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || strings.TrimSpace(p.MimeType) == "" || p.ExpectedSizeBytes <= 0 || p.ExpectedSizeBytes > MaxPetPhotoSizeBytes {
		return uuid.Nil, UploadInfo{}, ErrInvalidInput
	}
	if !isAllowedPetPhotoMimeType(p.MimeType) {
		return uuid.Nil, UploadInfo{}, ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionPetEdit)
	if err != nil {
		if err == ErrNotFound {
			return uuid.Nil, UploadInfo{}, ErrForbidden
		}
		return uuid.Nil, UploadInfo{}, err
	}
	if !allowed {
		return uuid.Nil, UploadInfo{}, ErrForbidden
	}

	pet, err := s.repo.GetByID(ctx, p.PetID)
	if err != nil {
		if err == repository.ErrNotFound {
			return uuid.Nil, UploadInfo{}, ErrNotFound
		}
		return uuid.Nil, UploadInfo{}, err
	}
	if pet.Status == "ARCHIVED" {
		return uuid.Nil, UploadInfo{}, ErrConflict
	}

	fileID, upload, err := s.file.InitUpload(ctx, strings.TrimSpace(strings.ToLower(p.MimeType)), p.ExpectedSizeBytes)
	if err != nil {
		return uuid.Nil, UploadInfo{}, err
	}
	return fileID, upload, nil
}

func (s *PetService) ConfirmPetPhotoUpload(ctx context.Context, p ConfirmPetPhotoUploadParams) (*model.Pet, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.RowVersion <= 0 || p.FileID == uuid.Nil || p.SizeBytes <= 0 || p.SizeBytes > MaxPetPhotoSizeBytes {
		return nil, ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionPetEdit)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrForbidden
		}
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}

	pet, err := s.repo.GetByID(ctx, p.PetID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if pet.Status == "ARCHIVED" {
		return nil, ErrConflict
	}

	if err := s.file.ConfirmUpload(ctx, p.FileID, p.SizeBytes); err != nil {
		return nil, err
	}
	if err := s.file.LinkPetAvatar(ctx, p.FileID, p.PetID); err != nil {
		return nil, err
	}

	updated, err := s.repo.UpdatePhoto(ctx, p.PetID, p.RowVersion, p.FileID)
	if err != nil {
		switch err {
		case repository.ErrNotFound:
			return nil, ErrNotFound
		case repository.ErrConflict:
			return nil, ErrConflict
		default:
			return nil, err
		}
	}
	return updated, nil
}

func isAllowedPetPhotoMimeType(raw string) bool {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "image/jpeg", "image/png":
		return true
	default:
		return false
	}
}
