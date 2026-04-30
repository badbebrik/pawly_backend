package handlers

import (
	healthuc "health/internal/application/usecase"
	domainmodel "health/internal/domain/model"
	"health/internal/transport/http/dto"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handlers) GetVaccinations(w http.ResponseWriter, r *http.Request) {
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
	resp, err := h.vaccinations.ListVaccinations(r.Context(), healthuc.ListVaccinationsParams{
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
	items := make([]dto.VaccinationListItemResponse, 0, len(resp.Items))
	for i := range resp.Items {
		items = append(items, vaccinationAppListItemToDTO(resp.Items[i]))
	}
	writeJSON(w, http.StatusOK, dto.VaccinationsListResponse{Items: items, NextCursor: encodeTimeCursor(resp.NextCursor)})
}

func (h *Handlers) GetVaccination(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	vaccinationID, err := uuid.Parse(chi.URLParam(r, "vaccination_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid vaccination_id")
		return
	}
	item, err := h.vaccinations.GetVaccination(r.Context(), userID, petID, vaccinationID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, vaccinationAppToDTO(item))
}

func (h *Handlers) CreateVaccination(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	var req dto.CreateOrUpdateVaccinationRequest
	if !decodeBody(w, r, &req) {
		return
	}
	item, err := h.createOrUpdateVaccination(r, userID, petID, uuid.Nil, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, vaccinationAppToDTO(item))
}

func (h *Handlers) UpdateVaccination(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	vaccinationID, err := uuid.Parse(chi.URLParam(r, "vaccination_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid vaccination_id")
		return
	}
	var req dto.CreateOrUpdateVaccinationRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "row_version is required")
		return
	}
	item, err := h.updateVaccination(r, userID, petID, vaccinationID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, vaccinationAppToDTO(item))
}

func (h *Handlers) DeleteVaccination(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	vaccinationID, err := uuid.Parse(chi.URLParam(r, "vaccination_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid vaccination_id")
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
	if err := h.vaccinations.DeleteVaccination(r.Context(), healthuc.DeleteVaccinationParams{UserID: userID, PetID: petID, VaccinationID: vaccinationID, RowVersion: req.RowVersion}); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) createOrUpdateVaccination(r *http.Request, userID, petID, vaccinationID uuid.UUID, req dto.CreateOrUpdateVaccinationRequest) (*domainmodel.Vaccination, error) {
	scheduledAt, err := parseOptionalRFC3339(req.ScheduledAt)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	administeredAt, err := parseOptionalRFC3339(req.AdministeredAt)
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
	vetVisitID, err := optionalUUIDFromString(req.VetVisitID)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	attachments, err := parseAttachmentParams(req.Attachments)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	targets, err := parseHealthDictionaryItemRefs(req.Targets)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	if vaccinationID == uuid.Nil {
		return h.vaccinations.CreateVaccination(r.Context(), healthuc.CreateVaccinationParams{
			UserID:              userID,
			PetID:               petID,
			Status:              req.Status,
			VaccineName:         req.VaccineName,
			CatalogMedicationID: catalogMedicationID,
			Targets:             targets,
			ScheduledAt:         scheduledAt,
			Reminder:            medicalEntityReminderToUsecaseParams(req.Reminder),
			AdministeredAt:      administeredAt,
			NextDueAt:           nextDueAt,
			VetVisitID:          vetVisitID,
			ClinicName:          req.ClinicName,
			VetName:             req.VetName,
			Notes:               req.Notes,
			Attachments:         attachments,
		})
	}
	return nil, healthuc.ErrInvalidInput
}

func (h *Handlers) updateVaccination(r *http.Request, userID, petID, vaccinationID uuid.UUID, req dto.CreateOrUpdateVaccinationRequest) (*domainmodel.Vaccination, error) {
	scheduledAt, err := parseOptionalRFC3339(req.ScheduledAt)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	administeredAt, err := parseOptionalRFC3339(req.AdministeredAt)
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
	vetVisitID, err := optionalUUIDFromString(req.VetVisitID)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	attachments, err := parseAttachmentParams(req.Attachments)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	targets, err := parseHealthDictionaryItemRefs(req.Targets)
	if err != nil {
		return nil, healthuc.ErrInvalidInput
	}
	return h.vaccinations.UpdateVaccination(r.Context(), healthuc.UpdateVaccinationParams{
		UserID:              userID,
		PetID:               petID,
		VaccinationID:       vaccinationID,
		RowVersion:          req.RowVersion,
		Status:              req.Status,
		VaccineName:         req.VaccineName,
		CatalogMedicationID: catalogMedicationID,
		Targets:             targets,
		ScheduledAt:         scheduledAt,
		Reminder:            medicalEntityReminderToUsecaseParams(req.Reminder),
		AdministeredAt:      administeredAt,
		NextDueAt:           nextDueAt,
		VetVisitID:          vetVisitID,
		ClinicName:          req.ClinicName,
		VetName:             req.VetName,
		Notes:               req.Notes,
		Attachments:         attachments,
	})
}
