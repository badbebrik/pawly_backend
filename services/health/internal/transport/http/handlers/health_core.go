package handlers

import (
	"encoding/base64"
	"encoding/json"
	"health/internal/model"
	"health/internal/repository"
	"health/internal/service"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type deleteRowVersionRequest struct {
	RowVersion int `json:"row_version"`
}

type createOrUpdateVetVisitRequest struct {
	Status            string   `json:"status"`
	VisitType         string   `json:"visit_type"`
	ScheduledAt       *string  `json:"scheduled_at"`
	CompletedAt       *string  `json:"completed_at"`
	ReasonText        *string  `json:"reason_text"`
	ResultText        *string  `json:"result_text"`
	ClinicName        *string  `json:"clinic_name"`
	VetName           *string  `json:"vet_name"`
	AttachmentFileIDs []string `json:"attachment_file_ids"`
	RowVersion        int      `json:"row_version"`
}

type linkVetVisitLogRequest struct {
	LogID string `json:"log_id"`
}

type createOrUpdateVaccinationRequest struct {
	Status              string   `json:"status"`
	VaccineName         string   `json:"vaccine_name"`
	CatalogMedicationID *string  `json:"catalog_medication_id"`
	ScheduledAt         *string  `json:"scheduled_at"`
	AdministeredAt      *string  `json:"administered_at"`
	NextDueAt           *string  `json:"next_due_at"`
	VetVisitID          *string  `json:"vet_visit_id"`
	ClinicName          *string  `json:"clinic_name"`
	VetName             *string  `json:"vet_name"`
	Notes               *string  `json:"notes"`
	AttachmentFileIDs   []string `json:"attachment_file_ids"`
	RowVersion          int      `json:"row_version"`
}

type createOrUpdateProcedureRequest struct {
	Status              string   `json:"status"`
	ProcedureType       string   `json:"procedure_type"`
	Title               string   `json:"title"`
	Description         *string  `json:"description"`
	CatalogMedicationID *string  `json:"catalog_medication_id"`
	ProductName         *string  `json:"product_name"`
	ScheduledAt         *string  `json:"scheduled_at"`
	PerformedAt         *string  `json:"performed_at"`
	NextDueAt           *string  `json:"next_due_at"`
	VetVisitID          *string  `json:"vet_visit_id"`
	Notes               *string  `json:"notes"`
	AttachmentFileIDs   []string `json:"attachment_file_ids"`
	RowVersion          int      `json:"row_version"`
}

type createOrUpdateMedicalRecordRequest struct {
	RecordType        string   `json:"record_type"`
	Status            string   `json:"status"`
	Title             string   `json:"title"`
	Description       *string  `json:"description"`
	StartedAt         *string  `json:"started_at"`
	ResolvedAt        *string  `json:"resolved_at"`
	AttachmentFileIDs []string `json:"attachment_file_ids"`
	RowVersion        int      `json:"row_version"`
}

type initAttachmentUploadRequest struct {
	MimeType          string `json:"mime_type"`
	OriginalFilename  string `json:"original_filename"`
	ExpectedSizeBytes int64  `json:"expected_size_bytes"`
}

type confirmAttachmentUploadRequest struct {
	FileID    string `json:"file_id"`
	SizeBytes int64  `json:"size_bytes"`
}

func (h *Handlers) GetHealthBootstrap(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	resp, err := h.svc.GetHealthBootstrap(r.Context(), service.HealthBootstrapParams{UserID: userID, PetID: petID})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, healthBootstrapToDTO(resp))
}

func (h *Handlers) GetHealthDay(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	dateRaw := strings.TrimSpace(r.URL.Query().Get("date"))
	if dateRaw == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "date is required")
		return
	}
	date, err := time.Parse("2006-01-02", dateRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid date")
		return
	}
	items, err := h.svc.GetHealthDay(r.Context(), userID, petID, date)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	out := make([]any, 0, len(items))
	for i := range items {
		out = append(out, calendarDayItemToDTO(items[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"date": date.Format("2006-01-02"), "items": out})
}

func (h *Handlers) InitAttachmentUpload(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}

	var req initAttachmentUploadRequest
	if !decodeBody(w, r, &req) {
		return
	}

	fileID, upload, err := h.svc.InitAttachmentUpload(r.Context(), service.InitAttachmentUploadParams{
		UserID:            userID,
		PetID:             petID,
		MimeType:          req.MimeType,
		OriginalFilename:  req.OriginalFilename,
		ExpectedSizeBytes: req.ExpectedSizeBytes,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"file_id": fileID.String(),
		"upload": map[string]any{
			"method":     upload.Method,
			"url":        upload.URL,
			"headers":    upload.Headers,
			"expires_at": upload.ExpiresAt.UTC().Format(time.RFC3339),
		},
	})
}

