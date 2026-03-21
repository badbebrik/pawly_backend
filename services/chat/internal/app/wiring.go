package app

import (
	"chat/internal/application/usecase"
	chatdb "chat/internal/infrastructure/db"
	"chat/internal/infrastructure/repository"
)

func (a *App) wire() error {
	pg, err := chatdb.NewPostgres(a.cfg)
	if err != nil {
		return err
	}
	a.pg = pg

	conversations := repository.NewConversationRepository(pg.Pool)
	participants := repository.NewParticipantRepository(pg.Pool)
	txManager := chatdb.NewTxManager(pg.Pool)

	a.useCases = &UseCases{
		OpenDirectConversation: usecase.NewOpenDirectConversation(conversations, participants, txManager, nil, nil, nil),
		ListConversations:      usecase.NewListConversations(nil, nil, nil),
		GetConversation:        usecase.NewGetConversation(nil, nil, nil, nil),
		GetUnreadSummary:       usecase.NewGetUnreadSummary(nil),
		GetMessageHistory:      usecase.NewGetMessageHistory(nil, nil, nil),
		SendMessage:            usecase.NewSendMessage(conversations, participants, nil, txManager, nil, nil),
		MarkRead:               usecase.NewMarkRead(participants, txManager, nil),
	}

	return nil
}
