package app

import (
	"net/http"
	"pawly/pkg/httpjson"
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
		httpjson.Write(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	h := handlers.New(a.useCases)

	r.Get("/v1/pet-dictionaries", h.GetDictionaries)

	r.Group(func(r chi.Router) {
		r.Use(appmw.WithUserID)
		r.Post("/v1/pets", h.CreatePet)
		r.Get("/v1/pets", h.ListPets)
		r.Get("/v1/pets/{pet_id}", h.GetPet)
		r.Put("/v1/pets/{pet_id}", h.UpdatePet)
		r.Post("/v1/pets/{pet_id}:transfer-ownership", h.TransferOwnership)
		r.Post("/v1/pets/{pet_id}:change-status", h.ChangePetStatus)
		r.Post("/v1/pets/{pet_id}/photo:init-upload", h.InitPetPhotoUpload)
		r.Post("/v1/pets/{pet_id}/photo:confirm-upload", h.ConfirmPetPhotoUpload)
		r.Delete("/v1/pets/{pet_id}/photo", h.DeletePetPhoto)
	})

	return r
}
