package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"email/internal/model"
	"email/internal/smtp"
	"email/internal/template"

	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type Dispatcher struct {
	renderer *template.Renderer

	primary  smtp.Provider
	fallback smtp.Provider

	requeueOnFail bool
}

func (d *Dispatcher) RequeueOnFail() bool {
	return d.requeueOnFail
}

func NewDispatcher(renderer *template.Renderer, primary smtp.Provider, fallback smtp.Provider, requeueOnFail bool) *Dispatcher {
	return &Dispatcher{
		renderer:      renderer,
		primary:       primary,
		fallback:      fallback,
		requeueOnFail: requeueOnFail,
	}
}

func (d *Dispatcher) Handle(ctx context.Context, msg amqp091.Delivery) {
	var job model.EmailJob
	if err := json.Unmarshal(msg.Body, &job); err != nil {
		log.Warn().Err(err).Msg("invalid email job json")
		_ = msg.Ack(false)
		return
	}

	if err := d.Process(ctx, job); err != nil {
		d.fail(msg)
		return
	}

	_ = msg.Ack(false)
}

func (d *Dispatcher) Process(ctx context.Context, job model.EmailJob) error {
	if err := validateJob(job); err != nil {
		log.Warn().Err(err).Msg("invalid email job payload")
		return nil
	}

	html, err := d.renderer.Render(job.Locale, job.Template, job.Data)
	if err != nil {
		log.Error().Err(err).Str("template", job.Template).Str("locale", job.Locale).Msg("template render failed")
		return errors.New("template render failed")
	}

	m := smtp.Message{
		To:      job.Email,
		Subject: job.Subject,
		HTML:    html,
	}

	if err := d.sendWithRetry(ctx, d.primary, m); err == nil {
		return nil
	}

	if d.fallback != nil {
		if err := d.sendWithRetry(ctx, d.fallback, m); err == nil {
			return nil
		}
	}

	log.Error().
		Str("to", job.Email).
		Str("template", job.Template).
		Msg("email send failed after retries and fallback")

	return errors.New("email send failed")
}

func (d *Dispatcher) sendWithRetry(ctx context.Context, p smtp.Provider, m smtp.Message) error {
	var lastErr error

	for i, delay := range BackoffSchedule {
		if err := p.Send(ctx, m); err == nil {
			log.Info().Str("provider", p.Name()).Str("to", m.To).Msg("email sent")
			return nil
		} else {
			lastErr = err
			log.Warn().
				Str("provider", p.Name()).
				Str("to", m.To).
				Int("attempt", i+1).
				Err(err).
				Msg("email send attempt failed")
		}

		if i < len(BackoffSchedule)-1 {
			select {
			case <-ctx.Done():
				return errors.New("context cancelled")
			case <-time.After(delay):
			}
		}
	}

	return lastErr
}

func (d *Dispatcher) fail(msg amqp091.Delivery) {
	if d.requeueOnFail {
		_ = msg.Nack(false, true)
		return
	}
	_ = msg.Nack(false, false)
}

func validateJob(j model.EmailJob) error {
	if j.Email == "" {
		return errors.New("empty email")
	}
	if j.Template == "" {
		return errors.New("empty template")
	}
	if j.Locale == "" {
		j.Locale = "ru"
	}
	if j.Subject == "" {
		j.Subject = "Notification"
	}
	if j.Data == nil {
		j.Data = map[string]any{}
	}
	return nil
}
