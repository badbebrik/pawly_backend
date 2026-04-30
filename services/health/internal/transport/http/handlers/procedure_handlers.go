package handlers

import (
	healthuc "health/internal/application/usecase"
	domainmodel "health/internal/domain/model"
	"health/internal/transport/http/dto"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handlers) GetProcedures(w http.ResponseWriter, r *http.Request) {
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
	var procedureTypeID *uuid.UUID
	if rawProcedureTypeID := r.URL.Query().Get("procedure_type_id"); rawProcedureTypeID != "" {
		parsed, err := uuid.Parse(rawProcedureTypeID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid procedure_type_id")
			return
		}
		procedureTypeID = &parsed
	}
	resp, err := h.procedures.ListProcedures(r.Context(), healthuc.ListProceduresParams{
		UserID:          userID,
		PetID:           petID,
		Cursor:          cursor,
		Limit:           parseIntOrDefault(r.URL.Query().Get("limit"), 20),
		Q:               r.URL.Query().Get("q"),
		Status:          r.URL.Query().Get("status"),
		Bucket:          r.URL.Query().Get("bucket"),
		ProcedureTypeID: procedureTypeID,
		DateFrom:        dateFrom,
		DateTo:          dateTo,
		Sort:            r.URL.Query().Get("sort"),
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	items := make([]dto.ProcedureListItemResponse, 0, len(resp.Items))
	for i := range resp.Items {
		items = append(items, procedureAppListItemToDTO(resp.Items[i]))
	}
	writeJSON(w, http.StatusOK, dto.ProceduresListResponse{Items: items, NextCursor: encodeTimeCursor(resp.NextCursor)})
}

func (h *Handlers) GetProcedure(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	procedureID, err := uuid.Parse(chi.URLParam(r, "procedure_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid procedure_id")
		return
	}
	item, err := h.procedures.GetProcedure(r.Context(), userID, petID, procedureID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, procedureAppToDTO(item))
}

func (h *Handlers) CreateProcedure(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	var req dto.CreateOrUpdateProcedureRequest
	if !decodeBody(w, r, &req) {
		return
	}
	item, err := h.createOrUpdateProcedure(r, userID, petID, uuid.Nil, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, procedureAppToDTO(item))
}

func (h *Handlers) UpdateProcedure(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	procedureID, err := uuid.Parse(chi.URLParam(r, "procedure_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid procedure_id")
		return
	}
	var req dto.CreateOrUpdateProcedureRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "row_version is required")
		return
	}
	item, err := h.updateProcedure(r, userID, petID, procedureID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, procedureAppToDTO(item))
}

func (h *Handlers) DeleteProcedure(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	procedureID, err := uuid.Parse(chi.URLParam(r, "procedure_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid procedure_id")
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
	if err := h.procedures.DeleteProcedure(r.Context(), healthuc.DeleteProcedureParams{UserID: userID, PetID: petID, ProcedureID: procedureID, RowVersion: req.RowVersion}); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) createOrUpdateProcedure(r *http.Request, userID, petID, procedureID uuid.UUID, req dto.CreateOrUpdateProcedureRequest) (*domainmodel.Procedure, error) {
	scheduledAt, err := parseOptionalRFC3339(req.ScheduledAt)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	performedAt, err := parseOptionalRFC3339(req.PerformedAt)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	nextDueAt, err := parseOptionalRFC3339(req.NextDueAt)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	catalogMedicationID, err := optionalUUIDFromString(req.CatalogMedicationID)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	procedureTypeID, err := optionalUUIDFromString(req.ProcedureTypeID)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	vetVisitID, err := optionalUUIDFromString(req.VetVisitID)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	attachments, err := parseAttachmentParams(req.Attachments)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	if procedureID == uuid.Nil {
		return h.procedures.CreateProcedure(r.Context(), healthuc.CreateProcedureParams{
			UserID:              userID,
			PetID:               petID,
			Status:              req.Status,
			ProcedureTypeID:     procedureTypeID,
			ProcedureTypeName:   req.ProcedureTypeName,
			Title:               req.Title,
			Description:         req.Description,
			CatalogMedicationID: catalogMedicationID,
			ProductName:         req.ProductName,
			ScheduledAt:         scheduledAt,
			Reminder:            medicalEntityReminderToUsecaseParams(req.Reminder),
			PerformedAt:         performedAt,
			NextDueAt:           nextDueAt,
			VetVisitID:          vetVisitID,
			Notes:               req.Notes,
			Attachments:         attachments,
		})
	}
	return nil, healthuc.ErrInvalidInput
}

func (h *Handlers) updateProcedure(r *http.Request, userID, petID, procedureID uuid.UUID, req dto.CreateOrUpdateProcedureRequest) (*domainmodel.Procedure, error) {
	scheduledAt, err := parseOptionalRFC3339(req.ScheduledAt)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	performedAt, err := parseOptionalRFC3339(req.PerformedAt)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	nextDueAt, err := parseOptionalRFC3339(req.NextDueAt)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	catalogMedicationID, err := optionalUUIDFromString(req.CatalogMedicationID)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	procedureTypeID, err := optionalUUIDFromString(req.ProcedureTypeID)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	vetVisitID, err := optionalUUIDFromString(req.VetVisitID)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	attachments, err := parseAttachmentParams(req.Attachments)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	return h.procedures.UpdateProcedure(r.Context(), healthuc.UpdateProcedureParams{
		UserID:              userID,
		PetID:               petID,
		ProcedureID:         procedureID,
		RowVersion:          req.RowVersion,
		Status:              req.Status,
		ProcedureTypeID:     procedureTypeID,
		ProcedureTypeName:   req.ProcedureTypeName,
		Title:               req.Title,
		Description:         req.Description,
		CatalogMedicationID: catalogMedicationID,
		ProductName:         req.ProductName,
		ScheduledAt:         scheduledAt,
		Reminder:            medicalEntityReminderToUsecaseParams(req.Reminder),
		PerformedAt:         performedAt,
		NextDueAt:           nextDueAt,
		VetVisitID:          vetVisitID,
		Notes:               req.Notes,
		Attachments:         attachments,
	})
}
