package handlers

import (
	healthuc "health/internal/application/usecase"
	"health/internal/transport/http/dto"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handlers) GetVetVisits(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	cursor, err := decodeTimeCursor(r.URL.Query().Get("cursor"))
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
	resp, err := h.vetVisits.ListVetVisits(r.Context(), healthuc.ListVetVisitsParams{
		UserID:   userID,
		PetID:    petID,
		Cursor:   cursor,
		Limit:    parseIntOrDefault(r.URL.Query().Get("limit"), 20),
		Q:        r.URL.Query().Get("q"),
		Status:   r.URL.Query().Get("status"),
		Bucket:   r.URL.Query().Get("bucket"),
		DateFrom: dateFrom,
		DateTo:   dateTo,
		Sort:     r.URL.Query().Get("sort"),
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	items := make([]dto.VetVisitListItemResponse, 0, len(resp.Items))
	for i := range resp.Items {
		items = append(items, vetVisitAppListItemToDTO(resp.Items[i]))
	}
	writeJSON(w, http.StatusOK, dto.VetVisitsListResponse{Items: items, NextCursor: encodeScheduledTimeCursor(resp.NextCursor)})
}

func (h *Handlers) GetVetVisit(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	visitID, err := uuid.Parse(chi.URLParam(r, "visit_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid visit_id")
		return
	}
	item, err := h.vetVisits.GetVetVisit(r.Context(), userID, petID, visitID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, vetVisitAppToDTO(item))
}

func (h *Handlers) CreateVetVisit(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	var req dto.CreateOrUpdateVetVisitRequest
	if !decodeBody(w, r, &req) {
		return
	}
	scheduledAt, err := parseOptionalRFC3339(req.ScheduledAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid scheduled_at")
		return
	}
	completedAt, err := parseOptionalRFC3339(req.CompletedAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid completed_at")
		return
	}
	attachments, err := parseAttachmentParams(req.Attachments)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid attachments")
		return
	}
	item, err := h.vetVisits.CreateVetVisit(r.Context(), healthuc.CreateVetVisitParams{UserID: userID, PetID: petID, Status: req.Status, VisitType: req.VisitType, Title: req.Title, ScheduledAt: scheduledAt, Reminder: medicalEntityReminderToUsecaseParams(req.Reminder), CompletedAt: completedAt, ReasonText: req.ReasonText, ResultText: req.ResultText, ClinicName: req.ClinicName, VetName: req.VetName, Attachments: attachments})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, vetVisitAppToDTO(item))
}

func (h *Handlers) UpdateVetVisit(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	visitID, err := uuid.Parse(chi.URLParam(r, "visit_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid visit_id")
		return
	}
	var req dto.CreateOrUpdateVetVisitRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "row_version is required")
		return
	}
	scheduledAt, err := parseOptionalRFC3339(req.ScheduledAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid scheduled_at")
		return
	}
	completedAt, err := parseOptionalRFC3339(req.CompletedAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid completed_at")
		return
	}
	attachments, err := parseAttachmentParams(req.Attachments)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid attachments")
		return
	}
	item, err := h.vetVisits.UpdateVetVisit(r.Context(), healthuc.UpdateVetVisitParams{UserID: userID, PetID: petID, VisitID: visitID, RowVersion: req.RowVersion, Status: req.Status, VisitType: req.VisitType, Title: req.Title, ScheduledAt: scheduledAt, Reminder: medicalEntityReminderToUsecaseParams(req.Reminder), CompletedAt: completedAt, ReasonText: req.ReasonText, ResultText: req.ResultText, ClinicName: req.ClinicName, VetName: req.VetName, Attachments: attachments})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, vetVisitAppToDTO(item))
}

func (h *Handlers) DeleteVetVisit(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	visitID, err := uuid.Parse(chi.URLParam(r, "visit_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid visit_id")
		return
	}
	var req dto.DeleteRowVersionRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "row_version is required")
		return
	}
	if err := h.vetVisits.DeleteVetVisit(r.Context(), healthuc.DeleteVetVisitParams{UserID: userID, PetID: petID, VisitID: visitID, RowVersion: req.RowVersion}); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) LinkVetVisitLog(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	visitID, err := uuid.Parse(chi.URLParam(r, "visit_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid visit_id")
		return
	}
	var req dto.LinkVetVisitLogRequest
	if !decodeBody(w, r, &req) {
		return
	}
	logID, err := uuid.Parse(req.LogID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid log_id")
		return
	}
	item, err := h.vetVisits.LinkVetVisitLog(r.Context(), userID, petID, visitID, logID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, relatedLogAppToDTO(*item))
}

func (h *Handlers) UnlinkVetVisitLog(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	visitID, err := uuid.Parse(chi.URLParam(r, "visit_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid visit_id")
		return
	}
	logID, err := uuid.Parse(chi.URLParam(r, "log_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid log_id")
		return
	}
	if err := h.vetVisits.UnlinkVetVisitLog(r.Context(), userID, petID, visitID, logID); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
