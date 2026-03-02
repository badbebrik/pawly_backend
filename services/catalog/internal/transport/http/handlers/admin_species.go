package handlers

import (
	"catalog/internal/infrastructure/db/repository"
	"catalog/internal/model"
	"catalog/internal/service"
	"catalog/internal/transport/http/dto"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AdminSpeciesHandler struct {
	svc     *service.CatalogService
	species *repository.SpeciesRepo
}

func NewAdminSpeciesHandler(svc *service.CatalogService, species *repository.SpeciesRepo) *AdminSpeciesHandler {
	return &AdminSpeciesHandler{svc: svc, species: species}
}

func (h *AdminSpeciesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSpeciesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	req.NameRu = strings.TrimSpace(req.NameRu)
	req.NameEn = strings.TrimSpace(req.NameEn)
	if req.NameRu == "" || req.NameEn == "" {
		http.Error(w, "name_ru and name_en required", http.StatusBadRequest)
		return
	}

	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}

	sp := &model.Species{
		NameRu:   req.NameRu,
		NameEn:   req.NameEn,
		IsActive: active,
	}

	newVer, err := h.svc.AdminCreateSpecies(r.Context(), sp)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":              sp.ID,
		"catalog_version": newVer,
	})
}

func (h *AdminSpeciesHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	existing, err := h.species.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var req dto.UpdateSpeciesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.NameRu != nil {
		v := strings.TrimSpace(*req.NameRu)
		if v == "" {
			http.Error(w, "name_ru cannot be empty", http.StatusBadRequest)
			return
		}
		existing.NameRu = v
	}
	if req.NameEn != nil {
		v := strings.TrimSpace(*req.NameEn)
		if v == "" {
			http.Error(w, "name_en cannot be empty", http.StatusBadRequest)
			return
		}
		existing.NameEn = v
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	newVer, err := h.svc.AdminUpdateSpecies(r.Context(), existing)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":              existing.ID,
		"catalog_version": newVer,
	})
}
