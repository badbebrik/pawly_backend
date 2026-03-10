package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"health/internal/model"
	"health/internal/repository"
	"health/internal/service"
	appmw "health/internal/transport/http/middleware"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handlers struct {
	svc *service.Service
}

func New(svc *service.Service) *Handlers {
	return &Handlers{svc: svc}
}

type metricValueRequest struct {
	MetricID string  `json:"metric_id"`
	ValueNum float64 `json:"value_num"`
}

type createOrUpdateLogRequest struct {
	OccurredAt        string               `json:"occurred_at"`
	LogTypeID         *string              `json:"log_type_id"`
	Description       *string              `json:"description"`
	MetricValues      []metricValueRequest `json:"metric_values"`
	AttachmentFileIDs []string             `json:"attachment_file_ids"`
	RowVersion        int                  `json:"row_version"`
}

type deleteLogRequest struct {
	RowVersion int `json:"row_version"`
}

type logTypeMetricRequirementRequest struct {
	MetricID   string `json:"metric_id"`
	IsRequired bool   `json:"is_required"`
}

type createLogTypeRequest struct {
	Name               string                            `json:"name"`
	MetricRequirements []logTypeMetricRequirementRequest `json:"metric_requirements"`
}

type updateLogTypeRequest struct {
	Name               string                            `json:"name"`
	MetricRequirements []logTypeMetricRequirementRequest `json:"metric_requirements"`
	RowVersion         int                               `json:"row_version"`
}

type createMetricRequest struct {
	Name      string   `json:"name"`
	InputKind string   `json:"input_kind"`
	UnitCode  *string  `json:"unit_code"`
	MinValue  *float64 `json:"min_value"`
	MaxValue  *float64 `json:"max_value"`
}

type updateMetricRequest struct {
	Name       string   `json:"name"`
	InputKind  string   `json:"input_kind"`
	UnitCode   *string  `json:"unit_code"`
	MinValue   *float64 `json:"min_value"`
	MaxValue   *float64 `json:"max_value"`
	RowVersion int      `json:"row_version"`
}

func (h *Handlers) GetLogsBootstrap(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}

	includeCatalog := true
	if raw := r.URL.Query().Get("include_catalog"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid include_catalog")
			return
		}
		includeCatalog = v
	}

	data, err := h.svc.GetLogsBootstrap(r.Context(), service.GetBootstrapParams{
		UserID:         userID,
		PetID:          petID,
		IncludeCatalog: includeCatalog,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	recent := make([]any, 0, len(data.RecentLogTypes))
	for i := range data.RecentLogTypes {
		recent = append(recent, logTypeToDTO(data.RecentLogTypes[i]))
	}
	systemTypes := make([]any, 0, len(data.SystemLogTypes))
	for i := range data.SystemLogTypes {
		systemTypes = append(systemTypes, logTypeToDTO(data.SystemLogTypes[i]))
	}
	customTypes := make([]any, 0, len(data.CustomLogTypes))
	for i := range data.CustomLogTypes {
		customTypes = append(customTypes, logTypeToDTO(data.CustomLogTypes[i]))
	}
	systemMetrics := make([]any, 0, len(data.SystemMetrics))
	for i := range data.SystemMetrics {
		systemMetrics = append(systemMetrics, metricToDTO(data.SystemMetrics[i]))
	}
	customMetrics := make([]any, 0, len(data.CustomMetrics))
	for i := range data.CustomMetrics {
		customMetrics = append(customMetrics, metricToDTO(data.CustomMetrics[i]))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"permissions": map[string]any{
			"log_read":  data.Permissions.LogRead,
			"log_write": data.Permissions.LogWrite,
		},
		"recent_log_types": recent,
		"system_log_types": systemTypes,
		"custom_log_types": customTypes,
		"system_metrics":   systemMetrics,
		"custom_metrics":   customMetrics,
	})
}