func (h *Handlers) ConfirmAttachmentUpload(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}

	var req confirmAttachmentUploadRequest
	if !decodeBody(w, r, &req) {
		return
	}

	fileID, err := uuid.Parse(req.FileID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid file_id")
		return
	}

	file, err := h.svc.ConfirmAttachmentUpload(r.Context(), service.ConfirmAttachmentUploadParams{
		UserID:    userID,
		PetID:     petID,
		FileID:    fileID,
		SizeBytes: req.SizeBytes,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"file": map[string]any{
			"id":                file.ID.String(),
			"mime_type":         file.MimeType,
			"size_bytes":        file.SizeBytes,
			"original_filename": strOrNil(file.OriginalFilename),
		},
	})
}

func (h *Handlers) GetPetDocuments(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	cursor, err := decodeTimeCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid cursor")
		return
	}
	entityType := optionalQueryString(r, "entity_type")
	fileType := optionalQueryString(r, "file_type")
	resp, err := h.svc.ListPetDocuments(r.Context(), service.ListPetDocumentsParams{
		UserID:     userID,
		PetID:      petID,
		Cursor:     cursor,
		Limit:      parseIntOrDefault(r.URL.Query().Get("limit"), 20),
		EntityType: entityType,
		FileType:   fileType,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	items := make([]any, 0, len(resp.Items))
	for i := range resp.Items {
		items = append(items, petDocumentToDTO(resp.Items[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "next_cursor": encodeTimeCursor(resp.NextCursor)})
}

func (h *Handlers) GetVetVisits(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	cursor, err := decodeTimeCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid cursor")
		return
	}
	dateFrom, err := parseOptionalDateTime(r.URL.Query().Get("date_from"), false)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid date_from")
		return
	}
	dateTo, err := parseOptionalDateTime(r.URL.Query().Get("date_to"), true)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid date_to")
		return
	}
	resp, err := h.svc.ListVetVisits(r.Context(), service.ListVetVisitsParams{
		UserID:   userID,
		PetID:    petID,
		Cursor:   cursor,
		Limit:    parseIntOrDefault(r.URL.Query().Get("limit"), 20),
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
	items := make([]any, 0, len(resp.Items))
	for i := range resp.Items {
		items = append(items, vetVisitListItemToDTO(resp.Items[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "next_cursor": encodeTimeCursor(resp.NextCursor)})
}

func (h *Handlers) GetVetVisit(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	visitID, err := uuid.Parse(chi.URLParam(r, "visit_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid visit_id")
		return
	}
	item, err := h.svc.GetVetVisit(r.Context(), userID, petID, visitID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, vetVisitToDTO(item))
}

func (h *Handlers) CreateVetVisit(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	var req createOrUpdateVetVisitRequest
	if !decodeBody(w, r, &req) {
		return
	}
	scheduledAt, err := parseOptionalRFC3339(req.ScheduledAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid scheduled_at")
		return
	}
	completedAt, err := parseOptionalRFC3339(req.CompletedAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid completed_at")
		return
	}
	attachmentIDs, err := parseUUIDList(req.AttachmentFileIDs)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid attachment_file_ids")
		return
	}
	item, err := h.svc.CreateVetVisit(r.Context(), service.CreateVetVisitParams{UserID: userID, PetID: petID, Status: req.Status, VisitType: req.VisitType, ScheduledAt: scheduledAt, CompletedAt: completedAt, ReasonText: req.ReasonText, ResultText: req.ResultText, ClinicName: req.ClinicName, VetName: req.VetName, AttachmentFileIDs: attachmentIDs})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, vetVisitToDTO(item))
}

func (h *Handlers) UpdateVetVisit(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	visitID, err := uuid.Parse(chi.URLParam(r, "visit_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid visit_id")
		return
	}
	var req createOrUpdateVetVisitRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "row_version is required")
		return
	}
	scheduledAt, err := parseOptionalRFC3339(req.ScheduledAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid scheduled_at")
		return
	}
	completedAt, err := parseOptionalRFC3339(req.CompletedAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid completed_at")
		return
	}
	attachmentIDs, err := parseUUIDList(req.AttachmentFileIDs)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid attachment_file_ids")
		return
	}
	item, err := h.svc.UpdateVetVisit(r.Context(), service.UpdateVetVisitParams{UserID: userID, PetID: petID, VisitID: visitID, RowVersion: req.RowVersion, Status: req.Status, VisitType: req.VisitType, ScheduledAt: scheduledAt, CompletedAt: completedAt, ReasonText: req.ReasonText, ResultText: req.ResultText, ClinicName: req.ClinicName, VetName: req.VetName, AttachmentFileIDs: attachmentIDs})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, vetVisitToDTO(item))
}

