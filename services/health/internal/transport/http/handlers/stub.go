package handlers

import "net/http"

type StubHandlers struct{}

func NewStubHandlers() *StubHandlers {
	return &StubHandlers{}
}

func (h *StubHandlers) GetLogsBootstrap(w http.ResponseWriter, r *http.Request) { h.notImplemented(w) }
func (h *StubHandlers) GetLogs(w http.ResponseWriter, r *http.Request)          { h.notImplemented(w) }
func (h *StubHandlers) GetLog(w http.ResponseWriter, r *http.Request)           { h.notImplemented(w) }
func (h *StubHandlers) CreateLog(w http.ResponseWriter, r *http.Request)        { h.notImplemented(w) }
func (h *StubHandlers) UpdateLog(w http.ResponseWriter, r *http.Request)        { h.notImplemented(w) }
func (h *StubHandlers) DeleteLog(w http.ResponseWriter, r *http.Request)        { h.notImplemented(w) }

func (h *StubHandlers) GetLogTypes(w http.ResponseWriter, r *http.Request)   { h.notImplemented(w) }
func (h *StubHandlers) CreateLogType(w http.ResponseWriter, r *http.Request) { h.notImplemented(w) }
func (h *StubHandlers) UpdateLogType(w http.ResponseWriter, r *http.Request) { h.notImplemented(w) }
func (h *StubHandlers) DeleteLogType(w http.ResponseWriter, r *http.Request) { h.notImplemented(w) }
func (h *StubHandlers) GetMetrics(w http.ResponseWriter, r *http.Request)    { h.notImplemented(w) }
func (h *StubHandlers) CreateMetric(w http.ResponseWriter, r *http.Request)  { h.notImplemented(w) }
func (h *StubHandlers) UpdateMetric(w http.ResponseWriter, r *http.Request)  { h.notImplemented(w) }
func (h *StubHandlers) DeleteMetric(w http.ResponseWriter, r *http.Request)  { h.notImplemented(w) }
func (h *StubHandlers) GetAnalyticsMetrics(w http.ResponseWriter, r *http.Request) {
	h.notImplemented(w)
}
func (h *StubHandlers) GetMetricSeries(w http.ResponseWriter, r *http.Request) { h.notImplemented(w) }

func (h *StubHandlers) notImplemented(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(`{"error":{"code":"NOT_IMPLEMENTED","message":"health service endpoint is not implemented yet"}}`))
}