func (h *Handlers) GetLogs(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}

	cursor, err := decodeCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid cursor")
		return
	}

	limit := parseIntOrDefault(r.URL.Query().Get("limit"), 30)
	if limit < 1 || limit > 100 {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid limit")
		return
	}

	source := sourceFilterOrNil(r.URL.Query().Get("source"))
	if r.URL.Query().Has("source") && source == nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid source")
		return
	}

	hasAttachments, err := parseOptionalBool(r.URL.Query(), "has_attachments")
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid has_attachments")
		return
	}
	hasMetrics, err := parseOptionalBool(r.URL.Query(), "has_metrics")
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid has_metrics")
		return
	}

	typeIDs, err := parseUUIDCSVOrMulti(r.URL.Query()["type_ids"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid type_ids")
		return
	}

	dateFrom, err := parseOptionalDateTime(r.URL.Query().Get("date_from"), false)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid date_from")
		return
	}
	dateTo, err := parseOptionalDateTime(r.URL.Query().Get("date_to"), true)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid date_to")
		return
	}

	sort := strings.TrimSpace(r.URL.Query().Get("sort"))
	if sort == "" {
		sort = "occurred_at_desc"
	}
	if sort != "occurred_at_desc" && sort != "occurred_at_asc" {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid sort")
		return
	}

	items, err := h.svc.ListLogs(r.Context(), service.ListLogsParams{
		UserID:          userID,
		PetID:           petID,
		Cursor:          cursor,
		Limit:           limit,
		Sort:            sort,
		Q:               strings.TrimSpace(r.URL.Query().Get("q")),
		DateFrom:        dateFrom,
		DateTo:          dateTo,
		TypeIDs:         typeIDs,
		Source:          source,
		HasAttachments:  hasAttachments,
		HasMetricValues: hasMetrics,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	list := make([]any, 0, len(items.Items))
	for i := range items.Items {
		list = append(list, logListItemToDTO(items.Items[i]))
	}

	includeFacets := true
	if raw := r.URL.Query().Get("include_facets"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid include_facets")
			return
		}
		includeFacets = v
	}

	resp := map[string]any{
		"items":       list,
		"next_cursor": encodeCursor(items.NextCursor),
	}
	if includeFacets {
		resp["facets"] = map[string]any{
			"sources":               []any{},
			"types":                 []any{},
			"has_attachments_count": 0,
			"has_metrics_count":     0,
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) GetLog(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	logID, err := uuid.Parse(chi.URLParam(r, "log_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid log_id")
		return
	}

	item, err := h.svc.GetLog(r.Context(), userID, petID, logID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, logToDTO(item))
}

func (h *Handlers) CreateLog(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}

	var req createOrUpdateLogRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}

	occurredAt, err := time.Parse(time.RFC3339, req.OccurredAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid occurred_at")
		return
	}
	logTypeID, err := optionalUUIDFromString(req.LogTypeID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid log_type_id")
		return
	}
	metricValues, err := parseMetricValues(req.MetricValues)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid metric_values")
		return
	}
	attachmentIDs, err := parseUUIDList(req.AttachmentFileIDs)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid attachment_file_ids")
		return
	}

	item, err := h.svc.CreateLog(r.Context(), service.CreateLogParams{
		UserID:            userID,
		PetID:             petID,
		OccurredAt:        occurredAt,
		LogTypeID:         logTypeID,
		Description:       req.Description,
		MetricValues:      metricValues,
		AttachmentFileIDs: attachmentIDs,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, logToDTO(item))
}