func (h *Handlers) DeleteVetVisit(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	visitID, err := uuid.Parse(chi.URLParam(r, "visit_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid visit_id")
		return
	}
	var req deleteRowVersionRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "row_version is required")
		return
	}
	if err := h.svc.DeleteVetVisit(r.Context(), service.DeleteVetVisitParams{UserID: userID, PetID: petID, VisitID: visitID, RowVersion: req.RowVersion}); err != nil {
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
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid visit_id")
		return
	}
	var req linkVetVisitLogRequest
	if !decodeBody(w, r, &req) {
		return
	}
	logID, err := uuid.Parse(req.LogID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid log_id")
		return
	}
	item, err := h.svc.LinkVetVisitLog(r.Context(), userID, petID, visitID, logID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, relatedLogToDTO(*item))
}

func (h *Handlers) UnlinkVetVisitLog(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	visitID, err := uuid.Parse(chi.URLParam(r, "visit_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid visit_id")
		return
	}
	logID, err := uuid.Parse(chi.URLParam(r, "log_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid log_id")
		return
	}
	if err := h.svc.UnlinkVetVisitLog(r.Context(), userID, petID, visitID, logID); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) GetVaccinations(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	cursor, err := decodeTimeCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid cursor")
		return
	}
	dateFrom, err := parseOptionalDateTime(r.URL.Query().Get("date_from"), false)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid date_from")
		return
	}
	dateTo, err := parseOptionalDateTime(r.URL.Query().Get("date_to"), true)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid date_to")
		return
	}
	resp, err := h.svc.ListVaccinations(r.Context(), service.ListVaccinationsParams{UserID: userID, PetID: petID, Cursor: cursor, Limit: parseIntOrDefault(r.URL.Query().Get("limit"), 20), Status: r.URL.Query().Get("status"), Bucket: r.URL.Query().Get("bucket"), DateFrom: dateFrom, DateTo: dateTo, Sort: r.URL.Query().Get("sort")})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	items := make([]any, 0, len(resp.Items))
	for i := range resp.Items {
		items = append(items, vaccinationListItemToDTO(resp.Items[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "next_cursor": encodeTimeCursor(resp.NextCursor)})
}

func (h *Handlers) GetVaccination(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	vaccinationID, err := uuid.Parse(chi.URLParam(r, "vaccination_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid vaccination_id")
		return
	}
	item, err := h.svc.GetVaccination(r.Context(), userID, petID, vaccinationID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, vaccinationToDTO(item))
}

func (h *Handlers) CreateVaccination(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	var req createOrUpdateVaccinationRequest
	if !decodeBody(w, r, &req) {
		return
	}
	item, err := h.createOrUpdateVaccination(r, userID, petID, uuid.Nil, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, vaccinationToDTO(item))
}

func (h *Handlers) UpdateVaccination(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	vaccinationID, err := uuid.Parse(chi.URLParam(r, "vaccination_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid vaccination_id")
		return
	}
	var req createOrUpdateVaccinationRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "row_version is required")
		return
	}
	item, err := h.updateVaccination(r, userID, petID, vaccinationID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, vaccinationToDTO(item))
}

