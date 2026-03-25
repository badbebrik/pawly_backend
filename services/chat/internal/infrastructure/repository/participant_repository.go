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
	"github.com/jackc/pgx/v5/pgxpool"
)

type ParticipantRepository struct {
	db *pgxpool.Pool
}

func NewParticipantRepository(db *pgxpool.Pool) *ParticipantRepository {
	return &ParticipantRepository{db: db}
}

func (r *ParticipantRepository) CreateBatch(ctx context.Context, participants []model.ConversationParticipant) error {
	if len(participants) == 0 {
		return nil
	}

	const query = `
		INSERT INTO conversation_participants (
			conversation_id,
			user_id,
			last_read_message_id,
			last_read_at,
			unread_count,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	exec := chatdb.GetExecutor(ctx, r.db)
	for i := range participants {
		p := participants[i]
		if _, err := exec.Exec(
			ctx,
			query,
			p.ConversationID,
			p.UserID,
			p.LastReadMessageID,
			p.LastReadAt,
			p.UnreadCount,
			p.CreatedAt,
			p.UpdatedAt,
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *ParticipantRepository) GetByConversationAndUser(ctx context.Context, conversationID, userID uuid.UUID) (*model.ConversationParticipant, error) {
	const query = `
		SELECT
			conversation_id,
			user_id,
			last_read_message_id,
			last_read_at,
			unread_count,
			created_at,
			updated_at
		FROM conversation_participants
		WHERE conversation_id = $1
		  AND user_id = $2
	`

	row := chatdb.GetExecutor(ctx, r.db).QueryRow(ctx, query, conversationID, userID)
	participant, err := scanParticipant(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}

	return participant, nil
}

func (r *ParticipantRepository) ListByConversation(ctx context.Context, conversationID uuid.UUID) ([]model.ConversationParticipant, error) {
	const query = `
		SELECT
			conversation_id,
			user_id,
			last_read_message_id,
			last_read_at,
			unread_count,
			created_at,
			updated_at
		FROM conversation_participants
		WHERE conversation_id = $1
		ORDER BY user_id
	`

	rows, err := chatdb.GetExecutor(ctx, r.db).Query(ctx, query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ConversationParticipant, 0)
	for rows.Next() {
		item, err := scanParticipant(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}

	return items, rows.Err()
}

func (r *ParticipantRepository) IncrementUnread(ctx context.Context, conversationID, userID uuid.UUID, delta int) error {
	const query = `
		UPDATE conversation_participants
		SET
			unread_count = unread_count + $3,
			updated_at = NOW()
		WHERE conversation_id = $1
		  AND user_id = $2
	`

	cmd, err := chatdb.GetExecutor(ctx, r.db).Exec(ctx, query, conversationID, userID, delta)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ports.ErrNotFound
	}

	return nil
}

func (r *ParticipantRepository) MarkRead(
	ctx context.Context,
	conversationID, userID, lastReadMessageID uuid.UUID,
	readAt time.Time,
) (*model.ConversationParticipant, error) {
	const query = `
		WITH target_message AS (
			SELECT created_at, message_id
			FROM messages
			WHERE message_id = $3
			  AND conversation_id = $1
		)
		UPDATE conversation_participants
		SET
			last_read_message_id = $3,
			last_read_at = $4,
			unread_count = COALESCE((
				SELECT COUNT(*)::INT
				FROM messages m, target_message tm
				WHERE m.conversation_id = $1
				  AND m.sender_user_id <> $2
				  AND (m.created_at, m.message_id) > (tm.created_at, tm.message_id)
			), 0),
			updated_at = $4
		WHERE conversation_id = $1
		  AND user_id = $2
		RETURNING
			conversation_id,
			user_id,
			last_read_message_id,
			last_read_at,
			unread_count,
			created_at,
			updated_at
	`

	row := chatdb.GetExecutor(ctx, r.db).QueryRow(ctx, query, conversationID, userID, lastReadMessageID, readAt)
	participant, err := scanParticipant(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}

	return participant, nil
}

type participantScanner interface {
	Scan(dest ...any) error
}

func scanParticipant(s participantScanner) (*model.ConversationParticipant, error) {
	var participant model.ConversationParticipant
	if err := s.Scan(
		&participant.ConversationID,
		&participant.UserID,
		&participant.LastReadMessageID,
		&participant.LastReadAt,
		&participant.UnreadCount,
		&participant.CreatedAt,
		&participant.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &participant, nil
}

var _ ports.ParticipantRepository = (*ParticipantRepository)(nil)
