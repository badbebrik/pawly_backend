package handlers

import (
	"net/http"
	"pet/internal/application/usecase"
	"pet/internal/domain/model"
	"pet/internal/transport/http/dto"
	"strconv"

	"github.com/google/uuid"
)

type Handlers struct {
	useCases *usecase.Set
}

func New(useCases *usecase.Set) *Handlers {
	return &Handlers{useCases: useCases}
}

func parseOptionalUUID(raw *string) (*uuid.UUID, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func parseColorRequests(items []dto.ColorRequest) ([]model.Color, error) {
	if len(items) == 0 {
		return []model.Color{}, nil
	}

	out := make([]model.Color, 0, len(items))
	for i := range items {
		var presetID *uuid.UUID
		if items[i].PresetID != nil && *items[i].PresetID != "" {
			id, err := uuid.Parse(*items[i].PresetID)
			if err != nil {
				return nil, err
			}
			presetID = &id
		}

		out = append(out, model.Color{
			PresetID:   presetID,
			CustomName: items[i].CustomName,
			CustomHex:  items[i].CustomHex,
			SortOrder:  items[i].SortOrder,
		})
	}

	return out, nil
}

func parseIntOrDefault(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func (h *Handlers) getPetPhotoDownloadURL(r *http.Request, pet *model.Pet) *string {
	return h.useCases.ResolveProfilePhotoDownloadURL(r.Context(), pet.ProfilePhotoFileID)
}
