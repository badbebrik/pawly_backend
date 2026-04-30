package usecase

import "chat/internal/application/ports"

type Dependencies struct {
	Conversations ports.ConversationRepository
	Participants  ports.ParticipantRepository
	Messages      ports.MessageRepository
	TxManager     ports.TxManager
	ACLClient     ports.ACLClient
	ProfileClient ports.ProfileClient
	PetClient     ports.PetClient
	Presence      conversationPresenceReader
	Realtime      ports.RealtimePublisher
}

type Set struct {
	OpenDirectConversation *OpenDirectConversation
	ListConversations      *ListConversations
	GetConversation        *GetConversation
	GetUnreadSummary       *GetUnreadSummary
	GetMessageHistory      *GetMessageHistory
	SendMessage            *SendMessage
	MarkRead               *MarkRead
}

func NewSet(in Dependencies) *Set {
	return &Set{
		OpenDirectConversation: NewOpenDirectConversation(
			in.Conversations,
			in.Participants,
			in.TxManager,
			in.ACLClient,
			in.ProfileClient,
			in.PetClient,
			in.Presence,
		),
		ListConversations: NewListConversations(
			in.Conversations,
			in.ACLClient,
			in.ProfileClient,
			in.PetClient,
		),
		GetConversation: NewGetConversation(
			in.Conversations,
			in.Participants,
			in.ACLClient,
			in.ProfileClient,
			in.PetClient,
			in.Presence,
		),
		GetUnreadSummary: NewGetUnreadSummary(in.Conversations),
		GetMessageHistory: NewGetMessageHistory(
			in.Conversations,
			in.Participants,
			in.Messages,
		),
		SendMessage: NewSendMessage(
			in.Conversations,
			in.Participants,
			in.Messages,
			in.TxManager,
			in.ACLClient,
			in.Realtime,
		),
		MarkRead: NewMarkRead(
			in.Conversations,
			in.Participants,
			in.Messages,
			in.TxManager,
			in.Realtime,
		),
	}
}
