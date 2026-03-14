package usecase

import (
	"context"
	"profile/internal/application/ports"
	"profile/internal/model"
	"strings"

	"github.com/google/uuid"
)

type InitAvatarUploadUseCase struct{ deps *dependencies }
type ConfirmAvatarUploadUseCase struct{ deps *dependencies }
type GetAvatarDownloadURLUseCase struct{ deps *dependencies }

func (uc *InitAvatarUploadUseCase) Execute(ctx context.Context, userID uuid.UUID, mimeType string, expectedSize *int64) (uuid.UUID, ports.UploadInfo, error) {
	if userID == uuid.Nil || strings.TrimSpace(mimeType) == "" {
		return uuid.Nil, ports.UploadInfo{}, ErrInvalidInput
	}
	if uc.deps.files == nil {
		return uuid.Nil, ports.UploadInfo{}, ErrAvatarUpload
	}

	size := int64(0)
	if expectedSize != nil {
		size = *expectedSize
	}
	fileID, upload, err := uc.deps.files.InitUpload(ctx, mimeType, size, userID)
	if err != nil {
		return uuid.Nil, ports.UploadInfo{}, ErrAvatarUpload
	}
	return fileID, upload, nil
}

func (uc *ConfirmAvatarUploadUseCase) Execute(ctx context.Context, userID uuid.UUID, fileID uuid.UUID, sizeBytes int64) (*model.Profile, error) {
	if userID == uuid.Nil || fileID == uuid.Nil || sizeBytes <= 0 {
		return nil, ErrInvalidInput
	}
	if uc.deps.files == nil {
		return nil, ErrAvatarUpload
	}

	if err := uc.deps.files.ConfirmUpload(ctx, fileID, sizeBytes); err != nil {
		return nil, ErrAvatarUpload
	}
	if err := uc.deps.files.LinkAvatar(ctx, fileID, userID); err != nil {
		return nil, ErrAvatarUpload
	}

	profile, err := uc.deps.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	profile.AvatarFileID = &fileID
	if err := uc.deps.profiles.Update(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func (uc *GetAvatarDownloadURLUseCase) Execute(ctx context.Context, fileID uuid.UUID) (string, error) {
	if fileID == uuid.Nil {
		return "", ErrInvalidInput
	}
	if uc.deps.files == nil {
		return "", ErrAvatarUpload
	}
	url, _, err := uc.deps.files.GetDownloadURL(ctx, fileID)
	if err != nil {
		return "", ErrAvatarUpload
	}
	if strings.TrimSpace(url) == "" {
		return "", ErrAvatarUpload
	}
	return url, nil
}
