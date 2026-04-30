package app

import (
	"net/http"

	"pawly/pkg/httpjson"
	"push/internal/transport/http/handlers"
)

func (a *App) setupRoutes() http.Handler {
	mux := http.NewServeMux()
	h := handlers.New(a.useCases)

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		httpjson.Write(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/v1/push/devices", h.HandleDevices)
	mux.HandleFunc("/v1/push/devices/", h.HandleDeviceByID)
	mux.HandleFunc("/v1/pets/", h.HandlePetPushSettings)

	return mux
}
