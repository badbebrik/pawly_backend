package app

import (
	"gateway/internal/transport/http/handlers"
	appmw "gateway/internal/transport/http/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *App) setupRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	hs := handlers.NewFileHandlers(a.fileClient)

	r.Route("/v1", func(r chi.Router) {
		r.Use(appmw.JWTAuth(a.cfg.JWTSecret))

		r.Post("/files:init-upload", hs.InitUpload)
		r.Post("/files/{file_id}:confirm-upload", hs.ConfirmUpload)
		r.Get("/files/{file_id}:download-url", hs.GetDownloadURL)
		r.Post("/files:batch-download-urls", hs.BatchDownloadURLs)
		r.Post("/files:link", hs.LinkFile)
		r.Post("/files:unlink", hs.UnlinkFile)
		r.Get("/files/{file_id}", hs.GetFile)
		r.Get("/files/{file_id}/links", hs.ListFileLinks)
	})

	return r
}
