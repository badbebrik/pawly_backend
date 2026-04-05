package app

import (
	"chat/internal/application/usecase"
	aclclient "chat/internal/infrastructure/aclclient"
	chatdb "chat/internal/infrastructure/db"
	petclient "chat/internal/infrastructure/petclient"
	profileclient "chat/internal/infrastructure/profileclient"
	"chat/internal/infrastructure/realtime"
	"chat/internal/infrastructure/repository"
	"context"

	"github.com/redis/go-redis/v9"
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
	messages := repository.NewMessageRepository(pg.Pool)
	participants := repository.NewParticipantRepository(pg.Pool)
	txManager := chatdb.NewTxManager(pg.Pool)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     a.cfg.RedisAddr,
		Password: a.cfg.RedisPassword,
		DB:       a.cfg.RedisDB,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		pet.Close()
		profile.Close()
		acl.Close()
		pg.Close()
		return err
	}
	a.redis = redisClient
	a.presence = realtime.NewRedisPresenceTracker(redisClient, "chat:presence")

	a.useCases = &UseCases{
		OpenDirectConversation: usecase.NewOpenDirectConversation(conversations, participants, txManager, acl, profile, pet, a.presence),
		ListConversations:      usecase.NewListConversations(conversations, acl, profile, pet),
		GetConversation:        usecase.NewGetConversation(conversations, participants, acl, profile, pet, a.presence),
		GetUnreadSummary:       usecase.NewGetUnreadSummary(conversations),
		GetMessageHistory:      usecase.NewGetMessageHistory(conversations, participants, messages),
		SendMessage:            usecase.NewSendMessage(conversations, participants, messages, txManager, acl, nil),
		MarkRead:               usecase.NewMarkRead(conversations, participants, messages, txManager, nil),
	}

	a.rtPub = realtime.NewRedisPublisher(redisClient, a.cfg.RedisChannel)
	a.rtSub = realtime.NewRedisSubscriber(redisClient, a.cfg.RedisChannel, a.hub, a.useCases.GetConversation, a.useCases.GetUnreadSummary)

	return nil
}
