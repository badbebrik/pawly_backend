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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
