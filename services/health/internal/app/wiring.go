package app

import (
	"fmt"
	healthuc "health/internal/application/usecase"
	"health/internal/config"
	"health/internal/infrastructure"
	"health/internal/infrastructure/aclclient"
	"health/internal/infrastructure/aclinternalclient"
	"health/internal/infrastructure/fileclient"
	pgrepo "health/internal/infrastructure/repository"
	"health/internal/queue"

	"github.com/rabbitmq/amqp091-go"
)

type runtime struct {
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
	pg             *infrastructure.Postgres
	acl            *aclclient.Client
	aclInternal    *aclinternalclient.Client
	file           *fileclient.Client
	rabbitConn     *amqp091.Connection
	rabbitCh       *amqp091.Channel
}

func buildRuntime(cfg *config.Config) (*runtime, error) {
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

	rabbitConn, rabbitCh, err := newRabbitChannel(cfg)
	if err != nil {
		file.Close()
		acl.Close()
		pg.Close()
		return nil, err
	}

	logRepo := pgrepo.NewLogRepository(pg.Pool)
	scheduledRepo := pgrepo.NewScheduledRepository(pg.Pool)
	pushPublisher := queue.NewPushPublisher(rabbitCh, cfg.RabbitPushQueue)
	logs := healthuc.NewLogs(logRepo, acl, file)
	analytics := healthuc.NewAnalytics(logRepo, acl)
	scheduled := healthuc.NewScheduled(scheduledRepo, acl)
	vetVisits := healthuc.NewVetVisits(logRepo, acl, file, scheduledRepo)
	vaccinations := healthuc.NewVaccinations(logRepo, logRepo, logRepo, acl, file, scheduledRepo)
	procedures := healthuc.NewProcedures(logRepo, logRepo, logRepo, acl, file, scheduledRepo)
	medicalRecords := healthuc.NewMedicalRecords(logRepo, logRepo, acl, file)
	documents := healthuc.NewDocuments(logRepo, acl, file)
	overview := healthuc.NewOverview(acl, acl, scheduledRepo, logRepo, logRepo)
	dispatcher := newScheduledDispatcher(healthuc.NewScheduledDispatcher(healthuc.ScheduledDispatcherDependencies{
		Repository:    scheduledRepo,
		PetUserLister: aclInternal,
		PushPublisher: pushPublisher,
	}), scheduled)

	return &runtime{
		logs:           logs,
		scheduled:      scheduled,
		vetVisits:      vetVisits,
		vaccinations:   vaccinations,
		procedures:     procedures,
		medicalRecords: medicalRecords,
		analytics:      analytics,
		documents:      documents,
		overview:       overview,
		dispatcher:     dispatcher,
		pg:             pg,
		acl:            acl,
		aclInternal:    aclInternal,
		file:           file,
		rabbitConn:     rabbitConn,
		rabbitCh:       rabbitCh,
	}, nil
}

func newRabbitChannel(cfg *config.Config) (*amqp091.Connection, *amqp091.Channel, error) {
	conn, err := amqp091.Dial(fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		cfg.RabbitUser,
		cfg.RabbitPassword,
		cfg.RabbitHost,
		cfg.RabbitPort,
	))
	if err != nil {
		return nil, nil, fmt.Errorf("rabbit connect: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("rabbit channel: %w", err)
	}

	if _, err := ch.QueueDeclare(cfg.RabbitPushQueue, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, fmt.Errorf("queue declare %s: %w", cfg.RabbitPushQueue, err)
	}

	return conn, ch, nil
}