func (h *Handlers) DeleteVaccination(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	vaccinationID, err := uuid.Parse(chi.URLParam(r, "vaccination_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid vaccination_id")
		return
	}
	var req deleteRowVersionRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "row_version is required")
		return
	}
	if err := h.svc.DeleteVaccination(r.Context(), service.DeleteVaccinationParams{UserID: userID, PetID: petID, VaccinationID: vaccinationID, RowVersion: req.RowVersion}); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) GetProcedures(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	cursor, err := decodeTimeCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid cursor")
		return
	}
	dateFrom, err := parseOptionalDateTime(r.URL.Query().Get("date_from"), false)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid date_from")
		return
	}
	dateTo, err := parseOptionalDateTime(r.URL.Query().Get("date_to"), true)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid date_to")
		return
	}
	resp, err := h.svc.ListProcedures(r.Context(), service.ListProceduresParams{UserID: userID, PetID: petID, Cursor: cursor, Limit: parseIntOrDefault(r.URL.Query().Get("limit"), 20), Status: r.URL.Query().Get("status"), Bucket: r.URL.Query().Get("bucket"), ProcedureType: r.URL.Query().Get("procedure_type"), DateFrom: dateFrom, DateTo: dateTo, Sort: r.URL.Query().Get("sort")})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	items := make([]any, 0, len(resp.Items))
	for i := range resp.Items {
		items = append(items, procedureListItemToDTO(resp.Items[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "next_cursor": encodeTimeCursor(resp.NextCursor)})
}

func (h *Handlers) GetProcedure(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	procedureID, err := uuid.Parse(chi.URLParam(r, "procedure_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid procedure_id")
		return
	}
	item, err := h.svc.GetProcedure(r.Context(), userID, petID, procedureID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, procedureToDTO(item))
}

func (h *Handlers) CreateProcedure(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	var req createOrUpdateProcedureRequest
	if !decodeBody(w, r, &req) {
		return
	}
	item, err := h.createOrUpdateProcedure(r, userID, petID, uuid.Nil, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, procedureToDTO(item))
}

func (h *Handlers) UpdateProcedure(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	procedureID, err := uuid.Parse(chi.URLParam(r, "procedure_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid procedure_id")
		return
	}
	var req createOrUpdateProcedureRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "row_version is required")
		return
	}
	item, err := h.updateProcedure(r, userID, petID, procedureID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, procedureToDTO(item))
}

func (h *Handlers) DeleteProcedure(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	procedureID, err := uuid.Parse(chi.URLParam(r, "procedure_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid procedure_id")
		return
	}
	var req deleteRowVersionRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "row_version is required")
		return
	}
	if err := h.svc.DeleteProcedure(r.Context(), service.DeleteProcedureParams{UserID: userID, PetID: petID, ProcedureID: procedureID, RowVersion: req.RowVersion}); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) GetMedicalRecords(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	cursor, err := decodeTimeCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid cursor")
		return
	}
	resp, err := h.svc.ListMedicalRecords(r.Context(), service.ListMedicalRecordsParams{UserID: userID, PetID: petID, Cursor: cursor, Limit: parseIntOrDefault(r.URL.Query().Get("limit"), 20), Status: r.URL.Query().Get("status"), Bucket: r.URL.Query().Get("bucket"), RecordType: r.URL.Query().Get("record_type"), Sort: r.URL.Query().Get("sort")})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	items := make([]any, 0, len(resp.Items))
	for i := range resp.Items {
		items = append(items, medicalRecordListItemToDTO(resp.Items[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "next_cursor": encodeTimeCursor(resp.NextCursor)})
}

func (h *Handlers) GetMedicalRecord(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	recordID, err := uuid.Parse(chi.URLParam(r, "record_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid record_id")
		return
	}
	item, err := h.svc.GetMedicalRecord(r.Context(), userID, petID, recordID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, medicalRecordToDTO(item))
}

func (h *Handlers) CreateMedicalRecord(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	var req createOrUpdateMedicalRecordRequest
	if !decodeBody(w, r, &req) {
		return
	}
	item, err := h.createOrUpdateMedicalRecord(r, userID, petID, uuid.Nil, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, medicalRecordToDTO(item))
}

func (h *Handlers) UpdateMedicalRecord(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	recordID, err := uuid.Parse(chi.URLParam(r, "record_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid record_id")
		return
	}
	var req createOrUpdateMedicalRecordRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "row_version is required")
		return
	}
	item, err := h.updateMedicalRecord(r, userID, petID, recordID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, medicalRecordToDTO(item))
}

func (h *Handlers) DeleteMedicalRecord(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	recordID, err := uuid.Parse(chi.URLParam(r, "record_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid record_id")
		return
	}
	var req deleteRowVersionRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "row_version is required")
		return
	}
	if err := h.svc.DeleteMedicalRecord(r.Context(), service.DeleteMedicalRecordParams{UserID: userID, PetID: petID, RecordID: recordID, RowVersion: req.RowVersion}); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) createOrUpdateVaccination(r *http.Request, userID, petID, vaccinationID uuid.UUID, req createOrUpdateVaccinationRequest) (*model.Vaccination, error) {
	scheduledAt, err := parseOptionalRFC3339(req.ScheduledAt)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	administeredAt, err := parseOptionalRFC3339(req.AdministeredAt)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	nextDueAt, err := parseOptionalRFC3339(req.NextDueAt)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	catalogMedicationID, err := optionalUUIDFromString(req.CatalogMedicationID)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	vetVisitID, err := optionalUUIDFromString(req.VetVisitID)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	attachmentIDs, err := parseUUIDList(req.AttachmentFileIDs)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	if vaccinationID == uuid.Nil {
		return h.svc.CreateVaccination(r.Context(), service.CreateVaccinationParams{UserID: userID, PetID: petID, Status: req.Status, VaccineName: req.VaccineName, CatalogMedicationID: catalogMedicationID, ScheduledAt: scheduledAt, AdministeredAt: administeredAt, NextDueAt: nextDueAt, VetVisitID: vetVisitID, ClinicName: req.ClinicName, VetName: req.VetName, Notes: req.Notes, AttachmentFileIDs: attachmentIDs})
	}
	return nil, service.ErrInvalidInput
}

