package usecase

import (
	"pet/internal/domain/model"
	"strings"

	"github.com/google/uuid"
)

func isAllowedPetPhotoMimeType(raw string) bool {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "image/jpeg", "image/png":
		return true
	default:
		return false
	}
}

func normalizeExclusiveTextChoice(id *uuid.UUID, raw *string) (*string, error) {
	if id != nil && *id == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if raw == nil {
		return nil, nil
	}

	value := strings.TrimSpace(*raw)
	if value == "" {
		raw = nil
	} else {
		raw = &value
	}

	if id != nil && raw != nil {
		return nil, ErrInvalidInput
	}

	return raw, nil
}

func normalizeSpeciesChoice(speciesID *uuid.UUID, raw *string) (*string, error) {
	if speciesID != nil && *speciesID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	if raw == nil {
		if speciesID == nil {
			return nil, ErrInvalidInput
		}
		return nil, nil
	}

	value := strings.TrimSpace(*raw)
	if value == "" {
		if speciesID == nil {
			return nil, ErrInvalidInput
		}
		return nil, nil
	}
	if len([]rune(value)) > 64 {
		return nil, ErrInvalidInput
	}
	if speciesID != nil {
		return nil, ErrInvalidInput
	}
	return &value, nil
}

func normalizeColors(colors []model.Color) ([]model.Color, error) {
	if len(colors) == 0 {
		return []model.Color{}, nil
	}

	out := make([]model.Color, 0, len(colors))
	for i := range colors {
		item := colors[i]
		if item.PresetID != nil && *item.PresetID == uuid.Nil {
			return nil, ErrInvalidInput
		}

		if item.CustomName != nil {
			value := strings.TrimSpace(*item.CustomName)
			if value == "" {
				item.CustomName = nil
			} else {
				item.CustomName = &value
			}
		}
		if item.CustomHex != nil {
			value := strings.TrimSpace(*item.CustomHex)
			if value == "" {
				item.CustomHex = nil
			} else {
				item.CustomHex = &value
			}
		}

		switch {
		case item.PresetID != nil && item.CustomName == nil && item.CustomHex == nil:
		case item.PresetID == nil && item.CustomName != nil && item.CustomHex != nil:
		default:
			return nil, ErrInvalidInput
		}

		out = append(out, model.Color{
			PresetID:   item.PresetID,
			CustomName: item.CustomName,
			CustomHex:  item.CustomHex,
			SortOrder:  item.SortOrder,
		})
	}

	return out, nil
}