func (h *Handlers) UpdateLog(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	logID, err := uuid.Parse(chi.URLParam(r, "log_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid log_id")
		return
	}

	var req createOrUpdateLogRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "row_version is required")
		return
	}

	occurredAt, err := time.Parse(time.RFC3339, req.OccurredAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid occurred_at")
		return
	}
	logTypeID, err := optionalUUIDFromString(req.LogTypeID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid log_type_id")
		return
	}
	metricValues, err := parseMetricValues(req.MetricValues)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid metric_values")
		return
	}
	attachmentIDs, err := parseUUIDList(req.AttachmentFileIDs)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid attachment_file_ids")
		return
	}

	item, err := h.svc.UpdateLog(r.Context(), service.UpdateLogParams{
		UserID:            userID,
		PetID:             petID,
		LogID:             logID,
		RowVersion:        req.RowVersion,
		OccurredAt:        occurredAt,
		LogTypeID:         logTypeID,
		Description:       req.Description,
		MetricValues:      metricValues,
		AttachmentFileIDs: attachmentIDs,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, logToDTO(item))
}

func (h *Handlers) DeleteLog(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	logID, err := uuid.Parse(chi.URLParam(r, "log_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid log_id")
		return
	}

	var req deleteLogRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "row_version is required")
		return
	}

	if err := h.svc.DeleteLog(r.Context(), service.DeleteLogParams{
		UserID:     userID,
		PetID:      petID,
		LogID:      logID,
		RowVersion: req.RowVersion,
	}); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) GetLogTypes(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}

	includeArchived := false
	if raw := r.URL.Query().Get("include_archived"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid include_archived")
			return
		}
		includeArchived = v
	}
	onlyWithMetrics := false
	if raw := r.URL.Query().Get("only_with_metrics"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid only_with_metrics")
			return
		}
		onlyWithMetrics = v
	}

	items, err := h.svc.ListLogTypes(r.Context(), service.ListLogTypesParams{
		UserID:          userID,
		PetID:           petID,
		Scope:           r.URL.Query().Get("scope"),
		Q:               r.URL.Query().Get("q"),
		IncludeArchived: includeArchived,
		OnlyWithMetrics: onlyWithMetrics,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	out := make([]any, 0, len(items))
	for i := range items {
		out = append(out, logTypeToDTO(items[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}
func (h *Handlers) CreateLogType(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}

	var req createLogTypeRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}
	metricRequirements, err := parseLogTypeRequirements(req.MetricRequirements)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid metric_requirements")
		return
	}

	item, err := h.svc.CreateLogType(r.Context(), service.CreateLogTypeParams{
		UserID:             userID,
		PetID:              petID,
		Name:               req.Name,
		MetricRequirements: metricRequirements,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, logTypeToDTO(*item))
}
func (h *Handlers) UpdateLogType(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	logTypeID, err := uuid.Parse(chi.URLParam(r, "log_type_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid log_type_id")
		return
	}

	var req updateLogTypeRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "row_version is required")
		return
	}
	metricRequirements, err := parseLogTypeRequirements(req.MetricRequirements)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid metric_requirements")
		return
	}

	item, err := h.svc.UpdateLogType(r.Context(), service.UpdateLogTypeParams{
		UserID:             userID,
		PetID:              petID,
		LogTypeID:          logTypeID,
		RowVersion:         req.RowVersion,
		Name:               req.Name,
		MetricRequirements: metricRequirements,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, logTypeToDTO(*item))
}
func (h *Handlers) DeleteLogType(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	logTypeID, err := uuid.Parse(chi.URLParam(r, "log_type_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid log_type_id")
		return
	}

	var req deleteLogRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "row_version is required")
		return
	}

	if err := h.svc.DeleteLogType(r.Context(), service.DeleteLogTypeParams{
		UserID:     userID,
		PetID:      petID,
		LogTypeID:  logTypeID,
		RowVersion: req.RowVersion,
	}); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handlers) GetMetrics(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}

	includeArchived := false
	if raw := r.URL.Query().Get("include_archived"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid include_archived")
			return
		}
		includeArchived = v
	}
	onlyWithData := false
	if raw := r.URL.Query().Get("only_with_data"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid only_with_data")
			return
		}
		onlyWithData = v
	}
	onlyWithUsage := false
	if raw := r.URL.Query().Get("only_with_usage"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid only_with_usage")
			return
		}
		onlyWithUsage = v
	}

	items, err := h.svc.ListMetrics(r.Context(), service.ListMetricsParams{
		UserID:          userID,
		PetID:           petID,
		Scope:           r.URL.Query().Get("scope"),
		Q:               r.URL.Query().Get("q"),
		IncludeArchived: includeArchived,
		OnlyWithData:    onlyWithData,
		OnlyWithUsage:   onlyWithUsage,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	out := make([]any, 0, len(items))
	for i := range items {
		out = append(out, metricToDTO(items[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}
func (h *Handlers) CreateMetric(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}

	var req createMetricRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}

	item, err := h.svc.CreateMetric(r.Context(), service.CreateMetricParams{
		UserID:    userID,
		PetID:     petID,
		Name:      req.Name,
		InputKind: req.InputKind,
		UnitCode:  req.UnitCode,
		MinValue:  req.MinValue,
		MaxValue:  req.MaxValue,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, metricToDTO(*item))
}
func (h *Handlers) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	metricID, err := uuid.Parse(chi.URLParam(r, "metric_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid metric_id")
		return
	}

	var req updateMetricRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "row_version is required")
		return
	}

	item, err := h.svc.UpdateMetric(r.Context(), service.UpdateMetricParams{
		UserID:     userID,
		PetID:      petID,
		MetricID:   metricID,
		RowVersion: req.RowVersion,
		Name:       req.Name,
		InputKind:  req.InputKind,
		UnitCode:   req.UnitCode,
		MinValue:   req.MinValue,
		MaxValue:   req.MaxValue,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, metricToDTO(*item))
}
func (h *Handlers) DeleteMetric(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	metricID, err := uuid.Parse(chi.URLParam(r, "metric_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid metric_id")
		return
	}

	var req deleteLogRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid body")
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "row_version is required")
		return
	}

	if err := h.svc.DeleteMetric(r.Context(), service.DeleteMetricParams{
		UserID:     userID,
		PetID:      petID,
		MetricID:   metricID,
		RowVersion: req.RowVersion,
	}); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handlers) GetAnalyticsMetrics(w http.ResponseWriter, r *http.Request) {
	h.notImplemented(w)
}
func (h *Handlers) GetMetricSeries(w http.ResponseWriter, r *http.Request) { h.notImplemented(w) }

