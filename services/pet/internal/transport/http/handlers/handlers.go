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
	Name                 string            `json:"name"`
	SpeciesID            string            `json:"species_id"`
	Sex                  string            `json:"sex"`
	BirthDate            *string           `json:"birth_date"`
	Breed                model.Breed       `json:"breed"`
	Colors               []model.Color     `json:"colors"`
	CoatPattern          model.CoatPattern `json:"coat_pattern"`
	IsNeutered           string            `json:"is_neutered"`
	IsOutdoor            bool              `json:"is_outdoor"`
	ProfilePhotoFileID   *string           `json:"profile_photo_file_id"`
	MicrochipID          *string           `json:"microchip_id"`
	MicrochipInstalledAt *string           `json:"microchip_installed_at"`
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
		Breed:                req.Breed,
		Colors:               req.Colors,
		CoatPattern:          req.CoatPattern,
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

	writeJSON(w, http.StatusCreated, map[string]any{"pet": petToDTO(pet)})
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
		out = append(out, map[string]any{"pet": petToDTO(&items[i])})
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

	writeJSON(w, http.StatusOK, map[string]any{"pet": petToDTO(pet)})
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

func petToDTO(p *model.Pet) map[string]any {
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
		"is_neutered":            p.IsNeutered,
		"is_outdoor":             p.IsOutdoor,
		"profile_photo_file_id":  uuidOrNil(p.ProfilePhotoFileID),
		"microchip_id":           strOrNil(p.MicrochipID),
		"microchip_installed_at": dateOrNil(p.MicrochipInstalledAt),
		"status":                 p.Status,
		"missing_since":          tsOrNil(p.MissingSince),
		"archived_at":            tsOrNil(p.ArchivedAt),
		"created_at":             p.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":             p.UpdatedAt.UTC().Format(time.RFC3339),
	}
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
