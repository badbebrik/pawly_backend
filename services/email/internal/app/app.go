package app

import (
	"context"
	"email/config"
	"email/internal/queue"
	"email/internal/service"
	"email/internal/smtp"
	"email/internal/template"
	"errors"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	config *config.Config

	ctx    context.Context
	cancel context.CancelFunc

	rabbitConn *amqp091.Connection
	rabbitCh   *amqp091.Channel

	consumer *queue.Consumer
	httpSrv  *http.Server
}

func New(cfg *config.Config) (*App, error) {
	conn, err := amqp091.Dial(fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		cfg.Rabbit.User,
		cfg.Rabbit.Password,
		cfg.Rabbit.Host,
		cfg.Rabbit.Port,
	))
	if err != nil {
		return nil, fmt.Errorf("rabbit connect: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbit channel: %w", err)
	}

	if _, err := ch.QueueDeclare(cfg.Rabbit.Queue, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("queue declare %s: %w", cfg.Rabbit.Queue, err)
	}

	renderer := template.NewRenderer(cfg.TemplateDir)

	primary := smtp.NewSMTPProvider("primary", smtp.Config{
		Host:           cfg.SMTPPrimary.Host,
		Port:           cfg.SMTPPrimary.Port,
		Username:       cfg.SMTPPrimary.Username,
		Password:       cfg.SMTPPrimary.Password,
		From:           cfg.SMTPPrimary.From,
		UseTLS:         cfg.SMTPPrimary.UseTLS,
		UseStartTLS:    cfg.SMTPPrimary.UseStartTLS,
		SkipTLSVerify:  cfg.SMTPPrimary.SkipTLSVerify,
		ConnectTimeout: cfg.SMTPPrimary.ConnectTimeout,
		SendTimeout:    cfg.SMTPPrimary.SendTimeout,
	})

	var fallback smtp.Provider
	if cfg.SMTPFallback.Host != "" {
		fallback = smtp.NewSMTPProvider("fallback", smtp.Config{
			Host:           cfg.SMTPFallback.Host,
			Port:           cfg.SMTPFallback.Port,
			Username:       cfg.SMTPFallback.Username,
			Password:       cfg.SMTPFallback.Password,
			From:           cfg.SMTPFallback.From,
			UseTLS:         cfg.SMTPFallback.UseTLS,
			UseStartTLS:    cfg.SMTPFallback.UseStartTLS,
			SkipTLSVerify:  cfg.SMTPFallback.SkipTLSVerify,
			ConnectTimeout: cfg.SMTPFallback.ConnectTimeout,
			SendTimeout:    cfg.SMTPFallback.SendTimeout,
		})
	}

	dispatcher := service.NewDispatcher(renderer, primary, fallback, cfg.RequeueOnFail)
	consumer := queue.NewConsumer(ch, cfg.Rabbit.Queue, dispatcher)

	ctx, cancel := context.WithCancel(context.Background())

	return &App{
		config:     cfg,
		ctx:        ctx,
		cancel:     cancel,
		rabbitConn: conn,
		rabbitCh:   ch,
		consumer:   consumer,
	}, nil
}

func (a *App) Close() {
	log.Info().Msg("closing Email App resources...")

	if a.cancel != nil {
		a.cancel()
	}

	if a.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = a.httpSrv.Shutdown(ctx)
		cancel()
	}

	if a.rabbitCh != nil {
		_ = a.rabbitCh.Close()
	}
	if a.rabbitConn != nil {
		_ = a.rabbitConn.Close()
	}
}

func (a *App) Run() error {
	if err := a.consumer.Start(a.ctx); err != nil {
		return err
	}
	log.Info().Str("queue", a.config.Rabbit.Queue).Msg("started email.jobs consumer")

	r := a.setupRoutes()
	a.httpSrv = &http.Server{
		Addr:    ":" + a.config.AppPort,
		Handler: r,
	}

	go func() {
		log.Info().Str("port", a.config.AppPort).Msg("starting HTTP server")
		if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("http server crash")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down Email service...")
	return nil
}
