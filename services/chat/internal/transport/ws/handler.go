package ws

import (
	"chat/internal/infrastructure/realtime"
	appmw "chat/internal/transport/http/middleware"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Handler struct {
	hub *realtime.Hub
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewHandler(hub *realtime.Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	send := make(chan []byte, 32)
	client := realtime.NewClient(userID, send)
	h.hub.AddClient(client)
	defer func() {
		h.hub.RemoveClient(client)
		_ = conn.Close()
	}()

	go h.writeLoop(conn, client, send)
	h.readLoop(conn)
}

func (h *Handler) readLoop(conn *websocket.Conn) {
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *Handler) writeLoop(conn *websocket.Conn, client *realtime.Client, send <-chan []byte) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-client.Done():
			_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(5*time.Second))
			return
		case message := <-send:
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
				return
			}
		}
	}
}
