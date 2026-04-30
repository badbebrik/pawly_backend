package handlers

import (
	"net/http"
	"strings"

	pushuc "push/internal/application/usecase"
	"push/internal/transport/http/dto"
)

func (h *Handlers) HandleDevices(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v1/push/devices" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, ok := requireUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	var req dto.UpsertDeviceTokenRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json")
		return
	}

	item, err := h.useCases.RegisterDeviceToken(r.Context(), pushuc.RegisterDeviceTokenParams{
		UserID:    userID,
		DeviceID:  req.DeviceID,
		Platform:  strings.TrimSpace(req.Platform),
		PushToken: req.PushToken,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, dto.DeviceTokenEnvelopeResponse{
		Item: deviceTokenToResponse(item),
	})
}

func (h *Handlers) HandleDeviceByID(w http.ResponseWriter, r *http.Request) {
	deviceID, ok := parseDeviceID(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, ok := requireUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	if err := h.useCases.DeleteDeviceToken(r.Context(), pushuc.DeleteDeviceTokenParams{
		UserID:   userID,
		DeviceID: deviceID,
	}); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
