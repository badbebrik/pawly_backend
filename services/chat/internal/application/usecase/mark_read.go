package usecase

import (
	"bytes"
	"chat/internal/application/ports"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type MarkRead struct {
	conversations ports.ConversationRepository
	participants  ports.ParticipantRepository
	messages      ports.MessageRepository
	tx            ports.TxManager
	realtime      ports.RealtimePublisher
}

func NewMarkRead(
	conversations ports.ConversationRepository,
	participants ports.ParticipantRepository,
	messages ports.MessageRepository,
	tx ports.TxManager,
	realtime ports.RealtimePublisher,
) *MarkRead {
	return &MarkRead{
		conversations: conversations,
		participants:  participants,
		messages:      messages,
		tx:            tx,
		realtime:      realtime,
	}
}

type MarkReadParams struct {
	CurrentUserID     uuid.UUID
	ConversationID    uuid.UUID
	LastReadMessageID uuid.UUID
}

type MarkReadResult struct {
	ConversationID    uuid.UUID
	LastReadMessageID uuid.UUID
}

func (uc *MarkRead) Execute(ctx context.Context, params MarkReadParams) (MarkReadResult, error) {
	if params.CurrentUserID == uuid.Nil || params.ConversationID == uuid.Nil || params.LastReadMessageID == uuid.Nil {
		return MarkReadResult{}, ErrInvalidInput
	}

	var result MarkReadResult
	changed := false
	err := uc.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		conversation, err := uc.conversations.GetByID(txCtx, params.ConversationID)
		if err != nil {
			return err
		}
		if !conversation.HasParticipant(params.CurrentUserID) {
			return ErrForbidden
		}

		participant, err := uc.participants.GetByConversationAndUser(txCtx, params.ConversationID, params.CurrentUserID)
		if err != nil {
			if errors.Is(err, ports.ErrNotFound) {
				return ErrForbidden
			}
			return err
		}

		targetMessage, err := uc.messages.GetByID(txCtx, params.LastReadMessageID)
		if err != nil {
			if errors.Is(err, ports.ErrNotFound) {
				return ErrInvalidInput
			}
			return err
		}
		if targetMessage.ConversationID != params.ConversationID {
			return ErrInvalidInput
		}

		if participant.LastReadMessageID != nil {
			currentMessage, err := uc.messages.GetByID(txCtx, *participant.LastReadMessageID)
			if err != nil {
				return err
			}
			if !messagePositionLess(currentMessage.CreatedAt, currentMessage.ID, targetMessage.CreatedAt, targetMessage.ID) {
				result = MarkReadResult{
					ConversationID:    params.ConversationID,
					LastReadMessageID: *participant.LastReadMessageID,
				}
				return nil
			}
		}

		if _, err := uc.participants.MarkRead(
			txCtx,
			params.ConversationID,
			params.CurrentUserID,
			params.LastReadMessageID,
			time.Now().UTC(),
		); err != nil {
			return err
		}
		changed = true

		result = MarkReadResult{
			ConversationID:    params.ConversationID,
			LastReadMessageID: params.LastReadMessageID,
		}
		return nil
	})
	if err != nil {
		return MarkReadResult{}, err
	}

	if changed && uc.realtime != nil {
		_ = uc.realtime.PublishReadUpdated(ctx, ports.ReadUpdatedEvent{
			ConversationID:    result.ConversationID,
			UserID:            params.CurrentUserID,
			LastReadMessageID: result.LastReadMessageID,
		})
	}

	return result, nil
}

func messagePositionLess(leftAt time.Time, leftID uuid.UUID, rightAt time.Time, rightID uuid.UUID) bool {
	if leftAt.Before(rightAt) {
		return true
	}
	if leftAt.After(rightAt) {
		return false
	}

	return bytes.Compare(leftID[:], rightID[:]) < 0
}
