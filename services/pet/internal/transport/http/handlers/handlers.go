package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"pet/internal/model"
	"pet/internal/service"
	appmw "pet/internal/transport/http/middleware"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handlers struct {
	svc *service.PetService
}

func New(svc *service.PetService) *Handlers {
	return &Handlers{svc: svc}
}

type createPetRequest struct {
	Name                 string             `json:"name"`
	SpeciesID            string             `json:"species_id"`
	Sex                  string             `json:"sex"`
	BirthDate            *string            `json:"birth_date"`
	Breed                breedRequest       `json:"breed"`
	Colors               []model.Color      `json:"colors"`
	CoatPattern          coatPatternRequest `json:"coat_pattern"`
	IsNeutered           string             `json:"is_neutered"`
	IsOutdoor            bool               `json:"is_outdoor"`
	ProfilePhotoFileID   *string            `json:"profile_photo_file_id"`
	MicrochipID          *string            `json:"microchip_id"`
	MicrochipInstalledAt *string            `json:"microchip_installed_at"`
}

type updatePetRequest struct {
	RowVersion           int                `json:"row_version"`
	Name                 string             `json:"name"`
	SpeciesID            string             `json:"species_id"`
	Sex                  string             `json:"sex"`
	BirthDate            *string            `json:"birth_date"`
	Breed                breedRequest       `json:"breed"`
	Colors               []model.Color      `json:"colors"`
	CoatPattern          coatPatternRequest `json:"coat_pattern"`
	IsNeutered           string             `json:"is_neutered"`
	IsOutdoor            bool               `json:"is_outdoor"`
	ProfilePhotoFileID   *string            `json:"profile_photo_file_id"`
	MicrochipID          *string            `json:"microchip_id"`
	MicrochipInstalledAt *string            `json:"microchip_installed_at"`
}

type breedRequest struct {
	Source          string     `json:"source"`
	SystemBreedID   *uuid.UUID `json:"system_breed_id"`
	CustomBreedName *string    `json:"custom_breed_name"`
}

func (b breedRequest) toModel() model.Breed {
	return model.Breed{
		Source:          b.Source,
		SystemBreedID:   b.SystemBreedID,
		CustomBreedName: b.CustomBreedName,
	}
}

type coatPatternRequest struct {
	Source                string     `json:"source"`
	SystemCoatPatternID   *uuid.UUID `json:"system_coat_pattern_id"`
	CustomCoatPatternName *string    `json:"custom_coat_pattern_name"`
}

func (c coatPatternRequest) toModel() model.CoatPattern {
	return model.CoatPattern{
		Source:                c.Source,
		SystemCoatPatternID:   c.SystemCoatPatternID,
		CustomCoatPatternName: c.CustomCoatPatternName,
	}
}

type changeStatusRequest struct {
	RowVersion   int     `json:"row_version"`
	Status       string  `json:"status"`
	MissingSince *string `json:"missing_since"`
}

type initPetPhotoUploadRequest struct {
	MimeType          string `json:"mime_type"`
	OriginalFilename  string `json:"original_filename"`
	ExpectedSizeBytes int64  `json:"expected_size_bytes"`
}

type confirmPetPhotoUploadRequest struct {
	RowVersion int    `json:"row_version"`
	FileID     string `json:"file_id"`
	SizeBytes  int64  `json:"size_bytes"`
}

func (h *Handlers) CreatePet(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user id")
		return
	}

	var req createPetRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}

	speciesID, err := uuid.Parse(req.SpeciesID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid species_id")
		return
	}

	var (
		birthDate            *time.Time
		microchipInstalledAt *time.Time
		profilePhotoID       *uuid.UUID
	)
	if req.BirthDate != nil && *req.BirthDate != "" {
		t, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid birth_date")
			return
		}
		birthDate = &t
	}
	if req.MicrochipInstalledAt != nil && *req.MicrochipInstalledAt != "" {
		t, err := time.Parse("2006-01-02", *req.MicrochipInstalledAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid microchip_installed_at")
			return
		}
		microchipInstalledAt = &t
	}
	if req.ProfilePhotoFileID != nil && *req.ProfilePhotoFileID != "" {
		id, err := uuid.Parse(*req.ProfilePhotoFileID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid profile_photo_file_id")
			return
		}
		profilePhotoID = &id
	}

	pet, err := h.svc.CreatePet(r.Context(), service.CreatePetParams{
		UserID:               userID,
		Name:                 req.Name,
		SpeciesID:            speciesID,
		Sex:                  req.Sex,
		BirthDate:            birthDate,
		Breed:                req.Breed.toModel(),
		Colors:               req.Colors,
		CoatPattern:          req.CoatPattern.toModel(),
		IsNeutered:           req.IsNeutered,
		IsOutdoor:            req.IsOutdoor,
		ProfilePhotoFileID:   profilePhotoID,
		MicrochipID:          req.MicrochipID,
		MicrochipInstalledAt: microchipInstalledAt,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"pet": petToDTO(pet, h.getPetPhotoDownloadURL(r, pet))})
}

