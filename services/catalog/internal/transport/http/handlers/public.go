package handlers

import (
	"catalog/internal/service"
	"catalog/internal/transport/http/dto"
	appmw "catalog/internal/transport/http/middleware"
	"encoding/json"
	"net/http"
	"strconv"
)

type PublicHandler struct {
	svc *service.CatalogService
}

func NewPublicHandler(svc *service.CatalogService) *PublicHandler {
	return &PublicHandler{svc: svc}
}

func (h *PublicHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.GetVersion(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, dto.VersionResponse{Version: v})
}

func (h *PublicHandler) ListSpecies(w http.ResponseWriter, r *http.Request) {
	locale := appmw.LocaleFromCtx(r.Context(), "ru")

	activeOnly := true
	if v := r.URL.Query().Get("active"); v != "" {
		if n, _ := strconv.Atoi(v); n == 0 {
			activeOnly = false
		}
	}

	items, err := h.svc.ListSpecies(r.Context(), activeOnly)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	out := make([]dto.SpeciesItem, 0, len(items))
	for _, s := range items {
		name := s.NameRu
		if locale == "en" {
			name = s.NameEn
		}
		out = append(out, dto.SpeciesItem{
			ID:       s.ID,
			Name:     name,
			IsActive: s.IsActive,
			Version:  s.Version,
		})
	}

	writeJSON(w, http.StatusOK, out)
}

func (h *PublicHandler) ListColors(w http.ResponseWriter, r *http.Request) {
	locale := appmw.LocaleFromCtx(r.Context(), "ru")

	activeOnly := true
	if r.URL.Query().Get("active") == "0" {
		activeOnly = false
	}

	items, err := h.svc.ListColors(r.Context(), activeOnly)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	out := make([]dto.ColorItem, 0, len(items))
	for _, c := range items {
		name := c.NameRu
		if locale == "en" {
			name = c.NameEn
		}
		out = append(out, dto.ColorItem{
			ID:       c.ID,
			Name:     name,
			Hex:      c.Hex,
			IsActive: c.IsActive,
			Version:  c.Version,
		})
	}

	writeJSON(w, http.StatusOK, out)
}

func (h *PublicHandler) ListPatterns(w http.ResponseWriter, r *http.Request) {
	locale := appmw.LocaleFromCtx(r.Context(), "ru")

	activeOnly := true
	if r.URL.Query().Get("active") == "0" {
		activeOnly = false
	}

	items, err := h.svc.ListPatterns(r.Context(), activeOnly)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	out := make([]dto.PatternItem, 0, len(items))
	for _, p := range items {
		name := p.NameRu
		if locale == "en" {
			name = p.NameEn
		}
		out = append(out, dto.PatternItem{
			ID:       p.ID,
			Name:     name,
			IconKey:  p.IconKey,
			IsActive: p.IsActive,
			Version:  p.Version,
		})
	}

	writeJSON(w, http.StatusOK, out)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
