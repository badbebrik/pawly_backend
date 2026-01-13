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

	admin := chi.NewRouter()
	admin.Use(appmw.RequireAdminToken(a.cfg.AdminToken))

	admin.Post("/species", a.AdminSpecies.Create)
	admin.Patch("/species/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		r.URL.RawQuery = "id=" + id
		a.AdminSpecies.Update(w, r)
	})

	r.Mount("/admin", admin)

	return r
}
