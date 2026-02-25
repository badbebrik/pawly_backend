package handlers

import (
	"encoding/json"
	"net/http"
)

type InternalHandlers struct{}

func NewInternalHandlers() *InternalHandlers {
	return &InternalHandlers{}
}

func (h *InternalHandlers) NotImplemented(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"code":    "NOT_IMPLEMENTED",
			"message": "internal endpoint scaffolded, implementation pending",
		},
	})
}
