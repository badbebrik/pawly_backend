package handlers

import (
	dbrepo "catalog/internal/infrastructure/db/repository"
	"catalog/internal/model"
	"catalog/internal/service"
	"catalog/internal/transport/http/dto"
	"catalog/internal/util"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
)

type AdminColorHandler struct {
	svc  *service.CatalogService
	repo *dbrepo.ColorRepo
}

func NewAdminColorHandler(svc *service.CatalogService, repo *dbrepo.ColorRepo) *AdminColorHandler {
	return &AdminColorHandler{svc: svc, repo: repo}
}

func (h *AdminColorHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateColorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	req.NameRu = strings.TrimSpace(req.NameRu)
	req.NameEn = strings.TrimSpace(req.NameEn)
	hex := strings.TrimSpace(req.Hex)

	if req.NameRu == "" || req.NameEn == "" {
		http.Error(w, "name_ru and name_en required", http.StatusBadRequest)
		return
	}
	if hex == "" || !util.IsHexColor(hex) {
		http.Error(w, "invalid hex (expected #RRGGBB)", http.StatusBadRequest)
		return
	}

	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}

	c := &model.Color{
		NameRu:   req.NameRu,
		NameEn:   req.NameEn,
		Hex:      hex,
		IsActive: active,
	}

	newVer, err := h.svc.AdminCreateColor(r.Context(), c)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":              c.ID,
		"catalog_version": newVer,
	})
}

func (h *AdminColorHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := util.AtoiPositive(idStr)
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

	var req dto.UpdateColorRequest
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
	if req.Hex != nil {
		v := strings.TrimSpace(*req.Hex)
		if v == "" || !util.IsHexColor(v) {
			http.Error(w, "invalid hex (expected #RRGGBB)", http.StatusBadRequest)
			return
		}
		existing.Hex = v
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	newVer, err := h.svc.AdminUpdateColor(r.Context(), existing)
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
