package usecase

import (
	"context"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"strings"

	"github.com/google/uuid"
)

type AttachmentParam struct {
	FileID   uuid.UUID
	FileName *string
}

func prepareHealthAttachments(ctx context.Context, fileClient ports.HealthFileClient, in []AttachmentParam) ([]ports.AttachmentInput, error) {
	ids := make([]uuid.UUID, 0, len(in))
	namesByID := make(map[uuid.UUID]*string, len(in))
	for i := range in {
		if in[i].FileID == uuid.Nil {
			return nil, ErrInvalidInput
		}
		if in[i].FileName != nil && len([]rune(strings.TrimSpace(*in[i].FileName))) > 255 {
			return nil, ErrInvalidInput
		}
		if _, exists := namesByID[in[i].FileID]; exists {
			return nil, ErrInvalidInput
		}
		ids = append(ids, in[i].FileID)
		namesByID[in[i].FileID] = normalizeAttachmentFileName(in[i].FileName)
	}
	if len(ids) == 0 {
		return []ports.AttachmentInput{}, nil
	}
	files, err := fileClient.GetFiles(ctx, ids)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if len(files) != len(ids) {
		return nil, ErrInvalidInput
	}
	attachments := make([]ports.AttachmentInput, 0, len(ids))
	for i := range ids {
		file, ok := files[ids[i]]
		if !ok {
			return nil, ErrInvalidInput
		}
		fileName := namesByID[ids[i]]
		if fileName == nil {
			fileName = file.FileName
		}
		attachments = append(attachments, ports.AttachmentInput{
			FileID:   ids[i],
			FileName: fileName,
			FileType: detectAttachmentFileType(file.MimeType),
		})
	}
	return attachments, nil
}

func normalizeAttachmentFileName(name *string) *string {
	if name == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*name)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func syncHealthAttachments(ctx context.Context, fileClient ports.HealthFileClient, petID uuid.UUID, entityType string, entityID uuid.UUID, sync ports.AttachmentSync) error {
	if len(sync.Add) > 0 {
		if err := fileClient.LinkAttachments(ctx, petID, entityType, entityID, sync.Add); err != nil {
			return mapRepoErr(err)
		}
	}
	if len(sync.Remove) > 0 {
		if err := fileClient.UnlinkAttachments(ctx, entityType, entityID, sync.Remove); err != nil {
			return mapRepoErr(err)
		}
		if err := fileClient.DeleteFilesIfUnlinked(ctx, sync.Remove); err != nil {
			return mapRepoErr(err)
		}
	}
	return nil
}

func enrichHealthAttachmentURLs(ctx context.Context, fileClient ports.HealthFileClient, attachments []model.HealthAttachment) {
	if len(attachments) == 0 {
		return
	}
	fileIDs := make([]uuid.UUID, 0, len(attachments))
	for i := range attachments {
		fileIDs = append(fileIDs, attachments[i].FileID)
	}
	urls, err := fileClient.BatchGetDownloadURLs(ctx, fileIDs)
	if err != nil {
		return
	}
	for i := range attachments {
		if url, ok := urls[attachments[i].FileID]; ok && strings.TrimSpace(url) != "" {
			urlCopy := url
			attachments[i].DownloadURL = &urlCopy
			if attachments[i].FileType == "image" {
				attachments[i].PreviewURL = &urlCopy
			}
		}
	}
}
