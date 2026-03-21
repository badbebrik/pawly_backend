package app

import (
	"chat/internal/application/usecase"
	"chat/internal/config"
	aclclient "chat/internal/infrastructure/aclclient"
	chatdb "chat/internal/infrastructure/db"
	petclient "chat/internal/infrastructure/petclient"
	profileclient "chat/internal/infrastructure/profileclient"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	cfg      *config.Config
	useCases *UseCases
	httpSrv  *http.Server
	pg       *chatdb.Postgres
	acl      *aclclient.Client
	profile  *profileclient.Client
	pet      *petclient.Client
}

func New(cfg *config.Config) (*App, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	a := &App{
		cfg: cfg,
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
	go func() {
		_ = a.httpSrv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	return nil
}

func (a *App) Close() {
	if a.httpSrv != nil {
		_ = a.httpSrv.Close()
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
