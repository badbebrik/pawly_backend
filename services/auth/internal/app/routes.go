package app

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (a *App) setupRoutes() http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			return
		}
	})

	return r
}
