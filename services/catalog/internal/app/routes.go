package app

import (
	appmw "catalog/internal/transport/http/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (a *App) routes() http.Handler {
	r := chi.NewRouter()

	r.Use(appmw.WithLocale("ru"))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/catalog/version", a.PublicHandler.GetVersion)
	r.Get("/catalog/species", a.PublicHandler.ListSpecies)
	r.Get("/catalog/breeds", a.PublicHandler.ListBreeds)
	r.Get("/catalog/colors", a.PublicHandler.ListColors)
	r.Get("/catalog/patterns", a.PublicHandler.ListPatterns)

	admin := chi.NewRouter()
	admin.Use(appmw.RequireAdminToken(a.cfg.AdminToken))

	admin.Post("/species", a.AdminSpecies.Create)
	admin.Patch("/species/{id}", a.AdminSpecies.Update)

	admin.Post("/colors", a.AdminColors.Create)
	admin.Patch("/colors/{id}", a.AdminColors.Update)

	admin.Post("/patterns", a.AdminPatterns.Create)
	admin.Patch("/patterns/{id}", a.AdminPatterns.Update)

	r.Mount("/admin", admin)

	return r
}
