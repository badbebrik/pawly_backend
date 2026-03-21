package usecase

import "chat/internal/application/ports"

type Deps struct {
	Conversations     ports.ConversationRepository
	Participants      ports.ParticipantRepository
	Messages          ports.MessageRepository
	ConversationRead  ports.ConversationReadRepository
	Tx                ports.TxManager
	ACL               ports.ACLClient
	Profiles          ports.ProfileClient
	Pets              ports.PetClient
	Realtime          ports.RealtimePublisher
}
