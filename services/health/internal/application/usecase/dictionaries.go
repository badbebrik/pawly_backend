package usecase

import (
	"context"
	"health/internal/application/ports"
	"health/internal/domain/model"
	"strings"

	"github.com/google/uuid"
)

type HealthDictionaryItemRefParam struct {
	ID   *uuid.UUID
	Name *string
}

func resolveDictionaryItem(ctx context.Context, repo ports.HealthDictionaryRepository, petID, userID uuid.UUID, kind string, itemID *uuid.UUID, customName *string) (*model.HealthDictionaryItem, error) {
	if itemID != nil {
		if *itemID == uuid.Nil {
			return nil, ErrInvalidInput
		}
		item, err := repo.GetHealthDictionaryItem(ctx, petID, *itemID, kind)
		if err != nil {
			return nil, mapRepoErr(err)
		}
		if item.IsArchived {
			return nil, ErrInvalidInput
		}
		return item, nil
	}

	if name := strings.TrimSpace(derefString(customName)); name != "" {
		item, err := repo.GetOrCreateCustomHealthDictionaryItem(ctx, ports.GetOrCreateCustomHealthDictionaryItemInput{
			PetID:     petID,
			Kind:      kind,
			Name:      name,
			CreatedBy: userID,
			UpdatedBy: userID,
		})
		if err != nil {
			return nil, mapRepoErr(err)
		}
		return item, nil
	}
	return nil, ErrInvalidInput
}

func resolveDictionaryItemRefs(ctx context.Context, repo ports.HealthDictionaryRepository, petID, userID uuid.UUID, kind string, refs []HealthDictionaryItemRefParam) ([]uuid.UUID, error) {
	if refs == nil {
		return nil, nil
	}
	if len(refs) == 0 {
		return []uuid.UUID{}, nil
	}
	out := make([]uuid.UUID, 0, len(refs))
	seen := make(map[uuid.UUID]struct{}, len(refs))
	for i := range refs {
		ref := refs[i]
		var item *model.HealthDictionaryItem
		if ref.ID != nil {
			if *ref.ID == uuid.Nil {
				return nil, ErrInvalidInput
			}
			var err error
			item, err = repo.GetHealthDictionaryItem(ctx, petID, *ref.ID, kind)
			if err != nil {
				return nil, mapRepoErr(err)
			}
		} else {
			name := strings.TrimSpace(derefString(ref.Name))
			if name == "" {
				return nil, ErrInvalidInput
			}
			var err error
			item, err = repo.GetOrCreateCustomHealthDictionaryItem(ctx, ports.GetOrCreateCustomHealthDictionaryItemInput{
				PetID:     petID,
				Kind:      kind,
				Name:      name,
				CreatedBy: userID,
				UpdatedBy: userID,
			})
			if err != nil {
				return nil, mapRepoErr(err)
			}
		}
		if item == nil || item.IsArchived {
			return nil, ErrInvalidInput
		}
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		out = append(out, item.ID)
	}
	return out, nil
}

func dictionaryItemsByKind(items []model.HealthDictionaryItem, kind string) []model.HealthDictionaryItem {
	out := make([]model.HealthDictionaryItem, 0)
	for i := range items {
		if items[i].Kind == kind {
			out = append(out, items[i])
		}
	}
	return out
}

func dictionaryItemIDPtr(item *model.HealthDictionaryItem) *uuid.UUID {
	if item == nil {
		return nil
	}
	id := item.ID
	return &id
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
