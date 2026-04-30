package handlers

import (
	healthuc "health/internal/application/usecase"
	"health/internal/transport/http/dto"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handlers) GetScheduledItems(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	cursor, err := decodeScheduledTimeCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid cursor")
		return
	}
	dateFrom, err := parseOptionalDateTime(r.URL.Query().Get("date_from"), false)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_from")
		return
	}
	dateTo, err := parseOptionalDateTime(r.URL.Query().Get("date_to"), true)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_to")
		return
	}
	resp, err := h.scheduled.ListScheduledItems(r.Context(), healthuc.ListScheduledItemsParams{
		UserID:      userID,
		PetID:       petID,
		Cursor:      cursor,
		Limit:       parseIntOrDefault(r.URL.Query().Get("limit"), 20),
		SourceType:  r.URL.Query().Get("source_type"),
		DateFrom:    dateFrom,
		DateTo:      dateTo,
		IncludePast: parseBoolOrDefault(r.URL.Query().Get("include_past"), false),
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	items := make([]dto.ScheduledItemListItemResponse, 0, len(resp.Items))
	for i := range resp.Items {
		items = append(items, scheduledItemListItemAppToDTO(resp.Items[i]))
	}
	writeJSON(w, http.StatusOK, dto.ScheduledItemsListResponse{Items: items, NextCursor: encodeScheduledTimeCursor(resp.NextCursor)})
}

func (h *Handlers) GetScheduledItem(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	itemID, err := uuid.Parse(chi.URLParam(r, "item_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid item_id")
		return
	}
	item, err := h.scheduled.GetScheduledItem(r.Context(), userID, petID, itemID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, scheduledItemAppToDTO(item))
}

func (h *Handlers) CreateScheduledItem(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	var req dto.CreateOrUpdateScheduledItemRequest
	if !decodeBody(w, r, &req) {
		return
	}
	params, ok := parseScheduledItemRequest(w, userID, petID, req)
	if !ok {
		return
	}
	item, err := h.scheduled.CreateScheduledItem(r.Context(), params)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, scheduledItemAppToDTO(item))
}

func (h *Handlers) UpdateScheduledItem(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	itemID, err := uuid.Parse(chi.URLParam(r, "item_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid item_id")
		return
	}
	var req dto.CreateOrUpdateScheduledItemRequest
	if !decodeBody(w, r, &req) {
		return
	}
	params, ok := parseScheduledItemRequest(w, userID, petID, req)
	if !ok {
		return
	}
	item, err := h.scheduled.UpdateScheduledItem(r.Context(), healthuc.UpdateScheduledItemParams{
		UserID:              params.UserID,
		PetID:               params.PetID,
		ItemID:              itemID,
		RowVersion:          req.RowVersion,
		Title:               params.Title,
		Note:                params.Note,
		StartsAt:            params.StartsAt,
		PushEnabled:         params.PushEnabled,
		RemindOffsetMinutes: params.RemindOffsetMinutes,
		RecurrenceRule:      params.RecurrenceRule,
		RecurrenceInterval:  params.RecurrenceInterval,
		RecurrenceUntil:     params.RecurrenceUntil,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, scheduledItemAppToDTO(item))
}

func (h *Handlers) DeleteScheduledItem(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	itemID, err := uuid.Parse(chi.URLParam(r, "item_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid item_id")
		return
	}
	var req dto.DeleteRowVersionRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if err := h.scheduled.DeleteScheduledItem(r.Context(), healthuc.DeleteScheduledItemParams{
		UserID:     userID,
		PetID:      petID,
		ItemID:     itemID,
		RowVersion: req.RowVersion,
	}); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) UpdateScheduledItemReminderSettings(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	itemID, err := uuid.Parse(chi.URLParam(r, "item_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid item_id")
		return
	}
	var req dto.UpdateScheduledItemReminderSettingsRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "row_version is required")
		return
	}
	item, err := h.scheduled.UpdateScheduledItemReminderSettings(r.Context(), healthuc.UpdateScheduledItemReminderSettingsParams{
		UserID:              userID,
		PetID:               petID,
		ItemID:              itemID,
		RowVersion:          req.RowVersion,
		PushEnabled:         req.PushEnabled,
		RemindOffsetMinutes: req.RemindOffsetMinutes,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, scheduledItemAppToDTO(item))
}

func (h *Handlers) GetScheduledItemOccurrences(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	cursor, err := decodeScheduledTimeCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid cursor")
		return
	}
	dateFrom, err := parseOptionalDateTime(r.URL.Query().Get("date_from"), false)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_from")
		return
	}
	dateTo, err := parseOptionalDateTime(r.URL.Query().Get("date_to"), true)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_to")
		return
	}
	resp, err := h.scheduled.ListScheduledItemOccurrences(r.Context(), healthuc.ListScheduledItemOccurrencesParams{
		UserID:     userID,
		PetID:      petID,
		Cursor:     cursor,
		Limit:      parseIntOrDefault(r.URL.Query().Get("limit"), 20),
		SourceType: r.URL.Query().Get("source_type"),
		DateFrom:   dateFrom,
		DateTo:     dateTo,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	items := make([]dto.ScheduledOccurrenceResponse, 0, len(resp.Items))
	for i := range resp.Items {
		items = append(items, scheduledOccurrenceAppToDTO(resp.Items[i]))
	}
	writeJSON(w, http.StatusOK, dto.ScheduledOccurrencesListResponse{Items: items, NextCursor: encodeScheduledTimeCursor(resp.NextCursor)})
}

func (h *Handlers) GetScheduledItemOccurrence(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	occurrenceID, err := uuid.Parse(chi.URLParam(r, "occurrence_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid occurrence_id")
		return
	}
	item, err := h.scheduled.GetScheduledItemOccurrence(r.Context(), userID, petID, occurrenceID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, scheduledOccurrenceAppToDTO(*item))
}
