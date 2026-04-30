package usecase

import (
	"context"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"strings"

	"github.com/google/uuid"
)

type Documents struct {
	repo ports.DocumentsRepository
	acl  ports.HealthAccessChecker
	file ports.HealthFileClient
}

func NewDocuments(repo ports.DocumentsRepository, acl ports.HealthAccessChecker, file ports.HealthFileClient) *Documents {
	return &Documents{repo: repo, acl: acl, file: file}
}

type ListPetDocumentsParams struct {
	UserID     uuid.UUID
	PetID      uuid.UUID
	Cursor     *ports.TimeCursor
	Limit      int
	Q          string
	EntityType *string
	FileType   *string
}

type ListPetDocumentsResult struct {
	Items      []model.PetDocument
	NextCursor *ports.TimeCursor
}

type RenamePetDocumentParams struct {
	UserID     uuid.UUID
	PetID      uuid.UUID
	DocumentID uuid.UUID
	FileName   string
}

func (u *Documents) ListPetDocuments(ctx context.Context, p ListPetDocumentsParams) (ListPetDocumentsResult, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil {
		return ListPetDocumentsResult{}, ErrInvalidInput
	}
	if _, err := requireHealthRead(ctx, u.acl, p.PetID, p.UserID); err != nil {
		return ListPetDocumentsResult{}, err
	}
	logAllowed, err := u.acl.Check(ctx, p.PetID, p.UserID, ActionLogRead)
	if err != nil {
		return ListPetDocumentsResult{}, err
	}
	entityType := normalizeHealthDocumentEntityType(p.EntityType)
	if p.EntityType != nil && entityType == nil {
		return ListPetDocumentsResult{}, ErrInvalidInput
	}
	fileType := normalizeDocumentFileType(p.FileType)
	if p.FileType != nil && fileType == nil {
		return ListPetDocumentsResult{}, ErrInvalidInput
	}
	if !logAllowed && entityType != nil && *entityType == "LOG" {
		return ListPetDocumentsResult{Items: []model.PetDocument{}}, nil
	}
	out, err := u.repo.ListPetDocuments(ctx, ports.ListPetDocumentsQuery{
		PetID:      p.PetID,
		Cursor:     p.Cursor,
		Limit:      p.Limit,
		Q:          strings.TrimSpace(p.Q),
		EntityType: entityType,
		FileType:   fileType,
		ExcludeLog: !logAllowed,
	})
	if err != nil {
		return ListPetDocumentsResult{}, mapRepoErr(err)
	}
	enrichPetDocumentURLs(ctx, u.file, out.Items)
	return ListPetDocumentsResult{Items: out.Items, NextCursor: out.NextCursor}, nil
}

func (u *Documents) RenamePetDocument(ctx context.Context, p RenamePetDocumentParams) (*model.PetDocument, error) {
	if p.UserID == uuid.Nil || p.PetID == uuid.Nil || p.DocumentID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	name := strings.TrimSpace(p.FileName)
	if name == "" || len([]rune(name)) > 255 {
		return nil, ErrInvalidInput
	}
	current, err := u.repo.GetPetDocument(ctx, p.PetID, p.DocumentID)
	if err != nil {
		return nil, mapRepoErr(err)
	}
	if current.EntityType == "LOG" {
		allowed, err := u.acl.Check(ctx, p.PetID, p.UserID, ActionLogWrite)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, ErrForbidden
		}
	} else if err := requireHealthWrite(ctx, u.acl, p.PetID, p.UserID); err != nil {
		return nil, err
	}
	item, err := u.repo.RenamePetDocument(ctx, ports.RenamePetDocumentInput{
		ID:        p.DocumentID,
		PetID:     p.PetID,
		FileName:  name,
		UpdatedBy: p.UserID,
	})
	if err != nil {
		return nil, mapRepoErr(err)
	}
	items := []model.PetDocument{*item}
	enrichPetDocumentURLs(ctx, u.file, items)
	return &items[0], nil
}

func enrichPetDocumentURLs(ctx context.Context, file ports.HealthFileClient, items []model.PetDocument) {
	if len(items) == 0 {
		return
	}
	fileIDs := make([]uuid.UUID, 0, len(items))
	for i := range items {
		fileIDs = append(fileIDs, items[i].FileID)
	}
	urls, err := file.BatchGetDownloadURLs(ctx, fileIDs)
	if err != nil {
		return
	}
	for i := range items {
		if url, ok := urls[items[i].FileID]; ok && strings.TrimSpace(url) != "" {
			urlCopy := url
			items[i].DownloadURL = &urlCopy
			if items[i].FileType == "image" {
				items[i].PreviewURL = &urlCopy
			}
		}
	}
}

func normalizeHealthDocumentEntityType(in *string) *string {
	if in == nil {
		return nil
	}
	v := strings.ToUpper(strings.TrimSpace(*in))
	switch v {
	case "LOG", "VET_VISIT", "VACCINATION", "PROCEDURE", "MEDICAL_RECORD":
		return &v
	default:
		return nil
	}
}

func normalizeDocumentFileType(in *string) *string {
	if in == nil {
		return nil
	}
	v := strings.ToLower(strings.TrimSpace(*in))
	switch v {
	case "image", "pdf", "other":
		return &v
	default:
		return nil
	}
}