func (h *Handlers) ListPets(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user id")
		return
	}

	includeArchived := r.URL.Query().Get("include_archived") == "true"
	offset := parseIntOrDefault(r.URL.Query().Get("offset"), 0)
	limit := parseIntOrDefault(r.URL.Query().Get("limit"), 50)

	items, total, err := h.svc.ListPets(r.Context(), service.ListPetsParams{
		UserID:          userID,
		IncludeArchived: includeArchived,
		Offset:          offset,
		Limit:           limit,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	out := make([]any, 0, len(items))
	for i := range items {
		out = append(out, map[string]any{
			"pet":       petToDTO(&items[i].Pet, items[i].ProfilePhotoDownloadURL),
			"my_access": accessToDTO(items[i].MyAccess),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":  out,
		"total":  total,
		"offset": offset,
		"limit":  limit,
	})
}

func (h *Handlers) GetPet(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user id")
		return
	}

	petID, err := uuid.Parse(chi.URLParam(r, "pet_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid pet_id")
		return
	}

	pet, err := h.svc.GetPet(r.Context(), userID, petID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"pet": petToDTO(pet, h.getPetPhotoDownloadURL(r, pet))})
}

func (h *Handlers) UpdatePet(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user id")
		return
	}

	petID, err := uuid.Parse(chi.URLParam(r, "pet_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid pet_id")
		return
	}

	var req updatePetRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}

	speciesID, err := uuid.Parse(req.SpeciesID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid species_id")
		return
	}

	var (
		birthDate            *time.Time
		microchipInstalledAt *time.Time
		profilePhotoID       *uuid.UUID
	)
	if req.BirthDate != nil && *req.BirthDate != "" {
		t, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid birth_date")
			return
		}
		birthDate = &t
	}
	if req.MicrochipInstalledAt != nil && *req.MicrochipInstalledAt != "" {
		t, err := time.Parse("2006-01-02", *req.MicrochipInstalledAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid microchip_installed_at")
			return
		}
		microchipInstalledAt = &t
	}
	if req.ProfilePhotoFileID != nil && *req.ProfilePhotoFileID != "" {
		id, err := uuid.Parse(*req.ProfilePhotoFileID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid profile_photo_file_id")
			return
		}
		profilePhotoID = &id
	}

	pet, err := h.svc.UpdatePet(r.Context(), service.UpdatePetParams{
		UserID:               userID,
		PetID:                petID,
		RowVersion:           req.RowVersion,
		Name:                 req.Name,
		SpeciesID:            speciesID,
		Sex:                  req.Sex,
		BirthDate:            birthDate,
		Breed:                req.Breed.toModel(),
		Colors:               req.Colors,
		CoatPattern:          req.CoatPattern.toModel(),
		IsNeutered:           req.IsNeutered,
		IsOutdoor:            req.IsOutdoor,
		ProfilePhotoFileID:   profilePhotoID,
		MicrochipID:          req.MicrochipID,
		MicrochipInstalledAt: microchipInstalledAt,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"pet": petToDTO(pet, h.getPetPhotoDownloadURL(r, pet))})
}

func (h *Handlers) ChangePetStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user id")
		return
	}

	petID, err := uuid.Parse(chi.URLParam(r, "pet_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid pet_id")
		return
	}

	var req changeStatusRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}

	var missingSince *time.Time
	if req.MissingSince != nil && *req.MissingSince != "" {
		t, err := time.Parse(time.RFC3339, *req.MissingSince)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid missing_since")
			return
		}
		missingSince = &t
	}

	pet, err := h.svc.ChangePetStatus(r.Context(), service.ChangePetStatusParams{
		UserID:       userID,
		PetID:        petID,
		RowVersion:   req.RowVersion,
		Status:       req.Status,
		MissingSince: missingSince,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"pet": petToDTO(pet, h.getPetPhotoDownloadURL(r, pet))})
}

