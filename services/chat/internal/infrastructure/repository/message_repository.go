package repository

import (
	"chat/internal/application/ports"
	"chat/internal/domain/model"
	chatdb "chat/internal/infrastructure/db"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageRepository struct {
	db *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, message *model.Message) error {
	const query = `
		INSERT INTO messages (
			message_id,
			conversation_id,
			sender_user_id,
			client_msg_id,
			text,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := chatdb.GetExecutor(ctx, r.db).Exec(
		ctx,
		query,
		message.ID,
		message.ConversationID,
		message.SenderUserID,
		message.ClientMsgID,
		message.Text,
		message.CreatedAt,
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

func (r *MessageRepository) GetByID(ctx context.Context, messageID uuid.UUID) (*model.Message, error) {
	const query = `
		SELECT
			message_id,
			conversation_id,
			sender_user_id,
			client_msg_id,
			text,
			created_at
		FROM messages
		WHERE message_id = $1
	`

	var message model.Message
	err := chatdb.GetExecutor(ctx, r.db).QueryRow(ctx, query, messageID).Scan(
		&message.ID,
		&message.ConversationID,
		&message.SenderUserID,
		&message.ClientMsgID,
		&message.Text,
		&message.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}

	return &message, nil
}

func (r *MessageRepository) FindByClientMsgID(
	ctx context.Context,
	conversationID, senderUserID, clientMsgID uuid.UUID,
) (*model.Message, error) {
	const query = `
		SELECT
			message_id,
			conversation_id,
			sender_user_id,
			client_msg_id,
			text,
			created_at
		FROM messages
		WHERE conversation_id = $1
		  AND sender_user_id = $2
		  AND client_msg_id = $3
	`

	var message model.Message
	err := chatdb.GetExecutor(ctx, r.db).QueryRow(ctx, query, conversationID, senderUserID, clientMsgID).Scan(
		&message.ID,
		&message.ConversationID,
		&message.SenderUserID,
		&message.ClientMsgID,
		&message.Text,
		&message.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}

	return &message, nil
}

func (r *MessageRepository) ListHistory(
	ctx context.Context,
	conversationID uuid.UUID,
	beforeMessageID *uuid.UUID,
	limit int,
) (ports.MessageHistoryPage, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT
			m.message_id,
			m.conversation_id,
			m.sender_user_id,
			m.client_msg_id,
			m.text,
			m.created_at
		FROM messages m
	`

	args := []any{conversationID}
	argPos := 2

	if beforeMessageID != nil {
		query += `
			INNER JOIN messages anchor
				ON anchor.message_id = $2
		`
		argPos++
	}

	query += `
		WHERE m.conversation_id = $1
	`

	if beforeMessageID != nil {
		query += `
		  AND anchor.conversation_id = $1
		  AND (m.created_at, m.message_id) < (anchor.created_at, anchor.message_id)
		`
		args = append(args, *beforeMessageID)
	}

	query += fmt.Sprintf(`
		ORDER BY m.created_at DESC, m.message_id DESC
		LIMIT $%d
	`, argPos)

	args = append(args, limit+1)

	rows, err := chatdb.GetExecutor(ctx, r.db).Query(ctx, query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ports.MessageHistoryPage{}, ports.ErrNotFound
		}
		return ports.MessageHistoryPage{}, err
	}
	defer rows.Close()

	messages := make([]model.Message, 0, limit)
	hasMore := false

	for rows.Next() {
		var message model.Message
		if err := rows.Scan(
			&message.ID,
			&message.ConversationID,
			&message.SenderUserID,
			&message.ClientMsgID,
			&message.Text,
			&message.CreatedAt,
		); err != nil {
			return ports.MessageHistoryPage{}, err
		}

		if len(messages) == limit {
			hasMore = true
			break
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return ports.MessageHistoryPage{}, err
	}

	return ports.MessageHistoryPage{
		Messages: messages,
		HasMore:  hasMore,
	}, nil
}
