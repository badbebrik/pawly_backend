package repository

import (
	"chat/internal/application/ports"
	"chat/internal/domain/model"
	chatdb "chat/internal/infrastructure/db"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
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

func (r *ConversationRepository) GetUnreadSummary(ctx context.Context, userID uuid.UUID) (ports.UnreadSummary, error) {
	const query = `
		SELECT
			COUNT(*) FILTER (WHERE unread_count > 0),
			COALESCE(SUM(unread_count), 0)
		FROM conversation_participants
		WHERE user_id = $1
	`

	var summary ports.UnreadSummary
	if err := r.db.QueryRow(ctx, query, userID).Scan(
		&summary.UnreadConversations,
		&summary.UnreadMessages,
	); err != nil {
		return ports.UnreadSummary{}, err
	}

	return summary, nil
}

func (r *ConversationRepository) ListConversations(
	ctx context.Context,
	userID uuid.UUID,
	params ports.ListConversationsParams,
) (ports.ListConversationsResult, error) {
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}

	var (
		cursorSortAt time.Time
		cursorID     uuid.UUID
		hasCursor    bool
	)
	if params.Cursor != nil && *params.Cursor != "" {
		var err error
		cursorSortAt, cursorID, err = decodeConversationCursor(*params.Cursor)
		if err != nil {
			return ports.ListConversationsResult{}, err
		}
		hasCursor = true
	}

	query := `
		SELECT
			c.conversation_id,
			c.pet_id,
			CASE
				WHEN c.user_low_id = $1 THEN c.user_high_id
				ELSE c.user_low_id
			END AS other_user_id,
			c.last_message_id,
			c.last_message_at,
			c.last_message_preview,
			c.last_message_sender_id,
			cp.last_read_message_id,
			cp.unread_count,
			COALESCE(c.last_message_at, c.created_at) AS sort_at
		FROM conversations c
		INNER JOIN conversation_participants cp
			ON cp.conversation_id = c.conversation_id
		WHERE cp.user_id = $1
	`

	args := []any{userID}
	argPos := 2

	if params.PetID != nil {
		query += fmt.Sprintf(" AND c.pet_id = $%d", argPos)
		args = append(args, *params.PetID)
		argPos++
	}

	if hasCursor {
		query += fmt.Sprintf(
			" AND (COALESCE(c.last_message_at, c.created_at), c.conversation_id) < ($%d, $%d)",
			argPos,
			argPos+1,
		)
		args = append(args, cursorSortAt, cursorID)
		argPos += 2
	}

	query += fmt.Sprintf(`
		ORDER BY COALESCE(c.last_message_at, c.created_at) DESC, c.conversation_id DESC
		LIMIT $%d
	`, argPos)
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return ports.ListConversationsResult{}, err
	}
	defer rows.Close()

	items := make([]ports.ConversationListRow, 0, limit)
	var nextCursor *string

	for rows.Next() {
		var (
			item   ports.ConversationListRow
			sortAt time.Time
		)
		if err := rows.Scan(
			&item.ConversationID,
			&item.PetID,
			&item.OtherUserID,
			&item.LastMessageID,
			&item.LastMessageAt,
			&item.LastMessagePreview,
			&item.LastMessageSenderID,
			&item.LastReadMessageID,
			&item.UnreadCount,
			&sortAt,
		); err != nil {
			return ports.ListConversationsResult{}, err
		}

		if len(items) == limit {
			cursor := encodeConversationCursor(sortAt, item.ConversationID)
			nextCursor = &cursor
			break
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return ports.ListConversationsResult{}, err
	}

	return ports.ListConversationsResult{
		Items:      items,
		NextCursor: nextCursor,
	}, nil
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

func encodeConversationCursor(sortAt time.Time, conversationID uuid.UUID) string {
	raw := sortAt.UTC().Format(time.RFC3339Nano) + "|" + conversationID.String()
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeConversationCursor(cursor string) (time.Time, uuid.UUID, error) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}

	parts := strings.SplitN(string(raw), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor")
	}

	sortAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}

	conversationID, err := uuid.Parse(parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}

	return sortAt, conversationID, nil
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
