package handlers

import (
	healthuc "health/internal/application/usecase"
	"health/internal/transport/http/dto"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handlers) GetLogsBootstrap(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}

	includeCatalog := true
	if raw := r.URL.Query().Get("include_catalog"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid include_catalog")
			return
		}
		includeCatalog = v
	}

	data, err := h.logs.GetLogsBootstrap(r.Context(), healthuc.GetLogsBootstrapParams{
		UserID:         userID,
		PetID:          petID,
		IncludeCatalog: includeCatalog,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	recent := make([]dto.LogTypeResponse, 0, len(data.RecentLogTypes))
	for i := range data.RecentLogTypes {
		recent = append(recent, logTypeAppToDTO(data.RecentLogTypes[i]))
	}
	systemTypes := make([]dto.LogTypeResponse, 0, len(data.SystemLogTypes))
	for i := range data.SystemLogTypes {
		systemTypes = append(systemTypes, logTypeAppToDTO(data.SystemLogTypes[i]))
	}
	customTypes := make([]dto.LogTypeResponse, 0, len(data.CustomLogTypes))
	for i := range data.CustomLogTypes {
		customTypes = append(customTypes, logTypeAppToDTO(data.CustomLogTypes[i]))
	}
	systemMetrics := make([]dto.MetricResponse, 0, len(data.SystemMetrics))
	for i := range data.SystemMetrics {
		systemMetrics = append(systemMetrics, metricAppToDTO(data.SystemMetrics[i]))
	}
	customMetrics := make([]dto.MetricResponse, 0, len(data.CustomMetrics))
	for i := range data.CustomMetrics {
		customMetrics = append(customMetrics, metricAppToDTO(data.CustomMetrics[i]))
	}

	writeJSON(w, http.StatusOK, dto.LogsBootstrapResponse{
		Permissions: dto.PermissionsResponse{
			LogRead:  data.Permissions.LogRead,
			LogWrite: data.Permissions.LogWrite,
		},
		RecentLogTypes: recent,
		SystemLogTypes: systemTypes,
		CustomLogTypes: customTypes,
		SystemMetrics:  systemMetrics,
		CustomMetrics:  customMetrics,
	})
}

func (h *Handlers) GetLogs(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}

	cursor, err := decodeLogCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid cursor")
		return
	}

	limit := parseIntOrDefault(r.URL.Query().Get("limit"), 30)
	if limit < 1 || limit > 100 {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid limit")
		return
	}

	hasAttachments, err := parseOptionalBool(r.URL.Query(), "has_attachments")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid has_attachments")
		return
	}
	hasMetrics, err := parseOptionalBool(r.URL.Query(), "has_metrics")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid has_metrics")
		return
	}

	typeIDs, err := parseUUIDCSVOrMulti(r.URL.Query()["type_ids"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid type_ids")
		return
	}

	dateFrom, err := parseOptionalDateTime(r.URL.Query().Get("date_from"), false)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_from")
		return
	}
	dateTo, err := parseOptionalDateTime(r.URL.Query().Get("date_to"), true)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_to")
		return
	}

	sort := strings.TrimSpace(r.URL.Query().Get("sort"))
	if sort == "" {
		sort = "occurred_at_desc"
	}
	if sort != "occurred_at_desc" && sort != "occurred_at_asc" {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid sort")
		return
	}

	items, err := h.logs.ListLogs(r.Context(), healthuc.ListLogsParams{
		UserID:          userID,
		PetID:           petID,
		Cursor:          cursor,
		Limit:           limit,
		Sort:            sort,
		Q:               strings.TrimSpace(r.URL.Query().Get("q")),
		DateFrom:        dateFrom,
		DateTo:          dateTo,
		TypeIDs:         typeIDs,
		HasAttachments:  hasAttachments,
		HasMetricValues: hasMetrics,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	list := make([]dto.LogListItemResponse, 0, len(items.Items))
	for i := range items.Items {
		list = append(list, logAppListItemToDTO(items.Items[i]))
	}

	includeFacets := true
	if raw := r.URL.Query().Get("include_facets"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid include_facets")
			return
		}
		includeFacets = v
	}

	resp := dto.LogsListResponse{
		Items:      list,
		NextCursor: encodeLogCursor(items.NextCursor),
	}
	if includeFacets {
		resp.Facets = &dto.LogFacetsResponse{
			Types:               []any{},
			HasAttachmentsCount: 0,
			HasMetricsCount:     0,
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
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid log_id")
		return
	}

	item, err := h.logs.GetLog(r.Context(), userID, petID, logID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, logAppToDTO(item))
}

func (h *Handlers) CreateLog(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}

	var req dto.CreateOrUpdateLogRequest
	if !decodeBody(w, r, &req) {
		return
	}

	occurredAt, err := time.Parse(time.RFC3339, req.OccurredAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid occurred_at")
		return
	}
	logTypeID, err := optionalUUIDFromString(req.LogTypeID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid log_type_id")
		return
	}
	metricValues, err := parseMetricValues(req.MetricValues)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid metric_values")
		return
	}
	attachments, err := parseAttachmentParams(req.Attachments)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid attachments")
		return
	}

	item, err := h.logs.CreateLog(r.Context(), healthuc.CreateLogParams{
		UserID:       userID,
		PetID:        petID,
		OccurredAt:   occurredAt,
		LogTypeID:    logTypeID,
		Description:  req.Description,
		MetricValues: metricValues,
		Attachments:  attachments,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, logAppToDTO(item))
}

