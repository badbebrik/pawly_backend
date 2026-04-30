package handlers

import (
	healthuc "health/internal/application/usecase"
	domainmodel "health/internal/domain/model"
	"health/internal/transport/http/dto"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handlers) GetMedicalRecords(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	cursor, err := decodeTimeCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid cursor")
		return
	}
	var recordTypeID *uuid.UUID
	if rawRecordTypeID := r.URL.Query().Get("record_type_id"); rawRecordTypeID != "" {
		parsed, err := uuid.Parse(rawRecordTypeID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid record_type_id")
			return
		}
		recordTypeID = &parsed
	}
	resp, err := h.medicalRecords.ListMedicalRecords(r.Context(), healthuc.ListMedicalRecordsParams{
		UserID:       userID,
		PetID:        petID,
		Cursor:       cursor,
		Limit:        parseIntOrDefault(r.URL.Query().Get("limit"), 20),
		Q:            r.URL.Query().Get("q"),
		Status:       r.URL.Query().Get("status"),
		Bucket:       r.URL.Query().Get("bucket"),
		RecordTypeID: recordTypeID,
		Sort:         r.URL.Query().Get("sort"),
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	items := make([]dto.MedicalRecordListItemResponse, 0, len(resp.Items))
	for i := range resp.Items {
		items = append(items, medicalRecordAppListItemToDTO(resp.Items[i]))
	}
	writeJSON(w, http.StatusOK, dto.MedicalRecordsListResponse{Items: items, NextCursor: encodeTimeCursor(resp.NextCursor)})
}

func (h *Handlers) GetMedicalRecord(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	recordID, err := uuid.Parse(chi.URLParam(r, "record_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid record_id")
		return
	}
	item, err := h.medicalRecords.GetMedicalRecord(r.Context(), userID, petID, recordID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, medicalRecordAppToDTO(item))
}

func (h *Handlers) CreateMedicalRecord(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	var req dto.CreateOrUpdateMedicalRecordRequest
	if !decodeBody(w, r, &req) {
		return
	}
	item, err := h.createOrUpdateMedicalRecord(r, userID, petID, uuid.Nil, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, medicalRecordAppToDTO(item))
}

func (h *Handlers) UpdateMedicalRecord(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	recordID, err := uuid.Parse(chi.URLParam(r, "record_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid record_id")
		return
	}
	var req dto.CreateOrUpdateMedicalRecordRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "row_version is required")
		return
	}
	item, err := h.updateMedicalRecord(r, userID, petID, recordID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, medicalRecordAppToDTO(item))
}

func (h *Handlers) DeleteMedicalRecord(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	recordID, err := uuid.Parse(chi.URLParam(r, "record_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid record_id")
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
	if err := h.medicalRecords.DeleteMedicalRecord(r.Context(), healthuc.DeleteMedicalRecordParams{UserID: userID, PetID: petID, RecordID: recordID, RowVersion: req.RowVersion}); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) createOrUpdateMedicalRecord(r *http.Request, userID, petID, recordID uuid.UUID, req dto.CreateOrUpdateMedicalRecordRequest) (*domainmodel.MedicalRecord, error) {
	startedAt, err := parseOptionalRFC3339(req.StartedAt)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	resolvedAt, err := parseOptionalRFC3339(req.ResolvedAt)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	attachments, err := parseAttachmentParams(req.Attachments)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	recordTypeID, err := optionalUUIDFromString(req.RecordTypeID)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	if recordID == uuid.Nil {
		return h.medicalRecords.CreateMedicalRecord(r.Context(), healthuc.CreateMedicalRecordParams{
			UserID:         userID,
			PetID:          petID,
			RecordTypeID:   recordTypeID,
			RecordTypeName: req.RecordTypeName,
			Status:         req.Status,
			Title:          req.Title,
			Description:    req.Description,
			StartedAt:      startedAt,
			ResolvedAt:     resolvedAt,
			Attachments:    attachments,
		})
	}
	return nil, healthuc.ErrInvalidInput
}

func (h *Handlers) updateMedicalRecord(r *http.Request, userID, petID, recordID uuid.UUID, req dto.CreateOrUpdateMedicalRecordRequest) (*domainmodel.MedicalRecord, error) {
	startedAt, err := parseOptionalRFC3339(req.StartedAt)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	resolvedAt, err := parseOptionalRFC3339(req.ResolvedAt)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	attachments, err := parseAttachmentParams(req.Attachments)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	recordTypeID, err := optionalUUIDFromString(req.RecordTypeID)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	return h.medicalRecords.UpdateMedicalRecord(r.Context(), healthuc.UpdateMedicalRecordParams{
		UserID:         userID,
		PetID:          petID,
		RecordID:       recordID,
		RowVersion:     req.RowVersion,
		RecordTypeID:   recordTypeID,
		RecordTypeName: req.RecordTypeName,
		Status:         req.Status,
		Title:          req.Title,
		Description:    req.Description,
		StartedAt:      startedAt,
		ResolvedAt:     resolvedAt,
		Attachments:    attachments,
	})
}
