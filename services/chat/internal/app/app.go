package app

import (
	"chat/internal/application/usecase"
	"chat/internal/config"
	aclclient "chat/internal/infrastructure/aclclient"
	chatdb "chat/internal/infrastructure/db"
	petclient "chat/internal/infrastructure/petclient"
	profileclient "chat/internal/infrastructure/profileclient"
	"chat/internal/infrastructure/realtime"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
)

type App struct {
	cfg      *config.Config
	useCases *UseCases
	httpSrv  *http.Server
	hub      *realtime.Hub
	redis    *redis.Client
	rtPub    realtime.EventPublisher
	rtSub    *realtime.RedisSubscriber
	presence realtime.PresenceTracker
	pg       *chatdb.Postgres
	acl      *aclclient.Client
	profile  *profileclient.Client
	pet      *petclient.Client
	stopBg   func()
}

func New(cfg *config.Config) (*App, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	a := &App{
		cfg: cfg,
		hub: realtime.NewHub(),
	}

	if err := a.wire(); err != nil {
		return nil, err
	}

	a.httpSrv = &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: a.setupRoutes(),
	}

	return a, nil
}

func (a *App) Run() error {
	if a.rtSub != nil {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		a.stopBg = cancel
		go func() {
			_ = a.rtSub.Run(ctx)
		}()
	} else {
		quitCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		a.stopBg = cancel
		_ = quitCtx
	}

	go func() {
		_ = a.httpSrv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	return nil
}

func (a *App) Close() {
	if a.stopBg != nil {
		a.stopBg()
	}
	if a.httpSrv != nil {
		_ = a.httpSrv.Close()
	}
	if a.redis != nil {
		_ = a.redis.Close()
	}
	if a.pg != nil {
		a.pg.Close()
	}
	if a.acl != nil {
		a.acl.Close()
	}
	if a.profile != nil {
		a.profile.Close()
	}
	if a.pet != nil {
		a.pet.Close()
	}
}

type UseCases struct {
	OpenDirectConversation *usecase.OpenDirectConversation
	ListConversations      *usecase.ListConversations
	GetConversation        *usecase.GetConversation
	GetUnreadSummary       *usecase.GetUnreadSummary
	GetMessageHistory      *usecase.GetMessageHistory
	SendMessage            *usecase.SendMessage
	MarkRead               *usecase.MarkRead
}