func (h *Handlers) UpdateLog(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	logID, err := uuid.Parse(chi.URLParam(r, "log_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid log_id")
		return
	}

	var req dto.CreateOrUpdateLogRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "row_version is required")
		return
	}

	occurredAt, err := time.Parse(time.RFC3339, req.OccurredAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid occurred_at")
		return
	}
	logTypeID, err := optionalUUIDFromString(req.LogTypeID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid log_type_id")
		return
	}
	metricValues, err := parseMetricValues(req.MetricValues)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid metric_values")
		return
	}
	attachments, err := parseAttachmentParams(req.Attachments)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid attachments")
		return
	}

	item, err := h.logs.UpdateLog(r.Context(), healthuc.UpdateLogParams{
		UserID:       userID,
		PetID:        petID,
		LogID:        logID,
		RowVersion:   req.RowVersion,
		OccurredAt:   occurredAt,
		LogTypeID:    logTypeID,
		Description:  req.Description,
		MetricValues: metricValues,
		Attachments:  attachments,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, logAppToDTO(item))
}

func (h *Handlers) DeleteLog(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	logID, err := uuid.Parse(chi.URLParam(r, "log_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid log_id")
		return
	}

	var req dto.DeleteLogRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "row_version is required")
		return
	}

	if err := h.logs.DeleteLog(r.Context(), healthuc.DeleteLogParams{
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
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid include_archived")
			return
		}
		includeArchived = v
	}
	onlyWithMetrics := false
	if raw := r.URL.Query().Get("only_with_metrics"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid only_with_metrics")
			return
		}
		onlyWithMetrics = v
	}

	items, err := h.logs.ListLogTypes(r.Context(), healthuc.ListLogTypesParams{
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

	out := make([]dto.LogTypeResponse, 0, len(items))
	for i := range items {
		out = append(out, logTypeAppToDTO(items[i]))
	}
	writeJSON(w, http.StatusOK, dto.LogTypesListResponse{Items: out})
}
func (h *Handlers) CreateLogType(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}

	var req dto.CreateLogTypeRequest
	if !decodeBody(w, r, &req) {
		return
	}
	metricRequirements, err := parseLogTypeRequirements(req.MetricRequirements)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid metric_requirements")
		return
	}

	item, err := h.logs.CreateLogType(r.Context(), healthuc.CreateLogTypeParams{
		UserID:             userID,
		PetID:              petID,
		Name:               req.Name,
		MetricRequirements: metricRequirements,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, logTypeAppToDTO(*item))
}
func (h *Handlers) UpdateLogType(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	logTypeID, err := uuid.Parse(chi.URLParam(r, "log_type_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid log_type_id")
		return
	}

	var req dto.UpdateLogTypeRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "row_version is required")
		return
	}
	metricRequirements, err := parseLogTypeRequirements(req.MetricRequirements)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid metric_requirements")
		return
	}

	item, err := h.logs.UpdateLogType(r.Context(), healthuc.UpdateLogTypeParams{
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
	writeJSON(w, http.StatusOK, logTypeAppToDTO(*item))
}
func (h *Handlers) DeleteLogType(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	logTypeID, err := uuid.Parse(chi.URLParam(r, "log_type_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid log_type_id")
		return
	}

	var req dto.DeleteLogRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "row_version is required")
		return
	}

	if err := h.logs.DeleteLogType(r.Context(), healthuc.DeleteLogTypeParams{
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
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid include_archived")
			return
		}
		includeArchived = v
	}
	onlyWithData := false
	if raw := r.URL.Query().Get("only_with_data"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid only_with_data")
			return
		}
		onlyWithData = v
	}
	onlyWithUsage := false
	if raw := r.URL.Query().Get("only_with_usage"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid only_with_usage")
			return
		}
		onlyWithUsage = v
	}

	items, err := h.logs.ListMetrics(r.Context(), healthuc.ListMetricsParams{
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

	out := make([]dto.MetricResponse, 0, len(items))
	for i := range items {
		out = append(out, metricAppToDTO(items[i]))
	}
	writeJSON(w, http.StatusOK, dto.MetricsListResponse{Items: out})
}
func (h *Handlers) CreateMetric(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}

	var req dto.CreateMetricRequest
	if !decodeBody(w, r, &req) {
		return
	}

	item, err := h.logs.CreateMetric(r.Context(), healthuc.CreateMetricParams{
		UserID:    userID,
		PetID:     petID,
		Name:      req.Name,
		InputKind: req.InputKind,
		Unit:      req.Unit,
		MinValue:  req.MinValue,
		MaxValue:  req.MaxValue,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, metricAppToDTO(*item))
}
func (h *Handlers) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	metricID, err := uuid.Parse(chi.URLParam(r, "metric_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid metric_id")
		return
	}

	var req dto.UpdateMetricRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "row_version is required")
		return
	}

	item, err := h.logs.UpdateMetric(r.Context(), healthuc.UpdateMetricParams{
		UserID:     userID,
		PetID:      petID,
		MetricID:   metricID,
		RowVersion: req.RowVersion,
		Name:       req.Name,
		InputKind:  req.InputKind,
		Unit:       req.Unit,
		MinValue:   req.MinValue,
		MaxValue:   req.MaxValue,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, metricAppToDTO(*item))
}
func (h *Handlers) DeleteMetric(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	metricID, err := uuid.Parse(chi.URLParam(r, "metric_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid metric_id")
		return
	}

	var req dto.DeleteLogRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.RowVersion <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_input", "row_version is required")
		return
	}

	if err := h.logs.DeleteMetric(r.Context(), healthuc.DeleteMetricParams{
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
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}

	typeIDs, err := parseUUIDCSVOrMulti(r.URL.Query()["type_ids"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid type_ids")
		return
	}

	dateFrom, err := parseOptionalDateTime(r.URL.Query().Get("date_from"), false)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_from")
		return
	}
	dateTo, err := parseOptionalDateTime(r.URL.Query().Get("date_to"), true)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_to")
		return
	}

	limit := parseIntOrDefault(r.URL.Query().Get("limit"), 100)
	if limit < 1 || limit > 500 {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid limit")
		return
	}

	items, err := h.analytics.ListAnalyticsMetrics(r.Context(), healthuc.ListAnalyticsMetricsParams{
		UserID:   userID,
		PetID:    petID,
		Q:        r.URL.Query().Get("q"),
		DateFrom: dateFrom,
		DateTo:   dateTo,
		TypeIDs:  typeIDs,
		Limit:    limit,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	out := make([]dto.AnalyticsMetricSummaryResponse, 0, len(items))
	for i := range items {
		out = append(out, analyticsMetricSummaryAppToDTO(items[i]))
	}
	writeJSON(w, http.StatusOK, dto.AnalyticsMetricsListResponse{Items: out})
}
func (h *Handlers) GetMetricSeries(w http.ResponseWriter, r *http.Request) {
	userID, petID, ok := h.getUserAndPet(w, r)
	if !ok {
		return
	}
	metricID, err := uuid.Parse(chi.URLParam(r, "metric_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid metric_id")
		return
	}

	typeIDs, err := parseUUIDCSVOrMulti(r.URL.Query()["type_ids"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid type_ids")
		return
	}

	dateFrom, err := parseOptionalDateTime(r.URL.Query().Get("date_from"), false)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_from")
		return
	}
	dateTo, err := parseOptionalDateTime(r.URL.Query().Get("date_to"), true)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid date_to")
		return
	}

	sort := strings.TrimSpace(r.URL.Query().Get("sort"))
	if sort == "" {
		sort = "occurred_at_asc"
	}
	if sort != "occurred_at_asc" && sort != "occurred_at_desc" {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid sort")
		return
	}

	includeSummary := true
	if raw := r.URL.Query().Get("include_summary"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_input", "invalid include_summary")
			return
		}
		includeSummary = v
	}

	result, err := h.analytics.GetMetricSeries(r.Context(), healthuc.GetMetricSeriesParams{
		UserID:         userID,
		PetID:          petID,
		MetricID:       metricID,
		DateFrom:       dateFrom,
		DateTo:         dateTo,
		TypeIDs:        typeIDs,
		Sort:           sort,
		IncludeSummary: includeSummary,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	points := make([]dto.MetricSeriesPointResponse, 0, len(result.Points))
	for i := range result.Points {
		points = append(points, metricSeriesPointAppToDTO(result.Points[i]))
	}

	writeJSON(w, http.StatusOK, dto.MetricSeriesResponse{
		Metric:  metricAppToDTO(result.Metric),
		Summary: metricSeriesSummaryAppToDTO(result.Summary),
		Points:  points,
	})
}
