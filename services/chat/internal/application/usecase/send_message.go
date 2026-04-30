package usecase

import (
	"chat/internal/application/ports"
	"chat/internal/domain/model"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type SendMessage struct {
	conversations ports.ConversationRepository
	participants  ports.ParticipantRepository
	messages      ports.MessageRepository
	tx            ports.TxManager
	acl           ports.ACLClient
	realtime      ports.RealtimePublisher
}

func NewSendMessage(
	conversations ports.ConversationRepository,
	participants ports.ParticipantRepository,
	messages ports.MessageRepository,
	tx ports.TxManager,
	acl ports.ACLClient,
	realtime ports.RealtimePublisher,
) *SendMessage {
	return &SendMessage{
		conversations: conversations,
		participants:  participants,
		messages:      messages,
		tx:            tx,
		acl:           acl,
		realtime:      realtime,
	}
}

type SendMessageParams struct {
	CurrentUserID  uuid.UUID
	OriginClientID uuid.UUID
	ConversationID uuid.UUID
	ClientMsgID    uuid.UUID
	Text           *string
}

type SendMessageMessage struct {
	MessageID      uuid.UUID
	ConversationID uuid.UUID
	SenderUserID   uuid.UUID
	ClientMsgID    uuid.UUID
	Text           *string
	CreatedAt      time.Time
}

type SendMessageResult struct {
	Message SendMessageMessage
}

func (uc *SendMessage) Execute(ctx context.Context, params SendMessageParams) (SendMessageResult, error) {
	if params.CurrentUserID == uuid.Nil || params.ConversationID == uuid.Nil {
		return SendMessageResult{}, ErrInvalidInput
	}

	conversation, err := uc.conversations.GetByID(ctx, params.ConversationID)
	if err != nil {
		return SendMessageResult{}, err
	}
	if !conversation.HasParticipant(params.CurrentUserID) {
		return SendMessageResult{}, ErrForbidden
	}

	otherUserID := conversation.OtherUserID(params.CurrentUserID)
	canSend, err := uc.canSendMessage(ctx, conversation.PetID, params.CurrentUserID, otherUserID)
	if err != nil {
		return SendMessageResult{}, err
	}
	if !canSend {
		return SendMessageResult{}, ErrForbidden
	}

	existingMessage, err := uc.messages.FindByClientMsgID(ctx, params.ConversationID, params.CurrentUserID, params.ClientMsgID)
	if err == nil {
		return SendMessageResult{Message: mapSendMessageResult(existingMessage)}, nil
	}
	if !errors.Is(err, ports.ErrNotFound) {
		return SendMessageResult{}, err
	}

	message := &model.Message{
		ID:             uuid.New(),
		ConversationID: params.ConversationID,
		SenderUserID:   params.CurrentUserID,
		ClientMsgID:    params.ClientMsgID,
		Text:           params.Text,
		CreatedAt:      time.Now().UTC(),
	}
	if err := message.ValidateForSend(); err != nil {
		return SendMessageResult{}, ErrInvalidInput
	}

	created := false
	err = uc.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		existingMessage, err := uc.messages.FindByClientMsgID(txCtx, params.ConversationID, params.CurrentUserID, params.ClientMsgID)
		if err == nil {
			*message = *existingMessage
			return nil
		}
		if !errors.Is(err, ports.ErrNotFound) {
			return err
		}

		if err := uc.messages.Create(txCtx, message); err != nil {
			if errors.Is(err, ports.ErrConflict) {
				existingMessage, findErr := uc.messages.FindByClientMsgID(txCtx, params.ConversationID, params.CurrentUserID, params.ClientMsgID)
				if findErr != nil {
					return findErr
				}
				*message = *existingMessage
				return nil
			}
			return err
		}
		created = true

		if err := uc.conversations.UpdateLastMessage(
			txCtx,
			params.ConversationID,
			message.ID,
			params.CurrentUserID,
			message.Text,
			message.CreatedAt,
		); err != nil {
			return err
		}

		if err := uc.participants.IncrementUnread(txCtx, params.ConversationID, otherUserID, 1); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return SendMessageResult{}, err
	}

	if created && uc.realtime != nil {
		_ = uc.realtime.PublishMessageSent(ctx, ports.MessageSentEvent{
			OriginClientID:  params.OriginClientID,
			ConversationID:  message.ConversationID,
			MessageID:       message.ID,
			SenderUserID:    message.SenderUserID,
			RecipientUserID: otherUserID,
			ClientMsgID:     message.ClientMsgID,
			Text:            message.Text,
			CreatedAt:       message.CreatedAt,
		})
	}

	return SendMessageResult{
		Message: mapSendMessageResult(message),
	}, nil
}

func (uc *SendMessage) canSendMessage(ctx context.Context, petID, currentUserID, otherUserID uuid.UUID) (bool, error) {
	currentUserActive, err := uc.acl.IsActiveMember(ctx, petID, currentUserID)
	if err != nil {
		return false, err
	}
	if !currentUserActive {
		return false, nil
	}

	otherUserActive, err := uc.acl.IsActiveMember(ctx, petID, otherUserID)
	if err != nil {
		return false, err
	}

	return otherUserActive, nil
}

func mapSendMessageResult(message *model.Message) SendMessageMessage {
	return SendMessageMessage{
		MessageID:      message.ID,
		ConversationID: message.ConversationID,
		SenderUserID:   message.SenderUserID,
		ClientMsgID:    message.ClientMsgID,
		Text:           message.Text,
		CreatedAt:      message.CreatedAt,
	}
}
