package usecase

import (
	"context"
	"pet/internal/application/ports"
	"pet/internal/domain/model"
	"strings"

	"github.com/google/uuid"
)

func (s *Pet) InitPetPhotoUpload(ctx context.Context, p InitPetPhotoUploadParams) (uuid.UUID, UploadInfo, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || strings.TrimSpace(p.MimeType) == "" || p.ExpectedSizeBytes <= 0 || p.ExpectedSizeBytes > MaxPetPhotoSizeBytes {
		return uuid.Nil, UploadInfo{}, ErrInvalidInput
	}
	if !isAllowedPetPhotoMimeType(p.MimeType) {
		return uuid.Nil, UploadInfo{}, ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionPetWrite)
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
		if err == ports.ErrNotFound {
			return uuid.Nil, UploadInfo{}, ErrNotFound
		}
		return uuid.Nil, UploadInfo{}, err
	}
	if pet.Status == "ARCHIVED" {
		return uuid.Nil, UploadInfo{}, ErrConflict
	}

	fileID, upload, err := s.file.InitUpload(ctx, strings.TrimSpace(strings.ToLower(p.MimeType)), p.ExpectedSizeBytes, p.OriginalFilename)
	if err != nil {
		return uuid.Nil, UploadInfo{}, err
	}

	return fileID, upload, nil
}

func (s *Pet) ConfirmPetPhotoUpload(ctx context.Context, p ConfirmPetPhotoUploadParams) (*model.Pet, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.RowVersion <= 0 || p.FileID == uuid.Nil || p.SizeBytes <= 0 || p.SizeBytes > MaxPetPhotoSizeBytes {
		return nil, ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionPetWrite)
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
		if err == ports.ErrNotFound {
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

	var oldPhotoID *uuid.UUID
	if pet.ProfilePhotoFileID != nil {
		oldID := *pet.ProfilePhotoFileID
		oldPhotoID = &oldID
	}

	if err := s.file.LinkPetAvatar(ctx, p.FileID, p.PetID); err != nil {
		return nil, err
	}

	updated, err := s.repo.UpdatePhoto(ctx, p.PetID, p.RowVersion, &p.FileID)
	if err != nil {
		_ = s.file.UnlinkPetAvatar(ctx, p.FileID, p.PetID)
		switch err {
		case ports.ErrNotFound:
			return nil, ErrNotFound
		case ports.ErrConflict:
			return nil, ErrConflict
		default:
			return nil, err
		}
	}

	if oldPhotoID != nil && *oldPhotoID != p.FileID {
		_ = s.file.UnlinkPetAvatar(ctx, *oldPhotoID, p.PetID)
		_ = s.file.DeleteFileIfUnlinked(ctx, *oldPhotoID)
	}

	return updated, nil
}

func (s *Pet) DeletePetPhoto(ctx context.Context, p DeletePetPhotoParams) (*model.Pet, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.RowVersion <= 0 {
		return nil, ErrInvalidInput
	}

	allowed, err := s.acl.Check(ctx, p.PetID, p.UserID, ActionPetWrite)
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
		if err == ports.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if pet.Status == "ARCHIVED" {
		return nil, ErrConflict
	}
	if pet.ProfilePhotoFileID == nil {
		return pet, nil
	}

	oldPhotoID := *pet.ProfilePhotoFileID
	updated, err := s.repo.UpdatePhoto(ctx, p.PetID, p.RowVersion, nil)
	if err != nil {
		switch err {
		case ports.ErrNotFound:
			return nil, ErrNotFound
		case ports.ErrConflict:
			return nil, ErrConflict
		default:
			return nil, err
		}
	}

	_ = s.file.UnlinkPetAvatar(ctx, oldPhotoID, p.PetID)
	_ = s.file.DeleteFileIfUnlinked(ctx, oldPhotoID)

	return updated, nil
}
