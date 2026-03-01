package app

import (
	"net/http"
	"pet/internal/transport/http/handlers"
	appmw "pet/internal/transport/http/middleware"

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

	h := handlers.New(a.petSvc)

	r.Group(func(r chi.Router) {
		r.Use(appmw.WithUserID)
		r.Post("/v1/pets", h.CreatePet)
		r.Get("/v1/pets", h.ListPets)
		r.Get("/v1/pets/{pet_id}", h.GetPet)
		r.Put("/v1/pets/{pet_id}", h.UpdatePet)
		r.Post("/v1/pets/{pet_id}/status", h.ChangePetStatus)
	})

	return r
}
