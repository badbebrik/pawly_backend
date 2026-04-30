package app

import (
	"chat/internal/application/usecase"
	"chat/internal/config"
	aclclient "chat/internal/infrastructure/aclclient"
	chatdb "chat/internal/infrastructure/db"
	petclient "chat/internal/infrastructure/petclient"
	profileclient "chat/internal/infrastructure/profileclient"
	rtinfra "chat/internal/infrastructure/realtime"
	"chat/internal/infrastructure/repository"
	"chat/internal/realtime"
	"chat/internal/transport/ws"
	"context"

	"github.com/redis/go-redis/v9"
)

type runtime struct {
	useCases *usecase.Set
	redis    *redis.Client
	rtPub    rtinfra.EventPublisher
	rtSub    *rtinfra.RedisSubscriber
	presence realtime.PresenceTracker
	pg       *chatdb.Postgres
	acl      *aclclient.Client
	profile  *profileclient.Client
	pet      *petclient.Client
}

func buildRuntime(cfg *config.Config, hub *realtime.Hub) (*runtime, error) {
	pg, err := chatdb.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	acl, err := aclclient.New(cfg.ACLGRPCAddr)
	if err != nil {
		pg.Close()
		return nil, err
	}

	profile, err := profileclient.New(cfg.ProfileGRPCAddr)
	if err != nil {
		acl.Close()
		pg.Close()
		return nil, err
	}

	pet, err := petclient.New(cfg.PetGRPCAddr)
	if err != nil {
		profile.Close()
		acl.Close()
		pg.Close()
		return nil, err
	}

	conversations := repository.NewConversationRepository(pg.Pool)
	messages := repository.NewMessageRepository(pg.Pool)
	participants := repository.NewParticipantRepository(pg.Pool)
	txManager := chatdb.NewTxManager(pg.Pool)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		_ = redisClient.Close()
		pet.Close()
		profile.Close()
		acl.Close()
		pg.Close()
		return nil, err
	}

	presence := rtinfra.NewRedisPresenceTracker(redisClient, "chat:presence")
	rtPub := rtinfra.NewRedisPublisher(redisClient, cfg.RedisChannel)
	useCases := usecase.NewSet(usecase.Dependencies{
		Conversations: conversations,
		Participants:  participants,
		Messages:      messages,
		TxManager:     txManager,
		ACLClient:     acl,
		ProfileClient: profile,
		PetClient:     pet,
		Presence:      presence,
		Realtime:      rtPub,
	})
	subscriberHandler := ws.NewSubscriberHandler(hub, useCases.GetConversation, useCases.GetUnreadSummary)
	rtSub := rtinfra.NewRedisSubscriber(
		redisClient,
		cfg.RedisChannel,
		subscriberHandler,
	)

	return &runtime{
		useCases: useCases,
		redis:    redisClient,
		rtPub:    rtPub,
		rtSub:    rtSub,
		presence: presence,
		pg:       pg,
		acl:      acl,
		profile:  profile,
		pet:      pet,
	}, nil
}