func (h *Handlers) InitPetPhotoUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user id")
		return
	}

	petID, err := uuid.Parse(chi.URLParam(r, "pet_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid pet_id")
		return
	}

	var req initPetPhotoUploadRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}

	fileID, upload, err := h.svc.InitPetPhotoUpload(r.Context(), service.InitPetPhotoUploadParams{
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

func (h *Handlers) ConfirmPetPhotoUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user id")
		return
	}

	petID, err := uuid.Parse(chi.URLParam(r, "pet_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid pet_id")
		return
	}

	var req confirmPetPhotoUploadRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}

	fileID, err := uuid.Parse(req.FileID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid file_id")
		return
	}

	pet, err := h.svc.ConfirmPetPhotoUpload(r.Context(), service.ConfirmPetPhotoUploadParams{
		UserID:     userID,
		PetID:      petID,
		RowVersion: req.RowVersion,
		FileID:     fileID,
		SizeBytes:  req.SizeBytes,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"pet": petToDTO(pet, h.getPetPhotoDownloadURL(r, pet))})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid input")
	case errors.Is(err, service.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "forbidden")
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "not found")
	case errors.Is(err, service.ErrConflict):
		writeError(w, http.StatusConflict, "CONFLICT", "conflict")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func parseIntOrDefault(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func petToDTO(p *model.Pet, profilePhotoDownloadURL *string) map[string]any {
	return map[string]any{
		"id":            p.ID.String(),
		"owner_user_id": p.OwnerUserID.String(),
		"row_version":   p.RowVersion,
		"name":          p.Name,
		"species_id":    p.SpeciesID.String(),
		"sex":           p.Sex,
		"birth_date":    dateOrNil(p.BirthDate),
		"breed": map[string]any{
			"source":            p.Breed.Source,
			"system_breed_id":   uuidOrNil(p.Breed.SystemBreedID),
			"custom_breed_name": strOrNil(p.Breed.CustomBreedName),
		},
		"colors": p.Colors,
		"coat_pattern": map[string]any{
			"source":                   p.CoatPattern.Source,
			"system_coat_pattern_id":   uuidOrNil(p.CoatPattern.SystemCoatPatternID),
			"custom_coat_pattern_name": strOrNil(p.CoatPattern.CustomCoatPatternName),
		},
		"is_neutered":                p.IsNeutered,
		"is_outdoor":                 p.IsOutdoor,
		"profile_photo_file_id":      uuidOrNil(p.ProfilePhotoFileID),
		"profile_photo_download_url": strOrNil(profilePhotoDownloadURL),
		"microchip_id":               strOrNil(p.MicrochipID),
		"microchip_installed_at":     dateOrNil(p.MicrochipInstalledAt),
		"status":                     p.Status,
		"missing_since":              tsOrNil(p.MissingSince),
		"archived_at":                tsOrNil(p.ArchivedAt),
		"created_at":                 p.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":                 p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func accessToDTO(access *service.ACLMembership) any {
	if access == nil {
		return nil
	}
	if access.MemberID == uuid.Nil {
		return nil
	}

	var roleCode any
	if access.Role.Code != nil {
		roleCode = *access.Role.Code
	}
	var rolePetID any
	if access.Role.PetID != nil {
		rolePetID = access.Role.PetID.String()
	}
	var roleCreatedBy any
	if access.Role.CreatedByUserID != nil {
		roleCreatedBy = access.Role.CreatedByUserID.String()
	}
	var roleID any
	if access.Role.ID != uuid.Nil {
		roleID = access.Role.ID.String()
	}

	return map[string]any{
		"member_id":        access.MemberID.String(),
		"status":           access.Status,
		"is_primary_owner": access.IsPrimaryOwner,
		"role": map[string]any{
			"id":                 roleID,
			"kind":               access.Role.Kind,
			"pet_id":             rolePetID,
			"code":               roleCode,
			"title":              access.Role.Title,
			"created_by_user_id": roleCreatedBy,
		},
		"policy": map[string]any{
			"pet_read":      access.Policy.PetRead,
			"pet_write":     access.Policy.PetWrite,
			"log_read":      access.Policy.LogRead,
			"log_write":     access.Policy.LogWrite,
			"health_read":   access.Policy.HealthRead,
			"health_write":  access.Policy.HealthWrite,
			"task_read":     access.Policy.TaskRead,
			"task_write":    access.Policy.TaskWrite,
			"members_read":  access.Policy.MembersRead,
			"members_write": access.Policy.MembersWrite,
		},
	}
}

func (h *Handlers) getPetPhotoDownloadURL(r *http.Request, pet *model.Pet) *string {
	return h.svc.ResolveProfilePhotoDownloadURL(r.Context(), pet.ProfilePhotoFileID)
}

func strOrNil(v *string) any {
	if v == nil {
		return nil
	}
	return *v
}

func uuidOrNil(v *uuid.UUID) any {
	if v == nil {
		return nil
	}
	return v.String()
}

func dateOrNil(v *time.Time) any {
	if v == nil {
		return nil
	}
	return v.UTC().Format("2006-01-02")
}

func tsOrNil(v *time.Time) any {
	if v == nil {
		return nil
	}
	return v.UTC().Format(time.RFC3339)
}
