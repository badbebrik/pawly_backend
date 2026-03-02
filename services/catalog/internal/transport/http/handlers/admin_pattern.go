package handlers

import (
	dbrepo "catalog/internal/infrastructure/db/repository"
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

type AdminPatternHandler struct {
	svc  *service.CatalogService
	repo *dbrepo.PatternRepo
}

func NewAdminPatternHandler(svc *service.CatalogService, repo *dbrepo.PatternRepo) *AdminPatternHandler {
	return &AdminPatternHandler{svc: svc, repo: repo}
}

func (h *AdminPatternHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePatternRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	req.NameRu = strings.TrimSpace(req.NameRu)
	req.NameEn = strings.TrimSpace(req.NameEn)
	iconKey := strings.TrimSpace(req.IconKey)

	if req.NameRu == "" || req.NameEn == "" {
		http.Error(w, "name_ru and name_en required", http.StatusBadRequest)
		return
	}
	if iconKey == "" {
		http.Error(w, "icon_key required", http.StatusBadRequest)
		return
	}

	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}

	p := &model.Pattern{
		NameRu:   req.NameRu,
		NameEn:   req.NameEn,
		IconKey:  iconKey,
		IsActive: active,
	}

	newVer, err := h.svc.AdminCreatePattern(r.Context(), p)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":              p.ID,
		"catalog_version": newVer,
	})
}

func (h *AdminPatternHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	existing, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, dbrepo.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var req dto.UpdatePatternRequest
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
	if req.IconKey != nil {
		v := strings.TrimSpace(*req.IconKey)
		if v == "" {
			http.Error(w, "icon_key cannot be empty", http.StatusBadRequest)
			return
		}
		existing.IconKey = v
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	newVer, err := h.svc.AdminUpdatePattern(r.Context(), existing)
	if err != nil {
		if errors.Is(err, dbrepo.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":              existing.ID,
		"catalog_version": newVer,
	})
}