func (h *Handlers) updateVaccination(r *http.Request, userID, petID, vaccinationID uuid.UUID, req createOrUpdateVaccinationRequest) (*model.Vaccination, error) {
	scheduledAt, err := parseOptionalRFC3339(req.ScheduledAt)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	administeredAt, err := parseOptionalRFC3339(req.AdministeredAt)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	nextDueAt, err := parseOptionalRFC3339(req.NextDueAt)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	catalogMedicationID, err := optionalUUIDFromString(req.CatalogMedicationID)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	vetVisitID, err := optionalUUIDFromString(req.VetVisitID)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	attachmentIDs, err := parseUUIDList(req.AttachmentFileIDs)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	return h.svc.UpdateVaccination(r.Context(), service.UpdateVaccinationParams{UserID: userID, PetID: petID, VaccinationID: vaccinationID, RowVersion: req.RowVersion, Status: req.Status, VaccineName: req.VaccineName, CatalogMedicationID: catalogMedicationID, ScheduledAt: scheduledAt, AdministeredAt: administeredAt, NextDueAt: nextDueAt, VetVisitID: vetVisitID, ClinicName: req.ClinicName, VetName: req.VetName, Notes: req.Notes, AttachmentFileIDs: attachmentIDs})
}

func (h *Handlers) createOrUpdateProcedure(r *http.Request, userID, petID, procedureID uuid.UUID, req createOrUpdateProcedureRequest) (*model.Procedure, error) {
	scheduledAt, err := parseOptionalRFC3339(req.ScheduledAt)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	performedAt, err := parseOptionalRFC3339(req.PerformedAt)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	nextDueAt, err := parseOptionalRFC3339(req.NextDueAt)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	catalogMedicationID, err := optionalUUIDFromString(req.CatalogMedicationID)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	vetVisitID, err := optionalUUIDFromString(req.VetVisitID)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	attachmentIDs, err := parseUUIDList(req.AttachmentFileIDs)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	if procedureID == uuid.Nil {
		return h.svc.CreateProcedure(r.Context(), service.CreateProcedureParams{UserID: userID, PetID: petID, Status: req.Status, ProcedureType: req.ProcedureType, Title: req.Title, Description: req.Description, CatalogMedicationID: catalogMedicationID, ProductName: req.ProductName, ScheduledAt: scheduledAt, PerformedAt: performedAt, NextDueAt: nextDueAt, VetVisitID: vetVisitID, Notes: req.Notes, AttachmentFileIDs: attachmentIDs})
	}
	return nil, service.ErrInvalidInput
}

func (h *Handlers) updateProcedure(r *http.Request, userID, petID, procedureID uuid.UUID, req createOrUpdateProcedureRequest) (*model.Procedure, error) {
	scheduledAt, err := parseOptionalRFC3339(req.ScheduledAt)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	performedAt, err := parseOptionalRFC3339(req.PerformedAt)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	nextDueAt, err := parseOptionalRFC3339(req.NextDueAt)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	catalogMedicationID, err := optionalUUIDFromString(req.CatalogMedicationID)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	vetVisitID, err := optionalUUIDFromString(req.VetVisitID)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	attachmentIDs, err := parseUUIDList(req.AttachmentFileIDs)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	return h.svc.UpdateProcedure(r.Context(), service.UpdateProcedureParams{UserID: userID, PetID: petID, ProcedureID: procedureID, RowVersion: req.RowVersion, Status: req.Status, ProcedureType: req.ProcedureType, Title: req.Title, Description: req.Description, CatalogMedicationID: catalogMedicationID, ProductName: req.ProductName, ScheduledAt: scheduledAt, PerformedAt: performedAt, NextDueAt: nextDueAt, VetVisitID: vetVisitID, Notes: req.Notes, AttachmentFileIDs: attachmentIDs})
}

