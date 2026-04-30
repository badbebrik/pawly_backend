package ports

import (
	"context"
	"health/internal/domain/model"

	"github.com/google/uuid"
)

const (
	HealthDictionaryKindProcedureType     = "PROCEDURE_TYPE"
	HealthDictionaryKindMedicalRecordType = "MEDICAL_RECORD_TYPE"
	HealthDictionaryKindVaccinationTarget = "VACCINATION_TARGET"
)

type ListHealthDictionaryItemsInput struct {
	PetID uuid.UUID
	Kinds []string
}

type GetOrCreateCustomHealthDictionaryItemInput struct {
	PetID     uuid.UUID
	Kind      string
	Name      string
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
}

type HealthDictionaryRepository interface {
	ListHealthDictionaryItems(ctx context.Context, in ListHealthDictionaryItemsInput) ([]model.HealthDictionaryItem, error)
	GetHealthDictionaryItem(ctx context.Context, petID, itemID uuid.UUID, kind string) (*model.HealthDictionaryItem, error)
	GetOrCreateCustomHealthDictionaryItem(ctx context.Context, in GetOrCreateCustomHealthDictionaryItemInput) (*model.HealthDictionaryItem, error)
}
