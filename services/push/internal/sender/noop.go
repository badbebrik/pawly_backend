package sender

import (
	"context"

	"push/internal/domain/model"

	"github.com/rs/zerolog/log"
)

type NoopSender struct{}

func NewNoopSender() *NoopSender {
	return &NoopSender{}
}

func (s *NoopSender) Send(_ context.Context, device model.DeviceToken, msg model.PushMessage) error {
	log.Warn().
		Str("device_id", device.DeviceID).
		Str("platform", device.Platform).
		Str("title", msg.Title).
		Msg("push sender is not configured, skipping send")
	return nil
}