func (h *Handlers) createOrUpdateMedicalRecord(r *http.Request, userID, petID, recordID uuid.UUID, req createOrUpdateMedicalRecordRequest) (*model.MedicalRecord, error) {
	startedAt, err := parseOptionalRFC3339(req.StartedAt)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	resolvedAt, err := parseOptionalRFC3339(req.ResolvedAt)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	attachmentIDs, err := parseUUIDList(req.AttachmentFileIDs)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	if recordID == uuid.Nil {
		return h.svc.CreateMedicalRecord(r.Context(), service.CreateMedicalRecordParams{UserID: userID, PetID: petID, RecordType: req.RecordType, Status: req.Status, Title: req.Title, Description: req.Description, StartedAt: startedAt, ResolvedAt: resolvedAt, AttachmentFileIDs: attachmentIDs})
	}
	return nil, service.ErrInvalidInput
}

func (h *Handlers) updateMedicalRecord(r *http.Request, userID, petID, recordID uuid.UUID, req createOrUpdateMedicalRecordRequest) (*model.MedicalRecord, error) {
	startedAt, err := parseOptionalRFC3339(req.StartedAt)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	resolvedAt, err := parseOptionalRFC3339(req.ResolvedAt)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	attachmentIDs, err := parseUUIDList(req.AttachmentFileIDs)
	if err != nil {
		return nil, service.ErrInvalidInput
	}
	return h.svc.UpdateMedicalRecord(r.Context(), service.UpdateMedicalRecordParams{UserID: userID, PetID: petID, RecordID: recordID, RowVersion: req.RowVersion, RecordType: req.RecordType, Status: req.Status, Title: req.Title, Description: req.Description, StartedAt: startedAt, ResolvedAt: resolvedAt, AttachmentFileIDs: attachmentIDs})
}

func decodeBody(w http.ResponseWriter, r *http.Request, out any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return false
	}
	return true
}

func parseOptionalRFC3339(raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func encodeTimeCursor(cur *repository.TimeCursor) any {
	if cur == nil {
		return nil
	}
	raw := cur.SortAt.UTC().Format(time.RFC3339Nano) + "|" + cur.ID.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeTimeCursor(raw string) (*repository.TimeCursor, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(trimmed)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return nil, service.ErrInvalidInput
	}
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, err
	}
	return &repository.TimeCursor{SortAt: ts, ID: id}, nil
}

func healthBootstrapToDTO(item *model.HealthBootstrap) map[string]any {
	return map[string]any{
		"permissions": map[string]any{"health_read": item.Permissions.HealthRead, "health_write": item.Permissions.HealthWrite, "log_read": item.Permissions.LogRead},
		"enums": map[string]any{
			"vet_visit_statuses":      item.Enums.VetVisitStatuses,
			"vet_visit_types":         item.Enums.VetVisitTypes,
			"vaccination_statuses":    item.Enums.VaccinationStatuses,
			"procedure_statuses":      item.Enums.ProcedureStatuses,
			"procedure_types":         item.Enums.ProcedureTypes,
			"medical_record_types":    item.Enums.MedicalRecordTypes,
			"medical_record_statuses": item.Enums.MedicalRecordStatuses,
		},
	}
}

func vetVisitListItemToDTO(item model.VetVisitListItem) map[string]any {
	return map[string]any{
		"id": item.ID.String(), "pet_id": item.PetID.String(), "status": item.Status, "visit_type": item.VisitType,
		"scheduled_at": timeOrNil(item.ScheduledAt), "completed_at": timeOrNil(item.CompletedAt),
		"reason_text": strOrNil(item.ReasonText), "result_text": strOrNil(item.ResultText),
		"clinic_name": strOrNil(item.ClinicName), "vet_name": strOrNil(item.VetName),
		"related_logs_count": item.RelatedLogsCount, "attachments_count": item.AttachmentsCount,
		"row_version": item.RowVersion, "created_at": item.CreatedAt.UTC().Format(time.RFC3339), "created_by_user_id": item.CreatedByUserID.String(),
		"updated_at": item.UpdatedAt.UTC().Format(time.RFC3339), "updated_by_user_id": item.UpdatedByUserID.String(),
	}
}

func vetVisitToDTO(item *model.VetVisit) map[string]any {
	relatedLogs := make([]any, 0, len(item.RelatedLogs))
	for i := range item.RelatedLogs {
		relatedLogs = append(relatedLogs, relatedLogToDTO(item.RelatedLogs[i]))
	}
	attachments := make([]any, 0, len(item.Attachments))
	for i := range item.Attachments {
		attachments = append(attachments, healthAttachmentToDTO(item.Attachments[i]))
	}
	return map[string]any{
		"id": item.ID.String(), "pet_id": item.PetID.String(), "status": item.Status, "visit_type": item.VisitType,
		"scheduled_at": timeOrNil(item.ScheduledAt), "completed_at": timeOrNil(item.CompletedAt),
		"reason_text": strOrNil(item.ReasonText), "result_text": strOrNil(item.ResultText),
		"clinic_name": strOrNil(item.ClinicName), "vet_name": strOrNil(item.VetName),
		"related_logs": relatedLogs, "attachments": attachments,
		"row_version": item.RowVersion, "created_at": item.CreatedAt.UTC().Format(time.RFC3339), "created_by_user_id": item.CreatedByUserID.String(),
		"updated_at": item.UpdatedAt.UTC().Format(time.RFC3339), "updated_by_user_id": item.UpdatedByUserID.String(), "can_edit": true, "can_delete": true,
	}
}

