package ws

import (
	"chat/internal/application/usecase"
	"chat/internal/infrastructure/realtime"
	appmw "chat/internal/transport/http/middleware"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

type Handler struct {
	hub      *realtime.Hub
	useCases *usecase.SendMessage
}

type inboundEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type subscribeConversationPayload struct {
	ConversationID uuid.UUID `json:"conversation_id"`
}

type sendMessagePayload struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	ClientMsgID    uuid.UUID `json:"client_msg_id"`
	Text           *string   `json:"text"`
}

type outboundEnvelope struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

type messageAckPayload struct {
	MessageID      uuid.UUID `json:"message_id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderUserID   uuid.UUID `json:"sender_user_id"`
	ClientMsgID    uuid.UUID `json:"client_msg_id"`
	Text           *string   `json:"text,omitempty"`
	CreatedAt      string    `json:"created_at"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewHandler(hub *realtime.Hub, sendMessage *usecase.SendMessage) *Handler {
	return &Handler{
		hub:      hub,
		useCases: sendMessage,
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
	h.readLoop(r.Context(), conn, client)
}

func (h *Handler) readLoop(ctx context.Context, conn *websocket.Conn, client *realtime.Client) {
	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var envelope inboundEnvelope
		if err := json.Unmarshal(payload, &envelope); err != nil {
			continue
		}

		switch envelope.Type {
		case "subscribe_inbox":
			h.hub.SubscribeInbox(client)
		case "subscribe_conversation":
			var msg subscribeConversationPayload
			if err := json.Unmarshal(envelope.Payload, &msg); err != nil {
				continue
			}
			if msg.ConversationID == uuid.Nil {
				continue
			}
			h.hub.SubscribeConversation(msg.ConversationID, client)
		case "unsubscribe_conversation":
			var msg subscribeConversationPayload
			if err := json.Unmarshal(envelope.Payload, &msg); err != nil {
				continue
			}
			if msg.ConversationID == uuid.Nil {
				continue
			}
			h.hub.UnsubscribeConversation(msg.ConversationID, client)
		case "send_message":
			var msg sendMessagePayload
			if err := json.Unmarshal(envelope.Payload, &msg); err != nil {
				continue
			}
			if msg.ConversationID == uuid.Nil || msg.ClientMsgID == uuid.Nil {
				continue
			}

			result, err := h.useCases.Execute(ctx, usecase.SendMessageParams{
				CurrentUserID:  client.UserID,
				ConversationID: msg.ConversationID,
				ClientMsgID:    msg.ClientMsgID,
				Text:           msg.Text,
			})
			if err != nil {
				continue
			}

			payload, err := json.Marshal(outboundEnvelope{
				Type: "message_ack",
				Payload: messageAckPayload{
					MessageID:      result.Message.MessageID,
					ConversationID: result.Message.ConversationID,
					SenderUserID:   result.Message.SenderUserID,
					ClientMsgID:    result.Message.ClientMsgID,
					Text:           result.Message.Text,
					CreatedAt:      result.Message.CreatedAt.UTC().Format(time.RFC3339),
				},
			})
			if err != nil {
				continue
			}

			h.hub.PublishToUser(client.UserID, payload)
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
