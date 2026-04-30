package handlers

import "net/http"

func (h *Handlers) GetDictionaries(w http.ResponseWriter, r *http.Request) {
	data, err := h.useCases.GetDictionaries(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dictionariesToResponse(data))
}