func vaccinationListItemToDTO(item model.VaccinationListItem) map[string]any {
	return map[string]any{
		"id": item.ID.String(), "pet_id": item.PetID.String(), "status": item.Status, "vaccine_name": item.VaccineName,
		"catalog_medication_id": uuidOrNil(item.CatalogMedicationID), "scheduled_at": timeOrNil(item.ScheduledAt),
		"administered_at": timeOrNil(item.AdministeredAt), "next_due_at": timeOrNil(item.NextDueAt), "vet_visit_id": uuidOrNil(item.VetVisitID),
		"clinic_name": strOrNil(item.ClinicName), "vet_name": strOrNil(item.VetName), "notes_preview": strOrNil(item.NotesPreview),
		"attachments_count": item.AttachmentsCount, "row_version": item.RowVersion,
		"created_at": item.CreatedAt.UTC().Format(time.RFC3339), "created_by_user_id": item.CreatedByUserID.String(),
		"updated_at": item.UpdatedAt.UTC().Format(time.RFC3339), "updated_by_user_id": item.UpdatedByUserID.String(),
	}
}

func vaccinationToDTO(item *model.Vaccination) map[string]any {
	attachments := make([]any, 0, len(item.Attachments))
	for i := range item.Attachments {
		attachments = append(attachments, healthAttachmentToDTO(item.Attachments[i]))
	}
	return map[string]any{
		"id": item.ID.String(), "pet_id": item.PetID.String(), "status": item.Status, "vaccine_name": item.VaccineName,
		"catalog_medication_id": uuidOrNil(item.CatalogMedicationID), "scheduled_at": timeOrNil(item.ScheduledAt),
		"administered_at": timeOrNil(item.AdministeredAt), "next_due_at": timeOrNil(item.NextDueAt), "vet_visit_id": uuidOrNil(item.VetVisitID),
		"clinic_name": strOrNil(item.ClinicName), "vet_name": strOrNil(item.VetName), "notes": strOrNil(item.Notes),
		"attachments": attachments, "row_version": item.RowVersion,
		"created_at": item.CreatedAt.UTC().Format(time.RFC3339), "created_by_user_id": item.CreatedByUserID.String(),
		"updated_at": item.UpdatedAt.UTC().Format(time.RFC3339), "updated_by_user_id": item.UpdatedByUserID.String(), "can_edit": true, "can_delete": true,
	}
}

func procedureListItemToDTO(item model.ProcedureListItem) map[string]any {
	return map[string]any{
		"id": item.ID.String(), "pet_id": item.PetID.String(), "status": item.Status, "procedure_type": item.ProcedureType, "title": item.Title,
		"description_preview": strOrNil(item.DescriptionPreview), "catalog_medication_id": uuidOrNil(item.CatalogMedicationID), "product_name": strOrNil(item.ProductName),
		"scheduled_at": timeOrNil(item.ScheduledAt), "performed_at": timeOrNil(item.PerformedAt), "next_due_at": timeOrNil(item.NextDueAt), "vet_visit_id": uuidOrNil(item.VetVisitID),
		"notes_preview": strOrNil(item.NotesPreview), "attachments_count": item.AttachmentsCount, "row_version": item.RowVersion,
		"created_at": item.CreatedAt.UTC().Format(time.RFC3339), "created_by_user_id": item.CreatedByUserID.String(),
		"updated_at": item.UpdatedAt.UTC().Format(time.RFC3339), "updated_by_user_id": item.UpdatedByUserID.String(),
	}
}

