package app

import (
	"email/internal/config"
	"email/internal/handler"
	"email/internal/queue"
	"email/internal/service"
	"email/internal/smtp"
	"email/internal/template"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type runtime struct {
	rabbitConn *amqp091.Connection
	rabbitCh   *amqp091.Channel
	consumer   *queue.Consumer
}

func buildRuntime(cfg *config.Config) (*runtime, error) {
	rabbitConn, rabbitCh, err := newRabbitChannel(cfg)
	if err != nil {
		return nil, err
	}

	renderer := template.NewRenderer(cfg.TemplateDir)
	primary := newSMTPProvider("primary", cfg.SMTPPrimary)
	fallback := newFallbackProvider(cfg.SMTPFallback)
	dispatcher := service.NewDispatcher(renderer, primary, fallback, cfg.RequeueOnFail)
	eventHandler := handler.NewEventHandler(dispatcher)

	return &runtime{
		rabbitConn: rabbitConn,
		rabbitCh:   rabbitCh,
		consumer:   queue.NewConsumer(rabbitCh, cfg.RabbitEmailQueue, eventHandler),
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

	if _, err := ch.QueueDeclare(cfg.RabbitEmailQueue, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, fmt.Errorf("queue declare %s: %w", cfg.RabbitEmailQueue, err)
	}

	return conn, ch, nil
}

func newSMTPProvider(name string, cfg config.SMTPConfig) smtp.Provider {
	return smtp.NewSMTPProvider(name, smtp.Config{
		Host:           cfg.Host,
		Port:           cfg.Port,
		Username:       cfg.Username,
		Password:       cfg.Password,
		From:           cfg.From,
		UseTLS:         cfg.UseTLS,
		UseStartTLS:    cfg.UseStartTLS,
		SkipTLSVerify:  cfg.SkipTLSVerify,
		ConnectTimeout: cfg.ConnectTimeout,
		SendTimeout:    cfg.SendTimeout,
	})
}

func newFallbackProvider(cfg config.SMTPConfig) smtp.Provider {
	if cfg.Host == "" {
		return nil
	}
	return newSMTPProvider("fallback", cfg)
}
