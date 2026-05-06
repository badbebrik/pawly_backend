package handler

import (
	"context"
	"encoding/json"
	"errors"
	"push/internal/application/ports"
	pushuc "push/internal/application/usecase"
	"push/internal/domain/model"
	"push/internal/sender"
	"testing"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type handlerPushRepo struct {
	tokens        []model.DeviceToken
	settings      *model.PetPushSettings
	settingsErr   error
	listErr       error
	deletedDevice string
}

func (r *handlerPushRepo) UpsertDeviceToken(context.Context, ports.UpsertDeviceTokenParams) (*model.DeviceToken, error) {
	return nil, nil
}

func (r *handlerPushRepo) DeleteDeviceToken(_ context.Context, in ports.DeleteDeviceTokenParams) error {
	r.deletedDevice = in.DeviceID
	return nil
}

func (r *handlerPushRepo) ListDeviceTokensByUser(context.Context, uuid.UUID) ([]model.DeviceToken, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.tokens, nil
}

func (r *handlerPushRepo) GetPetPushSettings(context.Context, uuid.UUID, uuid.UUID) (*model.PetPushSettings, error) {
	if r.settingsErr != nil {
		return nil, r.settingsErr
	}
	if r.settings != nil {
		return r.settings, nil
	}
	return nil, ports.ErrNotFound
}

func (r *handlerPushRepo) UpsertPetPushSettings(context.Context, ports.UpsertPetPushSettingsParams) (*model.PetPushSettings, error) {
	return nil, nil
}

type stubSender struct {
	sendFn func(context.Context, model.DeviceToken, model.PushMessage) error
	sent   []model.PushMessage
}

func (s *stubSender) Send(ctx context.Context, device model.DeviceToken, msg model.PushMessage) error {
	if s.sendFn != nil {
		return s.sendFn(ctx, device, msg)
	}
	s.sent = append(s.sent, msg)
	return nil
}

type ackRecorder struct {
	acked       bool
	nacked      bool
	rejected    bool
	nackRequeue bool
}

func (a *ackRecorder) Ack(uint64, bool) error {
	a.acked = true
	return nil
}

func (a *ackRecorder) Nack(_ uint64, _ bool, requeue bool) error {
	a.nacked = true
	a.nackRequeue = requeue
	return nil
}

func (a *ackRecorder) Reject(_ uint64, requeue bool) error {
	a.rejected = true
	a.nackRequeue = requeue
	return nil
}

func delivery(body []byte, ack *ackRecorder) amqp091.Delivery {
	return amqp091.Delivery{
		Acknowledger: ack,
		DeliveryTag:  1,
		Body:         body,
	}
}

func pushJobBody(t *testing.T, job model.ScheduledOccurrencePushJob) []byte {
	t.Helper()
	body, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("marshal job: %v", err)
	}
	return body
}

func TestPushJobHandlerNacksInvalidJSON(t *testing.T) {
	ack := &ackRecorder{}
	h := NewPushJobHandler(pushuc.New(&handlerPushRepo{}), &stubSender{})

	h.Handle(context.Background(), delivery([]byte("{"), ack))

	if !ack.nacked || ack.nackRequeue {
		t.Fatalf("expected nack without requeue, got %+v", ack)
	}
}

func TestPushJobHandlerAcksUnknownEvent(t *testing.T) {
	ack := &ackRecorder{}
	h := NewPushJobHandler(pushuc.New(&handlerPushRepo{}), &stubSender{})

	h.Handle(context.Background(), delivery(pushJobBody(t, model.ScheduledOccurrencePushJob{
		Event: "UNKNOWN",
	}), ack))

	if !ack.acked || ack.nacked {
		t.Fatalf("expected ack, got %+v", ack)
	}
}

func TestPushJobHandlerNacksInvalidPetID(t *testing.T) {
	ack := &ackRecorder{}
	h := NewPushJobHandler(pushuc.New(&handlerPushRepo{}), &stubSender{})

	h.Handle(context.Background(), delivery(pushJobBody(t, model.ScheduledOccurrencePushJob{
		Event: "SCHEDULED_OCCURRENCE_DUE",
		PetID: "bad",
	}), ack))

	if !ack.nacked || ack.nackRequeue {
		t.Fatalf("expected nack without requeue, got %+v", ack)
	}
}