func procedureToDTO(item *model.Procedure) map[string]any {
	attachments := make([]any, 0, len(item.Attachments))
	for i := range item.Attachments {
		attachments = append(attachments, healthAttachmentToDTO(item.Attachments[i]))
	}
	return map[string]any{
		"id": item.ID.String(), "pet_id": item.PetID.String(), "status": item.Status, "procedure_type": item.ProcedureType, "title": item.Title,
		"description": strOrNil(item.Description), "catalog_medication_id": uuidOrNil(item.CatalogMedicationID), "product_name": strOrNil(item.ProductName),
		"scheduled_at": timeOrNil(item.ScheduledAt), "performed_at": timeOrNil(item.PerformedAt), "next_due_at": timeOrNil(item.NextDueAt), "vet_visit_id": uuidOrNil(item.VetVisitID),
		"notes": strOrNil(item.Notes), "attachments": attachments, "row_version": item.RowVersion,
		"created_at": item.CreatedAt.UTC().Format(time.RFC3339), "created_by_user_id": item.CreatedByUserID.String(),
		"updated_at": item.UpdatedAt.UTC().Format(time.RFC3339), "updated_by_user_id": item.UpdatedByUserID.String(), "can_edit": true, "can_delete": true,
	}
}

func medicalRecordListItemToDTO(item model.MedicalRecordListItem) map[string]any {
	return map[string]any{
		"id": item.ID.String(), "pet_id": item.PetID.String(), "record_type": item.RecordType, "status": item.Status, "title": item.Title,
		"description_preview": strOrNil(item.DescriptionPreview), "started_at": timeOrNil(item.StartedAt), "resolved_at": timeOrNil(item.ResolvedAt),
		"attachments_count": item.AttachmentsCount, "row_version": item.RowVersion,
		"created_at": item.CreatedAt.UTC().Format(time.RFC3339), "created_by_user_id": item.CreatedByUserID.String(),
		"updated_at": item.UpdatedAt.UTC().Format(time.RFC3339), "updated_by_user_id": item.UpdatedByUserID.String(),
	}
}

func medicalRecordToDTO(item *model.MedicalRecord) map[string]any {
	attachments := make([]any, 0, len(item.Attachments))
	for i := range item.Attachments {
		attachments = append(attachments, healthAttachmentToDTO(item.Attachments[i]))
	}
	return map[string]any{
		"id": item.ID.String(), "pet_id": item.PetID.String(), "record_type": item.RecordType, "status": item.Status, "title": item.Title,
		"description": strOrNil(item.Description), "started_at": timeOrNil(item.StartedAt), "resolved_at": timeOrNil(item.ResolvedAt),
		"attachments": attachments, "row_version": item.RowVersion,
		"created_at": item.CreatedAt.UTC().Format(time.RFC3339), "created_by_user_id": item.CreatedByUserID.String(),
		"updated_at": item.UpdatedAt.UTC().Format(time.RFC3339), "updated_by_user_id": item.UpdatedByUserID.String(), "can_edit": true, "can_delete": true,
	}
}

func healthAttachmentToDTO(item model.HealthAttachment) map[string]any {
	return map[string]any{"id": item.ID.String(), "file_id": item.FileID.String(), "file_name": strOrNil(item.FileName), "file_type": item.FileType, "download_url": strOrNil(item.DownloadURL), "preview_url": strOrNil(item.PreviewURL), "added_by_user_id": item.AddedByUserID.String(), "added_at": item.AddedAt.UTC().Format(time.RFC3339)}
}

func petDocumentToDTO(item model.PetDocument) map[string]any {
	return map[string]any{
		"file_id":          item.FileID.String(),
		"file_name":        strOrNil(item.FileName),
		"file_type":        item.FileType,
		"download_url":     strOrNil(item.DownloadURL),
		"preview_url":      strOrNil(item.PreviewURL),
		"added_at":         item.AddedAt.UTC().Format(time.RFC3339),
		"added_by_user_id": item.AddedByUserID.String(),
		"entity_type":      item.EntityType,
		"entity_id":        item.EntityID.String(),
	}
}

func relatedLogToDTO(item model.RelatedLog) map[string]any {
	return map[string]any{"id": item.ID.String(), "occurred_at": item.OccurredAt.UTC().Format(time.RFC3339), "log_type_name": strOrNil(item.LogTypeName), "description_preview": strOrNil(item.DescriptionPreview), "source": item.Source}
}

func calendarDayItemToDTO(item model.CalendarDayItem) map[string]any {
	return map[string]any{
		"item_type": item.ItemType, "entity_id": item.EntityID.String(), "title": item.Title, "subtitle": strOrNil(item.Subtitle),
		"scheduled_for": item.ScheduledFor.UTC().Format(time.RFC3339), "status": item.Status,
		"source": map[string]any{"visit_id": uuidOrNil(item.VisitID), "vaccination_id": uuidOrNil(item.VaccinationID), "procedure_id": uuidOrNil(item.ProcedureID)},
	}
}

func timeOrNil(v *time.Time) any {
	if v == nil {
		return nil
	}
	return v.UTC().Format(time.RFC3339)
}
