package sender

import (
	"context"

	"push/internal/model"
)

type Sender interface {
	Send(ctx context.Context, device model.DeviceToken, msg model.PushMessage) error
}
