package handlers

import (
	"net/http"
	"strings"

	pushuc "push/internal/application/usecase"
	"push/internal/transport/http/dto"
)

func (h *Handlers) HandlePetPushSettings(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "/push-settings") {
		http.NotFound(w, r)
		return
	}

	petID, ok := parsePetID(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	userID, ok := requireUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		item, err := h.useCases.GetPetPushSettings(r.Context(), pushuc.GetPetPushSettingsParams{
			UserID: userID,
			PetID:  petID,
		})
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dto.PetPushSettingsEnvelopeResponse{
			Item: petPushSettingsToResponse(item),
		})
	case http.MethodPatch:
		var req dto.PetPushSettingsRequest
		if err := decodeRequest(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
			return
		}
		item, err := h.useCases.UpdatePetPushSettings(r.Context(), pushuc.UpdatePetPushSettingsParams{
			UserID:                userID,
			PetID:                 petID,
			ScheduledItemsEnabled: req.ScheduledItemsEnabled,
		})
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dto.PetPushSettingsEnvelopeResponse{
			Item: petPushSettingsToResponse(item),
		})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
