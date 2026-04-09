package app

import (
	"context"
	"errors"
	"fmt"
	"health/internal/config"
	"health/internal/infrastructure"
	"health/internal/infrastructure/aclclient"
	"health/internal/infrastructure/aclinternalclient"
	"health/internal/infrastructure/fileclient"
	pgrepo "health/internal/infrastructure/repository"
	"health/internal/model"
	"health/internal/repository"
	"health/internal/queue"
	"health/internal/service"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type App struct {
	cfg *config.Config

	pg   *infrastructure.Postgres
	acl  *aclclient.Client
	aclInternal *aclinternalclient.Client
	file *fileclient.Client

	repo repository.Repository
	logSvc *service.Service

	rabbitConn    *amqp091.Connection
	rabbitCh      *amqp091.Channel
	pushPublisher *queue.PushPublisher

	ctx    context.Context
	cancel context.CancelFunc

	httpSrv *http.Server
}

func New(cfg *config.Config) (*App, error) {
	pg, err := infrastructure.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	acl, err := aclclient.New(cfg.ACLGRPCAddr)
	if err != nil {
		pg.Close()
		return nil, err
	}

	file, err := fileclient.New(cfg.FileGRPCAddr)
	if err != nil {
		acl.Close()
		pg.Close()
		return nil, err
	}

	aclInternal, err := aclinternalclient.New(cfg.ACLHTTPBaseURL, cfg.InternalServiceToken)
	if err != nil {
		file.Close()
		acl.Close()
		pg.Close()
		return nil, err
	}

	conn, err := amqp091.Dial(fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		cfg.RabbitUser,
		cfg.RabbitPassword,
		cfg.RabbitHost,
		cfg.RabbitPort,
	))
	if err != nil {
		file.Close()
		acl.Close()
		pg.Close()
		return nil, fmt.Errorf("rabbit connect: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		file.Close()
		acl.Close()
		pg.Close()
		return nil, fmt.Errorf("rabbit channel: %w", err)
	}
	if _, err := ch.QueueDeclare(cfg.RabbitPushJobsQueue, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		file.Close()
		acl.Close()
		pg.Close()
		return nil, fmt.Errorf("queue declare %s: %w", cfg.RabbitPushJobsQueue, err)
	}

	logRepo := pgrepo.NewRepository(pg.Pool)
	logSvc := service.New(logRepo, acl, file)
	ctx, cancel := context.WithCancel(context.Background())

	a := &App{
		cfg:          cfg,
		pg:           pg,
		acl:          acl,
		aclInternal:  aclInternal,
		file:         file,
		repo:         logRepo,
		logSvc:       logSvc,
		rabbitConn:   conn,
		rabbitCh:     ch,
		pushPublisher: queue.NewPushPublisher(ch, cfg.RabbitPushJobsQueue),
		ctx:          ctx,
		cancel:       cancel,
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
	if a.httpSrv != nil {
		_ = a.httpSrv.Close()
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
	go a.runScheduledDispatchWorker()

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
	return nil
}

func (a *App) runScheduledDispatchWorker() {
	intervalSec, err := strconv.Atoi(a.cfg.ScheduledDispatchIntervalSec)
	if err != nil || intervalSec <= 0 {
		intervalSec = 60
	}
	batchSize, err := strconv.Atoi(a.cfg.ScheduledDispatchBatchSize)
	if err != nil || batchSize <= 0 {
		batchSize = 100
	}

	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	defer ticker.Stop()

	a.dispatchDueScheduledOccurrences(batchSize)
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.dispatchDueScheduledOccurrences(batchSize)
		}
	}
}

func (a *App) dispatchDueScheduledOccurrences(batchSize int) {
	const dispatchKey = "scheduled_occurrence_due"

	items, err := a.repo.ListDueScheduledItemOccurrences(a.ctx, repository.ListDueScheduledItemOccurrencesInput{
		Before:      time.Now().UTC(),
		Limit:       batchSize,
		DispatchKey: dispatchKey,
	})
	if err != nil {
		log.Error().Err(err).Msg("list due scheduled occurrences failed")
		return
	}

	for i := range items {
		item := items[i]

		userIDs, err := a.aclInternal.ListPetUserIDs(a.ctx, item.PetID)
		if err != nil {
			log.Error().Err(err).Str("pet_id", item.PetID.String()).Str("occurrence_id", item.ID.String()).Msg("list pet users failed")
			continue
		}
		if len(userIDs) == 0 {
			log.Warn().Str("pet_id", item.PetID.String()).Str("occurrence_id", item.ID.String()).Msg("no pet users for due occurrence")
			continue
		}

		if err := a.repo.CreateScheduledItemDispatch(a.ctx, repository.CreateScheduledItemDispatchInput{
			ID:                        uuid.New(),
			ScheduledItemOccurrenceID: item.ID,
			DispatchKey:               dispatchKey,
		}); err != nil {
			if err == repository.ErrConflict {
				continue
			}
			log.Error().Err(err).Str("occurrence_id", item.ID.String()).Msg("create occurrence dispatch failed")
			continue
		}

		job := model.ScheduledOccurrencePushJob{
			Event:           "SCHEDULED_OCCURRENCE_DUE",
			PetID:           item.PetID.String(),
			OccurrenceID:    item.ID.String(),
			ScheduledItemID: item.ScheduledItemID.String(),
			UserIDs:         uuidStrings(userIDs),
			SourceType:      item.Rule.SourceType,
			Title:           item.Rule.Title,
			ScheduledFor:    item.ScheduledFor.UTC().Format(time.RFC3339),
		}
		if item.Rule.Note != nil {
			job.Note = *item.Rule.Note
		}

		if err := a.pushPublisher.PublishScheduledOccurrenceDue(a.ctx, job); err != nil {
			log.Error().Err(err).Str("occurrence_id", item.ID.String()).Msg("publish scheduled occurrence push job failed")
			continue
		}
	}
}

func uuidStrings(items []uuid.UUID) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item == uuid.Nil {
			continue
		}
		out = append(out, item.String())
	}
	return out
}
