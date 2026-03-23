package app

import (
	"chat/internal/application/usecase"
	aclclient "chat/internal/infrastructure/aclclient"
	chatdb "chat/internal/infrastructure/db"
	petclient "chat/internal/infrastructure/petclient"
	profileclient "chat/internal/infrastructure/profileclient"
	"chat/internal/infrastructure/repository"
)

func (a *App) wire() error {
	pg, err := chatdb.NewPostgres(a.cfg)
	if err != nil {
		return err
	}
	a.pg = pg

	acl, err := aclclient.New(a.cfg.ACLGRPCAddr)
	if err != nil {
		pg.Close()
		return err
	}
	a.acl = acl

	profile, err := profileclient.New(a.cfg.ProfileGRPCAddr)
	if err != nil {
		acl.Close()
		pg.Close()
		return err
	}
	a.profile = profile

	pet, err := petclient.New(a.cfg.PetGRPCAddr)
	if err != nil {
		profile.Close()
		acl.Close()
		pg.Close()
		return err
	}
	a.pet = pet

	conversations := repository.NewConversationRepository(pg.Pool)
	participants := repository.NewParticipantRepository(pg.Pool)
	txManager := chatdb.NewTxManager(pg.Pool)

	a.useCases = &UseCases{
		OpenDirectConversation: usecase.NewOpenDirectConversation(conversations, participants, txManager, acl, profile, pet),
		ListConversations:      usecase.NewListConversations(conversations, profile, pet),
		GetConversation:        usecase.NewGetConversation(conversations, participants, acl, profile, pet),
		GetUnreadSummary:       usecase.NewGetUnreadSummary(conversations),
		GetMessageHistory:      usecase.NewGetMessageHistory(nil, nil, nil),
		SendMessage:            usecase.NewSendMessage(conversations, participants, nil, txManager, acl, nil),
		MarkRead:               usecase.NewMarkRead(participants, txManager, nil),
	}

	return nil
}
