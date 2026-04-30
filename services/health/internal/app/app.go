package app

import (
	"context"
	"errors"
	healthuc "health/internal/application/usecase"
	"health/internal/config"
	"health/internal/infrastructure"
	"health/internal/infrastructure/aclclient"
	"health/internal/infrastructure/aclinternalclient"
	"health/internal/infrastructure/fileclient"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type App struct {
	cfg *config.Config

	logs           *healthuc.Logs
	scheduled      *healthuc.Scheduled
	vetVisits      *healthuc.VetVisits
	vaccinations   *healthuc.Vaccinations
	procedures     *healthuc.Procedures
	medicalRecords *healthuc.MedicalRecords
	analytics      *healthuc.Analytics
	documents      *healthuc.Documents
	overview       *healthuc.Overview
	dispatcher     *scheduledDispatcher

	pg          *infrastructure.Postgres
	acl         *aclclient.Client
	aclInternal *aclinternalclient.Client
	file        *fileclient.Client

	rabbitConn *amqp091.Connection
	rabbitCh   *amqp091.Channel

	ctx    context.Context
	cancel context.CancelFunc

	httpSrv *http.Server
}

func New(cfg *config.Config) (*App, error) {
	rt, err := buildRuntime(cfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	a := &App{
		cfg:            cfg,
		logs:           rt.logs,
		scheduled:      rt.scheduled,
		vetVisits:      rt.vetVisits,
		vaccinations:   rt.vaccinations,
		procedures:     rt.procedures,
		medicalRecords: rt.medicalRecords,
		analytics:      rt.analytics,
		documents:      rt.documents,
		overview:       rt.overview,
		dispatcher:     rt.dispatcher,
		pg:             rt.pg,
		acl:            rt.acl,
		aclInternal:    rt.aclInternal,
		file:           rt.file,
		rabbitConn:     rt.rabbitConn,
		rabbitCh:       rt.rabbitCh,
		ctx:            ctx,
		cancel:         cancel,
	}

	a.httpSrv = &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: a.setupRoutes(),
	}

	return a, nil
}

func (a *App) Close() {
	if a.cancel != nil {
		a.cancel()
	}
	if a.acl != nil {
		a.acl.Close()
	}
	if a.file != nil {
		a.file.Close()
	}
	if a.rabbitCh != nil {
		_ = a.rabbitCh.Close()
	}
	if a.rabbitConn != nil {
		_ = a.rabbitConn.Close()
	}
	if a.pg != nil {
		a.pg.Close()
	}
}

func (a *App) Run() error {
	go a.dispatcher.Run(
		a.ctx,
		time.Duration(a.cfg.ScheduledDispatchIntervalSec)*time.Second,
		a.cfg.ScheduledDispatchBatchSize,
		time.Duration(a.cfg.ScheduledHorizonIntervalSec)*time.Second,
		a.cfg.ScheduledHorizonBatchSize,
	)

	go func() {
		log.Info().Str("port", a.cfg.AppPort).Msg("starting Health HTTP server")
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("health http server crash")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down Health service...")
	if a.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := a.httpSrv.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			cancel()
			return err
		}
		cancel()
	}

	return nil
}