func (h *Handlers) notImplemented(w http.ResponseWriter) {
	writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "health service endpoint is not implemented yet")
}

func (h *Handlers) getUserAndPet(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user id")
		return uuid.Nil, uuid.Nil, false
	}
	petID, err := uuid.Parse(chi.URLParam(r, "pet_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "invalid pet_id")
		return uuid.Nil, uuid.Nil, false
	}
	return userID, petID, true
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

func parseMetricValues(in []metricValueRequest) ([]service.CreateOrUpdateMetricValue, error) {
	out := make([]service.CreateOrUpdateMetricValue, 0, len(in))
	for i := range in {
		id, err := uuid.Parse(in[i].MetricID)
		if err != nil {
			return nil, err
		}
		out = append(out, service.CreateOrUpdateMetricValue{
			MetricID: id,
			ValueNum: in[i].ValueNum,
		})
	}
	return out, nil
}

func parseLogTypeRequirements(in []logTypeMetricRequirementRequest) ([]service.LogTypeRequirementInput, error) {
	out := make([]service.LogTypeRequirementInput, 0, len(in))
	for i := range in {
		id, err := uuid.Parse(in[i].MetricID)
		if err != nil {
			return nil, err
		}
		out = append(out, service.LogTypeRequirementInput{
			MetricID:   id,
			IsRequired: in[i].IsRequired,
		})
	}
	return out, nil
}

func parseUUIDList(in []string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(in))
	for i := range in {
		id, err := uuid.Parse(in[i])
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

func optionalUUIDFromString(raw *string) (*uuid.UUID, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	id, err := uuid.Parse(trimmed)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func parseOptionalBool(q map[string][]string, key string) (*bool, error) {
	vals, ok := q[key]
	if !ok || len(vals) == 0 {
		return nil, nil
	}
	v, err := strconv.ParseBool(vals[0])
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func parseUUIDCSVOrMulti(raw []string) ([]uuid.UUID, error) {
	if len(raw) == 0 {
		return []uuid.UUID{}, nil
	}
	pieces := make([]string, 0, len(raw))
	for i := range raw {
		for _, p := range strings.Split(raw[i], ",") {
			if strings.TrimSpace(p) != "" {
				pieces = append(pieces, strings.TrimSpace(p))
			}
		}
	}
	out := make([]uuid.UUID, 0, len(pieces))
	for i := range pieces {
		id, err := uuid.Parse(pieces[i])
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

func parseOptionalDateTime(raw string, endOfDay bool) (*time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	if strings.Contains(trimmed, "T") {
		t, err := time.Parse(time.RFC3339, trimmed)
		if err != nil {
			return nil, err
		}
		return &t, nil
	}
	t, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return nil, err
	}
	if endOfDay {
		t = t.Add(24*time.Hour - time.Nanosecond)
	}
	return &t, nil
}

func sourceFilterOrNil(raw string) *string {
	trimmed := strings.ToUpper(strings.TrimSpace(raw))
	if trimmed == "" || trimmed == "ALL" {
		return nil
	}
	if trimmed == "USER" || trimmed == "HEALTH" {
		return &trimmed
	}
	return nil
}

func encodeCursor(cur *repository.ListCursor) any {
	if cur == nil {
		return nil
	}
	raw := cur.OccurredAt.UTC().Format(time.RFC3339Nano) + "|" + cur.ID.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(raw string) (*repository.ListCursor, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(trimmed)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return nil, errors.New("bad cursor")
	}
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, err
	}
	return &repository.ListCursor{OccurredAt: ts, ID: id}, nil
}

func parseIntOrDefault(raw string, fallback int) int {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func logListItemToDTO(item model.LogListItem) map[string]any {
	return map[string]any{
		"id":                      item.ID.String(),
		"pet_id":                  item.PetID.String(),
		"occurred_at":             item.OccurredAt.UTC().Format(time.RFC3339),
		"log_type_id":             uuidOrNil(item.LogTypeID),
		"log_type_name":           strOrNil(item.LogTypeName),
		"log_type_scope":          strOrNil(item.LogTypeScope),
		"description_preview":     strOrNil(item.DescriptionPreview),
		"source":                  item.Source,
		"source_entity_type":      strOrNil(item.SourceEntityType),
		"source_entity_id":        uuidOrNil(item.SourceEntityID),
		"source_label":            nil,
		"metric_values_preview":   []any{},
		"attachments_count":       item.AttachmentsCount,
		"has_attachments":         item.HasAttachments,
		"created_by_user_id":      item.CreatedByUserID.String(),
		"created_by_display_name": nil,
	}
}

func logToDTO(item *model.Log) map[string]any {
	metricValues := make([]any, 0, len(item.MetricValues))
	for i := range item.MetricValues {
		mv := item.MetricValues[i]
		metricValues = append(metricValues, map[string]any{
			"metric_id":   mv.MetricID.String(),
			"metric_name": mv.MetricName,
			"input_kind":  mv.InputKind,
			"unit_code":   strOrNil(mv.UnitCode),
			"value_num":   mv.ValueNum,
		})
	}

	attachments := make([]any, 0, len(item.Attachments))
	for i := range item.Attachments {
		a := item.Attachments[i]
		attachments = append(attachments, map[string]any{
			"id":               a.ID.String(),
			"file_id":          a.FileID.String(),
			"file_name":        nil,
			"file_type":        a.FileType,
			"download_url":     strOrNil(a.DownloadURL),
			"preview_url":      strOrNil(a.PreviewURL),
			"added_at":         a.AddedAt.UTC().Format(time.RFC3339),
			"added_by_user_id": a.AddedByUserID.String(),
		})
	}

	canEdit := item.Source == "USER"
	return map[string]any{
		"id":                      item.ID.String(),
		"pet_id":                  item.PetID.String(),
		"occurred_at":             item.OccurredAt.UTC().Format(time.RFC3339),
		"log_type_id":             uuidOrNil(item.LogTypeID),
		"log_type_name":           strOrNil(item.LogTypeName),
		"log_type_scope":          strOrNil(item.LogTypeScope),
		"description":             strOrNil(item.Description),
		"source":                  item.Source,
		"source_entity_type":      strOrNil(item.SourceEntityType),
		"source_entity_id":        uuidOrNil(item.SourceEntityID),
		"source_label":            nil,
		"metric_values":           metricValues,
		"attachments":             attachments,
		"row_version":             item.RowVersion,
		"created_at":              item.CreatedAt.UTC().Format(time.RFC3339),
		"created_by_user_id":      item.CreatedByUserID.String(),
		"created_by_display_name": nil,
		"updated_at":              item.UpdatedAt.UTC().Format(time.RFC3339),
		"updated_by_user_id":      item.UpdatedByUserID.String(),
		"updated_by_display_name": nil,
		"can_edit":                canEdit,
		"can_delete":              canEdit,
	}
}

func logTypeToDTO(item model.LogType) map[string]any {
	metricRequirements := make([]any, 0, len(item.MetricRequirements))
	for i := range item.MetricRequirements {
		req := item.MetricRequirements[i]
		metricRequirements = append(metricRequirements, map[string]any{
			"metric_id":    req.MetricID.String(),
			"metric_name":  nil,
			"metric_scope": nil,
			"input_kind":   nil,
			"unit_code":    nil,
			"min_value":    nil,
			"max_value":    nil,
			"is_required":  req.IsRequired,
		})
	}

	return map[string]any{
		"id":                  item.ID.String(),
		"scope":               item.Scope,
		"pet_id":              uuidOrNil(item.PetID),
		"code":                strOrNil(item.Code),
		"name":                item.Name,
		"metric_requirements": metricRequirements,
		"created_at":          item.CreatedAt.UTC().Format(time.RFC3339),
		"created_by_user_id":  uuidOrNil(item.CreatedByUserID),
		"updated_at":          item.UpdatedAt.UTC().Format(time.RFC3339),
		"updated_by_user_id":  uuidOrNil(item.UpdatedByUserID),
		"row_version":         item.RowVersion,
		"is_archived":         item.DeletedAt != nil,
	}
}

func metricToDTO(item model.Metric) map[string]any {
	return map[string]any{
		"id":                 item.ID.String(),
		"scope":              item.Scope,
		"pet_id":             uuidOrNil(item.PetID),
		"code":               strOrNil(item.Code),
		"name":               item.Name,
		"input_kind":         item.InputKind,
		"unit_code":          strOrNil(item.UnitCode),
		"min_value":          float64OrNil(item.MinValue),
		"max_value":          float64OrNil(item.MaxValue),
		"created_at":         item.CreatedAt.UTC().Format(time.RFC3339),
		"created_by_user_id": uuidOrNil(item.CreatedByUserID),
		"updated_at":         item.UpdatedAt.UTC().Format(time.RFC3339),
		"updated_by_user_id": uuidOrNil(item.UpdatedByUserID),
		"row_version":        item.RowVersion,
		"is_archived":        item.DeletedAt != nil,
		"usage": map[string]any{
			"log_types_count": item.Usage.LogTypesCount,
			"logs_count":      item.Usage.LogsCount,
		},
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

func float64OrNil(v *float64) any {
	if v == nil {
		return nil
	}
	return *v
}
