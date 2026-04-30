package app

import (
	"chat/internal/application/usecase"
	"chat/internal/config"
	aclclient "chat/internal/infrastructure/aclclient"
	chatdb "chat/internal/infrastructure/db"
	petclient "chat/internal/infrastructure/petclient"
	profileclient "chat/internal/infrastructure/profileclient"
	rtinfra "chat/internal/infrastructure/realtime"
	"chat/internal/realtime"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	cfg      *config.Config
	useCases *usecase.Set
	httpSrv  *http.Server
	hub      *realtime.Hub
	redis    *redis.Client
	rtPub    rtinfra.EventPublisher
	rtSub    *rtinfra.RedisSubscriber
	presence realtime.PresenceTracker
	pg       *chatdb.Postgres
	acl      *aclclient.Client
	profile  *profileclient.Client
	pet      *petclient.Client
}

func New(cfg *config.Config) (*App, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	hub := realtime.NewHub()
	runtime, err := buildRuntime(cfg, hub)
	if err != nil {
		return nil, err
	}

	app := &App{
		cfg:      cfg,
		useCases: runtime.useCases,
		hub:      hub,
		redis:    runtime.redis,
		rtPub:    runtime.rtPub,
		rtSub:    runtime.rtSub,
		presence: runtime.presence,
		pg:       runtime.pg,
		acl:      runtime.acl,
		profile:  runtime.profile,
		pet:      runtime.pet,
	}

	app.httpSrv = &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: app.setupRoutes(),
	}

	return app, nil
}

func (a *App) Close() {
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

func (a *App) Run() error {
	runCtx, cancelRun := context.WithCancel(context.Background())
	defer cancelRun()

	if a.rtSub != nil {
		go func() {
			_ = a.rtSub.Run(runCtx)
		}()
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case <-quit:
	case err := <-serverErr:
		cancelRun()
		return err
	}

	cancelRun()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	return a.httpSrv.Shutdown(shutdownCtx)
}
