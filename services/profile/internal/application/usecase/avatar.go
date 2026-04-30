package usecase

import (
	"context"
	"profile/internal/application/ports"
	"profile/internal/domain/model"
	"strings"

	"github.com/google/uuid"
)

type InitAvatarUpload struct{ deps *dependencies }
type ConfirmAvatarUpload struct{ deps *dependencies }
type DeleteAvatar struct{ deps *dependencies }
type GetAvatarDownloadURL struct{ deps *dependencies }

func (uc *InitAvatarUpload) Execute(ctx context.Context, userID uuid.UUID, mimeType string, expectedSize *int64) (uuid.UUID, ports.UploadInfo, error) {
	if userID == uuid.Nil || strings.TrimSpace(mimeType) == "" {
		return uuid.Nil, ports.UploadInfo{}, ErrInvalidInput
	}
	if uc.deps.fileClient == nil {
		return uuid.Nil, ports.UploadInfo{}, ErrAvatarUpload
	}

	size := int64(0)
	if expectedSize != nil {
		size = *expectedSize
	}
	fileID, upload, err := uc.deps.fileClient.InitUpload(ctx, mimeType, size, userID)
	if err != nil {
		return uuid.Nil, ports.UploadInfo{}, ErrAvatarUpload
	}
	return fileID, upload, nil
}

func (uc *ConfirmAvatarUpload) Execute(ctx context.Context, userID uuid.UUID, fileID uuid.UUID, sizeBytes int64) (*model.Profile, error) {
	if userID == uuid.Nil || fileID == uuid.Nil || sizeBytes <= 0 {
		return nil, ErrInvalidInput
	}
	if uc.deps.fileClient == nil {
		return nil, ErrAvatarUpload
	}

	if err := uc.deps.fileClient.ConfirmUpload(ctx, fileID, sizeBytes); err != nil {
		return nil, ErrAvatarUpload
	}

	profile, err := uc.deps.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var oldAvatarID *uuid.UUID
	if profile.AvatarFileID != nil {
		oldID := *profile.AvatarFileID
		oldAvatarID = &oldID
	}

	if err := uc.deps.fileClient.LinkAvatar(ctx, fileID, userID); err != nil {
		return nil, ErrAvatarUpload
	}

	profile.AvatarFileID = &fileID
	if err := uc.deps.profiles.Update(ctx, profile); err != nil {
		_ = uc.deps.fileClient.UnlinkAvatar(ctx, fileID, userID)
		return nil, err
	}

	if oldAvatarID != nil && *oldAvatarID != fileID {
		_ = uc.deps.fileClient.UnlinkAvatar(ctx, *oldAvatarID, userID)
		_ = uc.deps.fileClient.DeleteFileIfUnlinked(ctx, *oldAvatarID)
	}

	return profile, nil
}

func (uc *DeleteAvatar) Execute(ctx context.Context, userID uuid.UUID) (*model.Profile, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidInput
	}

	profile, err := uc.deps.profiles.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile.AvatarFileID == nil {
		return profile, nil
	}

	oldAvatarID := *profile.AvatarFileID
	profile.AvatarFileID = nil
	if err := uc.deps.profiles.Update(ctx, profile); err != nil {
		return nil, err
	}

	if uc.deps.fileClient != nil {
		_ = uc.deps.fileClient.UnlinkAvatar(ctx, oldAvatarID, userID)
		_ = uc.deps.fileClient.DeleteFileIfUnlinked(ctx, oldAvatarID)
	}

	return profile, nil
}

func (uc *GetAvatarDownloadURL) Execute(ctx context.Context, fileID uuid.UUID) (string, error) {
	if fileID == uuid.Nil {
		return "", ErrInvalidInput
	}
	if uc.deps.fileClient == nil {
		return "", ErrAvatarUpload
	}
	url, _, err := uc.deps.fileClient.GetDownloadURL(ctx, fileID)
	if err != nil {
		return "", ErrAvatarUpload
	}
	if strings.TrimSpace(url) == "" {
		return "", ErrAvatarUpload
	}
	return url, nil
}
