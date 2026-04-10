package app

import (
	"encoding/json"
	"net/http"
	"strings"

	"push/internal/service"
)

func (a *App) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/v1/push/devices", a.handleDevices)
	mux.HandleFunc("/v1/push/devices/", a.handleDeviceByID)
	mux.HandleFunc("/v1/pets/", a.handlePetPushSettings)

	return mux
}

type upsertDeviceTokenRequest struct {
	DeviceID  string `json:"device_id"`
	Platform  string `json:"platform"`
	PushToken string `json:"push_token"`
}

type petPushSettingsRequest struct {
	ScheduledItemsEnabled bool `json:"scheduled_items_enabled"`
}

func (a *App) handleDevices(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v1/push/devices" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, ok := a.getUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user id")
		return
	}

	var req upsertDeviceTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}

	item, err := a.svc.RegisterDeviceToken(r.Context(), service.RegisterDeviceTokenParams{
		UserID:    userID,
		DeviceID:  req.DeviceID,
		Platform:  strings.TrimSpace(req.Platform),
		PushToken: req.PushToken,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":         item.ID,
		"user_id":    item.UserID,
		"device_id":  item.DeviceID,
		"platform":   item.Platform,
		"push_token": item.PushToken,
		"created_at": item.CreatedAt,
		"updated_at": item.UpdatedAt,
	})
}

func (a *App) handleDeviceByID(w http.ResponseWriter, r *http.Request) {
	deviceID, ok := parseDeviceID(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, ok := a.getUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user id")
		return
	}

	if err := a.svc.DeleteDeviceToken(r.Context(), service.DeleteDeviceTokenParams{
		UserID:   userID,
		DeviceID: deviceID,
	}); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handlePetPushSettings(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "/push-settings") {
		http.NotFound(w, r)
		return
	}

	petID, ok := parsePetID(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	userID, ok := a.getUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		item, err := a.svc.GetPetPushSettings(r.Context(), service.GetPetPushSettingsParams{
			UserID: userID,
			PetID:  petID,
		})
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"pet_id":                   item.PetID,
			"scheduled_items_enabled":  item.ScheduledItemsEnabled,
			"created_at":               item.CreatedAt,
			"updated_at":               item.UpdatedAt,
		})
	case http.MethodPatch:
		var req petPushSettingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
			return
		}
		item, err := a.svc.UpdatePetPushSettings(r.Context(), service.UpdatePetPushSettingsParams{
			UserID:                userID,
			PetID:                 petID,
			ScheduledItemsEnabled: req.ScheduledItemsEnabled,
		})
		if err != nil {
			writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"pet_id":                   item.PetID,
			"scheduled_items_enabled":  item.ScheduledItemsEnabled,
			"created_at":               item.CreatedAt,
			"updated_at":               item.UpdatedAt,
		})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
