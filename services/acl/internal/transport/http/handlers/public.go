package handlers

import (
	"encoding/json"
	"net/http"
)

type PublicHandlers struct{}

func NewPublicHandlers() *PublicHandlers {
	return &PublicHandlers{}
}

func (h *PublicHandlers) NotImplemented(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"code":    "NOT_IMPLEMENTED",
			"message": "endpoint scaffolded, implementation pending",
		},
	})
}
