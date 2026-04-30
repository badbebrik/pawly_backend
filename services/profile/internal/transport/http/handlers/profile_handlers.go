package handlers

import (
	"net/http"

	"github.com/google/uuid"

	profileuc "profile/internal/application/usecase"
	"profile/internal/transport/http/dto"
	"profile/internal/transport/http/middleware"
)

func (h *Handlers) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	profile, err := h.useCases.GetProfile.Execute(r.Context(), userID)
	if err != nil {
		writeProfileQueryError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, h.profileToResponse(r.Context(), profile))
}

func (h *Handlers) PatchMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var req dto.UpdateProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	profile, err := h.useCases.UpdateProfileInfo.Execute(r.Context(), userID, profileuc.UpdateProfileInfoParams{
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		writeUpdateProfileInfoError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, h.profileToResponse(r.Context(), profile))
}

func (h *Handlers) PatchPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var req dto.UpdatePreferencesRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	profile, err := h.useCases.UpdatePreferences.Execute(r.Context(), userID, profileuc.UpdatePreferencesParams{
		Locale:   req.Locale,
		Timezone: req.Timezone,
	})
	if err != nil {
		writeUpdatePreferencesError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, h.profileToResponse(r.Context(), profile))
}

func (h *Handlers) BatchProfilesBrief(w http.ResponseWriter, r *http.Request) {
	var req dto.BatchProfilesBriefRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}
	if len(req.UserIDs) == 0 {
		writeJSON(w, http.StatusOK, dto.BatchProfilesBriefResponse{
			Items:           []dto.ProfileBriefResponse{},
			NotFoundUserIDs: []uuid.UUID{},
		})
		return
	}

	userIDs := make([]uuid.UUID, 0, len(req.UserIDs))
	for i := range req.UserIDs {
		userID, err := uuid.Parse(req.UserIDs[i])
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_user_ids", "invalid user ids")
			return
		}
		userIDs = append(userIDs, userID)
	}

	items, notFound, err := h.useCases.BatchProfilesBrief.Execute(r.Context(), userIDs)
	if err != nil {
		writeBatchProfilesBriefError(w, err)
		return
	}

	out := make([]dto.ProfileBriefResponse, 0, len(items))
	for i := range items {
		out = append(out, profileBriefToResponse(items[i]))
	}

	writeJSON(w, http.StatusOK, dto.BatchProfilesBriefResponse{
		Items:           out,
		NotFoundUserIDs: notFound,
	})
}
