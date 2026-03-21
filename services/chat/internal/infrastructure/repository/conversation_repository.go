package repository

import (
	"chat/internal/application/ports"
	"chat/internal/domain/model"
	chatdb "chat/internal/infrastructure/db"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConversationRepository struct {
	db *pgxpool.Pool
}

func NewConversationRepository(db *pgxpool.Pool) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) Create(ctx context.Context, conversation *model.Conversation) error {
	const query = `
		INSERT INTO conversations (
			conversation_id,
			pet_id,
			user_low_id,
			user_high_id,
			last_message_id,
			last_message_at,
			last_message_preview,
			last_message_sender_id,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := chatdb.GetExecutor(ctx, r.db).Exec(
		ctx,
		query,
		conversation.ID,
		conversation.PetID,
		conversation.UserLowID,
		conversation.UserHighID,
		conversation.LastMessageID,
		conversation.LastMessageAt,
		conversation.LastMessagePreview,
		conversation.LastMessageSenderID,
		conversation.CreatedAt,
		conversation.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ports.ErrConflict
		}
		return err
	}

	return nil
}

func (r *ConversationRepository) GetByID(ctx context.Context, conversationID uuid.UUID) (*model.Conversation, error) {
	const query = `
		SELECT
			conversation_id,
			pet_id,
			user_low_id,
			user_high_id,
			last_message_id,
			last_message_at,
			last_message_preview,
			last_message_sender_id,
			created_at,
			updated_at
		FROM conversations
		WHERE conversation_id = $1
	`

	row := chatdb.GetExecutor(ctx, r.db).QueryRow(ctx, query, conversationID)
	conversation, err := scanConversation(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}

	return conversation, nil
}

func (r *ConversationRepository) GetDirectByPetAndUsers(ctx context.Context, petID, userLowID, userHighID uuid.UUID) (*model.Conversation, error) {
	const query = `
		SELECT
			conversation_id,
			pet_id,
			user_low_id,
			user_high_id,
			last_message_id,
			last_message_at,
			last_message_preview,
			last_message_sender_id,
			created_at,
			updated_at
		FROM conversations
		WHERE pet_id = $1
		  AND user_low_id = $2
		  AND user_high_id = $3
	`

	row := chatdb.GetExecutor(ctx, r.db).QueryRow(ctx, query, petID, userLowID, userHighID)
	conversation, err := scanConversation(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}

	return conversation, nil
}

func (r *ConversationRepository) UpdateLastMessage(
	ctx context.Context,
	conversationID, messageID, senderUserID uuid.UUID,
	preview *string,
	createdAt time.Time,
) error {
	const query = `
		UPDATE conversations
		SET
			last_message_id = $2,
			last_message_at = $3,
			last_message_preview = $4,
			last_message_sender_id = $5,
			updated_at = $3
		WHERE conversation_id = $1
	`

	cmd, err := chatdb.GetExecutor(ctx, r.db).Exec(ctx, query, conversationID, messageID, createdAt, preview, senderUserID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ports.ErrNotFound
	}

	return nil
}

type conversationScanner interface {
	Scan(dest ...any) error
}

func scanConversation(s conversationScanner) (*model.Conversation, error) {
	var conversation model.Conversation
	if err := s.Scan(
		&conversation.ID,
		&conversation.PetID,
		&conversation.UserLowID,
		&conversation.UserHighID,
		&conversation.LastMessageID,
		&conversation.LastMessageAt,
		&conversation.LastMessagePreview,
		&conversation.LastMessageSenderID,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &conversation, nil
}

var _ ports.ConversationRepository = (*ConversationRepository)(nil)
