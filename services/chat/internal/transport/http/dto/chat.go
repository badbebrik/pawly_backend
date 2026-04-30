package dto

import "github.com/google/uuid"

type OpenConversationRequest struct {
	PetID       uuid.UUID `json:"pet_id"`
	OtherUserID uuid.UUID `json:"other_user_id"`
}

type ConversationPetResponse struct {
	PetID     uuid.UUID `json:"pet_id"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
}

type ConversationOtherUserResponse struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName *string   `json:"display_name,omitempty"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
}

type ConversationResponse struct {
	ConversationID             uuid.UUID                     `json:"conversation_id"`
	Pet                        ConversationPetResponse       `json:"pet"`
	OtherUser                  ConversationOtherUserResponse `json:"other_user"`
	LastMessageID              *uuid.UUID                    `json:"last_message_id,omitempty"`
	LastMessageAt              *string                       `json:"last_message_at,omitempty"`
	LastMessagePreview         *string                       `json:"last_message_preview,omitempty"`
	LastMessageSenderID        *uuid.UUID                    `json:"last_message_sender_id,omitempty"`
	LastReadMessageID          *uuid.UUID                    `json:"last_read_message_id,omitempty"`
	OtherUserLastReadMessageID *uuid.UUID                    `json:"other_user_last_read_message_id,omitempty"`
	OtherUserInChat            bool                          `json:"other_user_in_chat"`
	UnreadCount                int                           `json:"unread_count"`
	CanSend                    bool                          `json:"can_send"`
}

type ListConversationsResponse struct {
	Items      []ConversationResponse `json:"items"`
	NextCursor *string                `json:"next_cursor,omitempty"`
}

type UnreadSummaryResponse struct {
	UnreadConversations int `json:"unread_conversations"`
	UnreadMessages      int `json:"unread_messages"`
}

type MessageResponse struct {
	MessageID      uuid.UUID `json:"message_id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderUserID   uuid.UUID `json:"sender_user_id"`
	ClientMsgID    uuid.UUID `json:"client_msg_id"`
	Text           *string   `json:"text,omitempty"`
	CreatedAt      string    `json:"created_at"`
}

type MessageHistoryResponse struct {
	ConversationID uuid.UUID         `json:"conversation_id"`
	Messages       []MessageResponse `json:"messages"`
	HasMore        bool              `json:"has_more"`
}

type MarkReadRequest struct {
	LastReadMessageID uuid.UUID `json:"last_read_message_id"`
}

type MarkReadResponse struct {
	ConversationID    uuid.UUID `json:"conversation_id"`
	LastReadMessageID uuid.UUID `json:"last_read_message_id"`
}