func TestPushJobHandlerSendsEligibleTokensAndAcks(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	device := model.DeviceToken{ID: uuid.New(), UserID: userID, DeviceID: "phone", Platform: model.PlatformIOS, PushToken: "token"}
	repo := &handlerPushRepo{tokens: []model.DeviceToken{device}}
	sender := &stubSender{}
	ack := &ackRecorder{}
	h := NewPushJobHandler(pushuc.New(repo), sender)

	h.Handle(context.Background(), delivery(pushJobBody(t, model.ScheduledOccurrencePushJob{
		Event:           "SCHEDULED_OCCURRENCE_DUE",
		PetID:           petID.String(),
		OccurrenceID:    uuid.New().String(),
		ScheduledItemID: uuid.New().String(),
		UserIDs:         []string{"bad-user", userID.String()},
		SourceType:      "MANUAL",
		Title:           " Walk ",
		Note:            "Take a walk",
		ScheduledFor:    "2026-05-14T10:00:00Z",
	}), ack))

	if !ack.acked || ack.nacked {
		t.Fatalf("expected ack, got %+v", ack)
	}
	if len(sender.sent) != 1 {
		t.Fatalf("expected one sent message, got %d", len(sender.sent))
	}
	msg := sender.sent[0]
	if msg.Title != "Walk" || msg.Body != "Take a walk" {
		t.Fatalf("unexpected push message: %+v", msg)
	}
	if msg.Data["type"] != "scheduled_occurrence" || msg.Data["pet_id"] != petID.String() {
		t.Fatalf("unexpected push data: %+v", msg.Data)
	}
}

func TestPushJobHandlerUsesDefaultBodyWhenNoteEmpty(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	device := model.DeviceToken{ID: uuid.New(), UserID: userID, DeviceID: "phone", Platform: model.PlatformIOS, PushToken: "token"}
	sender := &stubSender{}
	ack := &ackRecorder{}
	h := NewPushJobHandler(pushuc.New(&handlerPushRepo{tokens: []model.DeviceToken{device}}), sender)

	h.Handle(context.Background(), delivery(pushJobBody(t, model.ScheduledOccurrencePushJob{
		Event:   "SCHEDULED_OCCURRENCE_DUE",
		PetID:   petID.String(),
		UserIDs: []string{userID.String()},
		Title:   "Reminder",
	}), ack))

	if !ack.acked || len(sender.sent) != 1 {
		t.Fatalf("expected ack and one message, ack=%+v sent=%d", ack, len(sender.sent))
	}
	if sender.sent[0].Body == "" {
		t.Fatalf("expected default body, got %+v", sender.sent[0])
	}
}

func TestPushJobHandlerDeletesInvalidDeviceTokenAndContinues(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()
	device := model.DeviceToken{ID: uuid.New(), UserID: userID, DeviceID: "phone", Platform: model.PlatformIOS, PushToken: "token"}
	repo := &handlerPushRepo{tokens: []model.DeviceToken{device}}
	ack := &ackRecorder{}
	h := NewPushJobHandler(pushuc.New(repo), &stubSender{
		sendFn: func(context.Context, model.DeviceToken, model.PushMessage) error {
			return sender.ErrInvalidDeviceToken
		},
	})

	h.Handle(context.Background(), delivery(pushJobBody(t, model.ScheduledOccurrencePushJob{
		Event:   "SCHEDULED_OCCURRENCE_DUE",
		PetID:   petID.String(),
		UserIDs: []string{userID.String()},
	}), ack))

	if !ack.acked || ack.nacked {
		t.Fatalf("expected ack after deleting invalid token, got %+v", ack)
	}
	if repo.deletedDevice != "phone" {
		t.Fatalf("expected invalid device deletion, got %q", repo.deletedDevice)
	}
}

func TestPushJobHandlerNacksWithRequeueOnResolveOrSendError(t *testing.T) {
	userID := uuid.New()
	petID := uuid.New()

	t.Run("resolve tokens", func(t *testing.T) {
		ack := &ackRecorder{}
		h := NewPushJobHandler(pushuc.New(&handlerPushRepo{listErr: errors.New("repo error")}), &stubSender{})

		h.Handle(context.Background(), delivery(pushJobBody(t, model.ScheduledOccurrencePushJob{
			Event:   "SCHEDULED_OCCURRENCE_DUE",
			PetID:   petID.String(),
			UserIDs: []string{userID.String()},
		}), ack))

		if !ack.nacked || !ack.nackRequeue {
			t.Fatalf("expected nack with requeue, got %+v", ack)
		}
	})

	t.Run("send", func(t *testing.T) {
		device := model.DeviceToken{ID: uuid.New(), UserID: userID, DeviceID: "phone", Platform: model.PlatformIOS, PushToken: "token"}
		ack := &ackRecorder{}
		h := NewPushJobHandler(pushuc.New(&handlerPushRepo{tokens: []model.DeviceToken{device}}), &stubSender{
			sendFn: func(context.Context, model.DeviceToken, model.PushMessage) error {
				return errors.New("send error")
			},
		})

		h.Handle(context.Background(), delivery(pushJobBody(t, model.ScheduledOccurrencePushJob{
			Event:   "SCHEDULED_OCCURRENCE_DUE",
			PetID:   petID.String(),
			UserIDs: []string{userID.String()},
		}), ack))

		if !ack.nacked || !ack.nackRequeue {
			t.Fatalf("expected nack with requeue, got %+v", ack)
		}
	})
}

var _ ports.PushRepository = (*handlerPushRepo)(nil)
var _ sender.Sender = (*stubSender)(nil)
var _ amqp091.Acknowledger = (*ackRecorder)(nil)
