package ports

import (
	"context"
	"health/internal/domain/model"

	"github.com/google/uuid"
)

type ListPetDocumentsQuery struct {
	PetID      uuid.UUID
	Cursor     *TimeCursor
	Limit      int
	Q          string
	EntityType *string
	FileType   *string
	ExcludeLog bool
}

type ListPetDocumentsResult struct {
	Items      []model.PetDocument
	NextCursor *TimeCursor
}

type RenamePetDocumentInput struct {
	ID        uuid.UUID
	PetID     uuid.UUID
	FileName  string
	UpdatedBy uuid.UUID
}

type DocumentsRepository interface {
	ListPetDocuments(ctx context.Context, query ListPetDocumentsQuery) (ListPetDocumentsResult, error)
	GetPetDocument(ctx context.Context, petID, documentID uuid.UUID) (*model.PetDocument, error)
	RenamePetDocument(ctx context.Context, input RenamePetDocumentInput) (*model.PetDocument, error)
}
