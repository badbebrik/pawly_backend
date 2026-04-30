package realtime

import (
	rtcore "chat/internal/realtime"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisPresenceTracker struct {
	client *redis.Client
	prefix string
}

func NewRedisPresenceTracker(client *redis.Client, prefix string) *RedisPresenceTracker {
	return &RedisPresenceTracker{client: client, prefix: prefix}
}

func (t *RedisPresenceTracker) IsUserInConversation(ctx context.Context, conversationID, userID uuid.UUID) (bool, error) {
	key := t.key(conversationID, userID)
	now := float64(time.Now().UTC().Unix())
	if err := t.client.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%f", now)).Err(); err != nil {
		return false, err
	}
	count, err := t.client.ZCount(ctx, key, fmt.Sprintf("(%f", now), "+inf").Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (t *RedisPresenceTracker) SetInConversation(ctx context.Context, conversationID, userID, clientID uuid.UUID, ttl time.Duration) (rtcore.PresenceChange, bool, error) {
	return t.updatePresence(ctx, conversationID, userID, clientID, ttl, true)
}

func (t *RedisPresenceTracker) RefreshConversations(ctx context.Context, userID, clientID uuid.UUID, conversationIDs []uuid.UUID, ttl time.Duration) error {
	if len(conversationIDs) == 0 {
		return nil
	}

	pipeline := t.client.Pipeline()
	expiry := float64(time.Now().UTC().Add(ttl).Unix())
	for _, conversationID := range conversationIDs {
		key := t.key(conversationID, userID)
		pipeline.ZAdd(ctx, key, redis.Z{Score: expiry, Member: clientID.String()})
	}
	_, err := pipeline.Exec(ctx)
	return err
}

func (t *RedisPresenceTracker) ClearInConversation(ctx context.Context, conversationID, userID, clientID uuid.UUID) (rtcore.PresenceChange, bool, error) {
	return t.updatePresence(ctx, conversationID, userID, clientID, 0, false)
}

func (t *RedisPresenceTracker) ClearClient(ctx context.Context, userID, clientID uuid.UUID, conversationIDs []uuid.UUID) ([]rtcore.PresenceChange, error) {
	if len(conversationIDs) == 0 {
		return nil, nil
	}

	changes := make([]rtcore.PresenceChange, 0, len(conversationIDs))
	for _, conversationID := range conversationIDs {
		change, changed, err := t.ClearInConversation(ctx, conversationID, userID, clientID)
		if err != nil {
			return nil, err
		}
		if changed {
			changes = append(changes, change)
		}
	}
	return changes, nil
}

func (t *RedisPresenceTracker) updatePresence(
	ctx context.Context,
	conversationID, userID, clientID uuid.UUID,
	ttl time.Duration,
	present bool,
) (rtcore.PresenceChange, bool, error) {
	key := t.key(conversationID, userID)
	now := time.Now().UTC()
	nowScore := float64(now.Unix())
	before, err := t.hasActiveMembers(ctx, key, nowScore)
	if err != nil {
		return rtcore.PresenceChange{}, false, err
	}

	if present {
		if err := t.client.ZAdd(ctx, key, redis.Z{
			Score:  float64(now.Add(ttl).Unix()),
			Member: clientID.String(),
		}).Err(); err != nil {
			return rtcore.PresenceChange{}, false, err
		}
	} else {
		if err := t.client.ZRem(ctx, key, clientID.String()).Err(); err != nil {
			return rtcore.PresenceChange{}, false, err
		}
	}

	after, err := t.hasActiveMembers(ctx, key, nowScore)
	if err != nil {
		return rtcore.PresenceChange{}, false, err
	}

	change := rtcore.PresenceChange{
		ConversationID: conversationID,
		UserID:         userID,
		IsInChat:       after,
	}
	return change, before != after, nil
}

func (t *RedisPresenceTracker) hasActiveMembers(ctx context.Context, key string, now float64) (bool, error) {
	if err := t.client.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%f", now)).Err(); err != nil {
		return false, err
	}
	count, err := t.client.ZCount(ctx, key, fmt.Sprintf("(%f", now), "+inf").Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (t *RedisPresenceTracker) key(conversationID, userID uuid.UUID) string {
	return fmt.Sprintf("%s:conv:%s:user:%s", t.prefix, conversationID.String(), userID.String())
}
