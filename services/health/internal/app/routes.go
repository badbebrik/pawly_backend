package app

import (
	"health/internal/transport/http/handlers"
	appmw "health/internal/transport/http/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *App) setupRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	h := handlers.New(a.logSvc)

	r.Group(func(r chi.Router) {
		r.Use(appmw.WithUserID)
		r.Get("/v1/pets/{pet_id}/health/bootstrap", h.GetHealthBootstrap)
		r.Get("/v1/pets/{pet_id}/health/day", h.GetHealthDay)

		r.Get("/v1/pets/{pet_id}/vet-visits", h.GetVetVisits)
		r.Post("/v1/pets/{pet_id}/vet-visits", h.CreateVetVisit)
		r.Get("/v1/pets/{pet_id}/vet-visits/{visit_id}", h.GetVetVisit)
		r.Patch("/v1/pets/{pet_id}/vet-visits/{visit_id}", h.UpdateVetVisit)
		r.Delete("/v1/pets/{pet_id}/vet-visits/{visit_id}", h.DeleteVetVisit)
		r.Post("/v1/pets/{pet_id}/vet-visits/{visit_id}/logs", h.LinkVetVisitLog)
		r.Delete("/v1/pets/{pet_id}/vet-visits/{visit_id}/logs/{log_id}", h.UnlinkVetVisitLog)

		r.Get("/v1/pets/{pet_id}/vaccinations", h.GetVaccinations)
		r.Post("/v1/pets/{pet_id}/vaccinations", h.CreateVaccination)
		r.Get("/v1/pets/{pet_id}/vaccinations/{vaccination_id}", h.GetVaccination)
		r.Patch("/v1/pets/{pet_id}/vaccinations/{vaccination_id}", h.UpdateVaccination)
		r.Delete("/v1/pets/{pet_id}/vaccinations/{vaccination_id}", h.DeleteVaccination)

		r.Get("/v1/pets/{pet_id}/procedures", h.GetProcedures)
		r.Post("/v1/pets/{pet_id}/procedures", h.CreateProcedure)
		r.Get("/v1/pets/{pet_id}/procedures/{procedure_id}", h.GetProcedure)
		r.Patch("/v1/pets/{pet_id}/procedures/{procedure_id}", h.UpdateProcedure)
		r.Delete("/v1/pets/{pet_id}/procedures/{procedure_id}", h.DeleteProcedure)

		r.Get("/v1/pets/{pet_id}/medical-records", h.GetMedicalRecords)
		r.Post("/v1/pets/{pet_id}/medical-records", h.CreateMedicalRecord)
		r.Get("/v1/pets/{pet_id}/medical-records/{record_id}", h.GetMedicalRecord)
		r.Patch("/v1/pets/{pet_id}/medical-records/{record_id}", h.UpdateMedicalRecord)
		r.Delete("/v1/pets/{pet_id}/medical-records/{record_id}", h.DeleteMedicalRecord)

		r.Get("/v1/pets/{pet_id}/logs/bootstrap", h.GetLogsBootstrap)
		r.Get("/v1/pets/{pet_id}/logs", h.GetLogs)
		r.Get("/v1/pets/{pet_id}/logs/{log_id}", h.GetLog)
		r.Post("/v1/pets/{pet_id}/logs", h.CreateLog)
		r.Patch("/v1/pets/{pet_id}/logs/{log_id}", h.UpdateLog)
		r.Delete("/v1/pets/{pet_id}/logs/{log_id}", h.DeleteLog)

		r.Get("/v1/pets/{pet_id}/log-types", h.GetLogTypes)
		r.Post("/v1/pets/{pet_id}/log-types", h.CreateLogType)
		r.Patch("/v1/pets/{pet_id}/log-types/{log_type_id}", h.UpdateLogType)
		r.Delete("/v1/pets/{pet_id}/log-types/{log_type_id}", h.DeleteLogType)

		r.Get("/v1/pets/{pet_id}/metrics", h.GetMetrics)
		r.Post("/v1/pets/{pet_id}/metrics", h.CreateMetric)
		r.Patch("/v1/pets/{pet_id}/metrics/{metric_id}", h.UpdateMetric)
		r.Delete("/v1/pets/{pet_id}/metrics/{metric_id}", h.DeleteMetric)

		r.Get("/v1/pets/{pet_id}/analytics/metrics", h.GetAnalyticsMetrics)
		r.Get("/v1/pets/{pet_id}/analytics/metrics/{metric_id}/series", h.GetMetricSeries)
	})

	return r
}
