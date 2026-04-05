package ws

import (
	"chat/internal/application/usecase"
	"chat/internal/infrastructure/realtime"
	appmw "chat/internal/transport/http/middleware"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Handler struct {
	hub              *realtime.Hub
	publisher        realtime.EventPublisher
	presence         realtime.PresenceTracker
	presenceTTL      time.Duration
	heartbeatEvery   time.Duration
	jwtSecret        string
	sendMessage      *usecase.SendMessage
	markRead         *usecase.MarkRead
	getConversation  *usecase.GetConversation
	getUnreadSummary *usecase.GetUnreadSummary
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

type markReadPayload struct {
	ConversationID    uuid.UUID `json:"conversation_id"`
	LastReadMessageID uuid.UUID `json:"last_read_message_id"`
}

type outboundEnvelope struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

type errorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type messageAckPayload struct {
	MessageID      uuid.UUID `json:"message_id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderUserID   uuid.UUID `json:"sender_user_id"`
	ClientMsgID    uuid.UUID `json:"client_msg_id"`
	Text           *string   `json:"text,omitempty"`
	CreatedAt      string    `json:"created_at"`
}

type messageNewPayload struct {
	MessageID      uuid.UUID `json:"message_id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderUserID   uuid.UUID `json:"sender_user_id"`
	ClientMsgID    uuid.UUID `json:"client_msg_id"`
	Text           *string   `json:"text,omitempty"`
	CreatedAt      string    `json:"created_at"`
}

type readUpdatedPayload struct {
	ConversationID    uuid.UUID `json:"conversation_id"`
	UserID            uuid.UUID `json:"user_id"`
	LastReadMessageID uuid.UUID `json:"last_read_message_id"`
}

type conversationPetPayload struct {
	PetID     uuid.UUID `json:"pet_id"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
}

type conversationOtherUserPayload struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName *string   `json:"display_name,omitempty"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
}

type conversationUpdatedPayload struct {
	ConversationID             uuid.UUID                    `json:"conversation_id"`
	Pet                        conversationPetPayload       `json:"pet"`
	OtherUser                  conversationOtherUserPayload `json:"other_user"`
	LastMessageID              *uuid.UUID                   `json:"last_message_id,omitempty"`
	LastMessageAt              *string                      `json:"last_message_at,omitempty"`
	LastMessagePreview         *string                      `json:"last_message_preview,omitempty"`
	LastMessageSenderID        *uuid.UUID                   `json:"last_message_sender_id,omitempty"`
	LastReadMessageID          *uuid.UUID                   `json:"last_read_message_id,omitempty"`
	OtherUserLastReadMessageID *uuid.UUID                   `json:"other_user_last_read_message_id,omitempty"`
	OtherUserInChat            bool                         `json:"other_user_in_chat"`
	UnreadCount                int                          `json:"unread_count"`
	CanSend                    bool                         `json:"can_send"`
}

type globalUnreadUpdatedPayload struct {
	UnreadConversations int `json:"unread_conversations"`
	UnreadMessages      int `json:"unread_messages"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewHandler(
	hub *realtime.Hub,
	publisher realtime.EventPublisher,
	presence realtime.PresenceTracker,
	presenceTTL time.Duration,
	heartbeatEvery time.Duration,
	jwtSecret string,
	sendMessage *usecase.SendMessage,
	markRead *usecase.MarkRead,
	getConversation *usecase.GetConversation,
	getUnreadSummary *usecase.GetUnreadSummary,
) *Handler {
	return &Handler{
		hub:              hub,
		publisher:        publisher,
		presence:         presence,
		presenceTTL:      presenceTTL,
		heartbeatEvery:   heartbeatEvery,
		jwtSecret:        jwtSecret,
		sendMessage:      sendMessage,
		markRead:         markRead,
		getConversation:  getConversation,
		getUnreadSummary: getUnreadSummary,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmw.UserIDFromContext(r.Context())
	if !ok {
		var err error
		userID, err = h.userIDFromAuthorization(r.Header.Get("Authorization"))
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}
	if userID == uuid.Nil {
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
		h.clearPresence(client)
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
			h.publishError(client, "invalid_json", "invalid json")
			continue
		}

		switch envelope.Type {
		case "subscribe_inbox":
			h.hub.SubscribeInbox(client)
		case "subscribe_conversation":
			var msg subscribeConversationPayload
			if err := json.Unmarshal(envelope.Payload, &msg); err != nil {
				h.publishError(client, "invalid_payload", "invalid subscribe_conversation payload")
				continue
			}
			if msg.ConversationID == uuid.Nil {
				h.publishError(client, "invalid_payload", "conversation_id is required")
				continue
			}

			if _, err := h.getConversation.Execute(ctx, usecase.GetConversationParams{
				CurrentUserID:  client.UserID,
				ConversationID: msg.ConversationID,
			}); err != nil {
				h.publishError(client, "forbidden", "conversation access denied")
				continue
			}

			h.hub.SubscribeConversation(msg.ConversationID, client)
			client.AddConversationSubscription(msg.ConversationID)
			h.setPresence(ctx, client, msg.ConversationID)
		case "unsubscribe_conversation":
			var msg subscribeConversationPayload
			if err := json.Unmarshal(envelope.Payload, &msg); err != nil {
				h.publishError(client, "invalid_payload", "invalid unsubscribe_conversation payload")
				continue
			}
			if msg.ConversationID == uuid.Nil {
				h.publishError(client, "invalid_payload", "conversation_id is required")
				continue
			}
			h.hub.UnsubscribeConversation(msg.ConversationID, client)
			client.RemoveConversationSubscription(msg.ConversationID)
			h.clearConversationPresence(ctx, client, msg.ConversationID)
		case "send_message":
			var msg sendMessagePayload
			if err := json.Unmarshal(envelope.Payload, &msg); err != nil {
				h.publishError(client, "invalid_payload", "invalid send_message payload")
				continue
			}
			if msg.ConversationID == uuid.Nil || msg.ClientMsgID == uuid.Nil {
				h.publishError(client, "invalid_payload", "conversation_id and client_msg_id are required")
				continue
			}

			result, err := h.sendMessage.Execute(ctx, usecase.SendMessageParams{
				CurrentUserID:  client.UserID,
				ConversationID: msg.ConversationID,
				ClientMsgID:    msg.ClientMsgID,
				Text:           msg.Text,
			})
			if err != nil {
				h.publishError(client, "send_message_failed", "unable to send message")
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

			h.hub.PublishToClient(client, payload)

			conversation, err := h.getConversation.Execute(ctx, usecase.GetConversationParams{
				CurrentUserID:  client.UserID,
				ConversationID: msg.ConversationID,
			})
			if err != nil {
				h.publishError(client, "conversation_lookup_failed", "unable to refresh conversation")
				continue
			}

			if err := h.publisher.PublishMessageSent(ctx, realtime.MessageSentEvent{
				OriginClientID:  client.ID,
				ConversationID:  result.Message.ConversationID,
				MessageID:       result.Message.MessageID,
				SenderUserID:    result.Message.SenderUserID,
				RecipientUserID: conversation.Conversation.OtherUser.UserID,
				ClientMsgID:     result.Message.ClientMsgID,
				Text:            result.Message.Text,
				CreatedAt:       result.Message.CreatedAt,
			}); err != nil {
				h.publishError(client, "realtime_publish_failed", "message saved but realtime update failed")
			}
		case "mark_read":
			var msg markReadPayload
			if err := json.Unmarshal(envelope.Payload, &msg); err != nil {
				h.publishError(client, "invalid_payload", "invalid mark_read payload")
				continue
			}
			if msg.ConversationID == uuid.Nil || msg.LastReadMessageID == uuid.Nil {
				h.publishError(client, "invalid_payload", "conversation_id and last_read_message_id are required")
				continue
			}

			result, err := h.markRead.Execute(ctx, usecase.MarkReadParams{
				CurrentUserID:     client.UserID,
				ConversationID:    msg.ConversationID,
				LastReadMessageID: msg.LastReadMessageID,
			})
			if err != nil {
				h.publishError(client, "mark_read_failed", "unable to mark read")
				continue
			}

			if err := h.publisher.PublishReadUpdated(ctx, realtime.ReadUpdatedEvent{
				ConversationID:    result.ConversationID,
				UserID:            client.UserID,
				LastReadMessageID: result.LastReadMessageID,
			}); err != nil {
				h.publishError(client, "realtime_publish_failed", "read state saved but realtime update failed")
			}
		default:
			h.publishError(client, "unsupported_event", "unsupported event type")
		}
	}
}

func (h *Handler) writeLoop(conn *websocket.Conn, client *realtime.Client, send <-chan []byte) {
	ticker := time.NewTicker(h.heartbeatEvery)
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
			h.refreshPresence(client)
			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
				return
			}
		}
	}
}

func formatTimePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}

func (h *Handler) publishError(client *realtime.Client, code, message string) {
	payload, err := json.Marshal(outboundEnvelope{
		Type: "server_error",
		Payload: errorPayload{
			Code:    code,
			Message: message,
		},
	})
	if err != nil {
		return
	}

	h.hub.PublishToClient(client, payload)
}

func (h *Handler) setPresence(ctx context.Context, client *realtime.Client, conversationID uuid.UUID) {
	if h.presence == nil {
		return
	}

	change, changed, err := h.presence.SetInConversation(ctx, conversationID, client.UserID, client.ID, h.presenceTTL)
	if err != nil || !changed {
		return
	}
	h.publishPresenceChange(ctx, change)
}

func (h *Handler) clearConversationPresence(ctx context.Context, client *realtime.Client, conversationID uuid.UUID) {
	if h.presence == nil {
		return
	}

	change, changed, err := h.presence.ClearInConversation(ctx, conversationID, client.UserID, client.ID)
	if err != nil || !changed {
		return
	}
	h.publishPresenceChange(ctx, change)
}

func (h *Handler) clearPresence(client *realtime.Client) {
	if h.presence == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	changes, err := h.presence.ClearClient(ctx, client.UserID, client.ID, client.ConversationSubscriptions())
	if err != nil {
		return
	}
	for _, change := range changes {
		h.publishPresenceChange(ctx, change)
	}
}

func (h *Handler) refreshPresence(client *realtime.Client) {
	if h.presence == nil {
		return
	}

	subscriptions := client.ConversationSubscriptions()
	if len(subscriptions) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = h.presence.RefreshConversations(ctx, client.UserID, client.ID, subscriptions, h.presenceTTL)
}

func (h *Handler) publishPresenceChange(ctx context.Context, change realtime.PresenceChange) {
	if h.publisher == nil {
		return
	}
	_ = h.publisher.PublishConversationPresenceUpdated(ctx, realtime.ConversationPresenceUpdatedEvent{
		ConversationID: change.ConversationID,
		UserID:         change.UserID,
		IsInChat:       change.IsInChat,
	})
}

func (h *Handler) userIDFromAuthorization(authorization string) (uuid.UUID, error) {
	if h.jwtSecret == "" {
		return uuid.Nil, errors.New("jwt secret is empty")
	}
	if !strings.HasPrefix(authorization, "Bearer ") {
		return uuid.Nil, errors.New("authorization header missing")
	}

	tokenString := strings.TrimPrefix(authorization, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid claims")
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return uuid.Nil, errors.New("sub claim missing")
	}

	return uuid.Parse(sub)
}
