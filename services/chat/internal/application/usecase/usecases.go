package usecase

import "chat/internal/application/ports"

type UseCases struct {
	OpenDirectConversation *OpenDirectConversation
	ListConversations      *ListConversations
	GetConversation        *GetConversation
	GetUnreadSummary       *GetUnreadSummary
	GetMessageHistory      *GetMessageHistory
	SendMessage            *SendMessage
	MarkRead               *MarkRead
}

func New(d Deps) *UseCases {
	return &UseCases{
		OpenDirectConversation: &OpenDirectConversation{
			conversations: d.Conversations,
			participants:  d.Participants,
			tx:            d.Tx,
			acl:           d.ACL,
		},
		ListConversations: &ListConversations{
			read:     d.ConversationRead,
			profiles: d.Profiles,
			pets:     d.Pets,
		},
		GetConversation: &GetConversation{
			conversations: d.Conversations,
			participants:  d.Participants,
			profiles:      d.Profiles,
			pets:          d.Pets,
		},
		GetUnreadSummary: &GetUnreadSummary{
			read: d.ConversationRead,
		},
		GetMessageHistory: &GetMessageHistory{
			conversations: d.Conversations,
			participants:  d.Participants,
			messages:      d.Messages,
		},
		SendMessage: &SendMessage{
			conversations: d.Conversations,
			participants:  d.Participants,
			messages:      d.Messages,
			tx:            d.Tx,
			acl:           d.ACL,
			realtime:      d.Realtime,
		},
		MarkRead: &MarkRead{
			participants: d.Participants,
			tx:           d.Tx,
			realtime:     d.Realtime,
		},
	}
}

type OpenDirectConversation struct {
	conversations ports.ConversationRepository
	participants  ports.ParticipantRepository
	tx            ports.TxManager
	acl           ports.ACLClient
}

type ListConversations struct {
	read     ports.ConversationReadRepository
	profiles ports.ProfileClient
	pets     ports.PetClient
}

type GetConversation struct {
	conversations ports.ConversationRepository
	participants  ports.ParticipantRepository
	profiles      ports.ProfileClient
	pets          ports.PetClient
}

type GetUnreadSummary struct {
	read ports.ConversationReadRepository
}

type GetMessageHistory struct {
	conversations ports.ConversationRepository
	participants  ports.ParticipantRepository
	messages      ports.MessageRepository
}

type SendMessage struct {
	conversations ports.ConversationRepository
	participants  ports.ParticipantRepository
	messages      ports.MessageRepository
	tx            ports.TxManager
	acl           ports.ACLClient
	realtime      ports.RealtimePublisher
}

type MarkRead struct {
	participants ports.ParticipantRepository
	tx           ports.TxManager
	realtime     ports.RealtimePublisher
}
