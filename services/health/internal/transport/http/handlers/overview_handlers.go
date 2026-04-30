package handlers

import (
	"health/internal/transport/http/dto"
	"net/http"
	"strings"
	"time"
)

func (h *Handlers) GetHealthBootstrap(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	resp, err := h.overview.GetHealthBootstrap(r.Context(), userID, petID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, healthBootstrapToDTO(resp))
}

func (h *Handlers) GetGlobalHealthDay(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(w, r)
	if !ok {
		return
	}
	dateRaw := strings.TrimSpace(r.URL.Query().Get("date"))
	if dateRaw == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "date is required")
		return
	}
	date, err := time.Parse("2006-01-02", dateRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date")
		return
	}
	items, err := h.overview.GetGlobalHealthDay(r.Context(), userID, date)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	out := make([]dto.CalendarDayItemResponse, 0, len(items))
	for i := range items {
		out = append(out, calendarDayItemAppToDTO(items[i]))
	}
	writeJSON(w, http.StatusOK, dto.DayResponse{Date: date.Format("2006-01-02"), Items: out})
}

func (h *Handlers) GetGlobalHealthCalendar(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(w, r)
	if !ok {
		return
	}
	dateFromRaw := strings.TrimSpace(r.URL.Query().Get("date_from"))
	if dateFromRaw == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "date_from is required")
		return
	}
	dateToRaw := strings.TrimSpace(r.URL.Query().Get("date_to"))
	if dateToRaw == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "date_to is required")
		return
	}
	dateFrom, err := time.Parse("2006-01-02", dateFromRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_from")
		return
	}
	dateTo, err := time.Parse("2006-01-02", dateToRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_to")
		return
	}
	items, err := h.overview.GetGlobalHealthCalendar(r.Context(), userID, dateFrom, dateTo)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	out := make([]dto.CalendarMarkerResponse, 0, len(items))
	for i := range items {
		out = append(out, calendarDateMarkerAppToDTO(items[i]))
	}
	writeJSON(w, http.StatusOK, dto.CalendarRangeResponse{
		DateFrom: dateFrom.Format("2006-01-02"),
		DateTo:   dateTo.Format("2006-01-02"),
		Items:    out,
	})
}

func (h *Handlers) GetHealthDay(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	dateRaw := strings.TrimSpace(r.URL.Query().Get("date"))
	if dateRaw == "" {
		writeError(w, http.StatusBadRequest, "invalid_input", "date is required")
		return
	}
	date, err := time.Parse("2006-01-02", dateRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date")
		return
	}
	items, err := h.overview.GetHealthDay(r.Context(), userID, petID, date)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	out := make([]dto.CalendarDayItemResponse, 0, len(items))
	for i := range items {
		out = append(out, calendarDayItemAppToDTO(items[i]))
	}
	writeJSON(w, http.StatusOK, dto.DayResponse{Date: date.Format("2006-01-02"), Items: out})
}
