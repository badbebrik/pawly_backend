-- +goose Up
CREATE TABLE conversations (
    conversation_id UUID PRIMARY KEY,
    pet_id UUID NOT NULL,
    user_low_id UUID NOT NULL,
    user_high_id UUID NOT NULL,
    last_message_id UUID NULL,
    last_message_at TIMESTAMPTZ NULL,
    last_message_preview TEXT NULL,
    last_message_sender_id UUID NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX conversations_pet_user_pair_uq
    ON conversations (pet_id, user_low_id, user_high_id);

CREATE TABLE conversation_participants (
    conversation_id UUID NOT NULL REFERENCES conversations(conversation_id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    last_read_message_id UUID NULL,
    last_read_at TIMESTAMPTZ NULL,
    unread_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (conversation_id, user_id),
    CONSTRAINT conversation_participants_unread_non_negative CHECK (unread_count >= 0)
);

CREATE INDEX conversation_participants_user_id_idx
    ON conversation_participants (user_id);

CREATE TABLE messages (
    message_id UUID PRIMARY KEY,
    conversation_id UUID NOT NULL REFERENCES conversations(conversation_id) ON DELETE CASCADE,
    sender_user_id UUID NOT NULL,
    client_msg_id UUID NOT NULL,
    text TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX messages_conversation_created_at_idx
    ON messages (conversation_id, created_at DESC, message_id DESC);

CREATE UNIQUE INDEX messages_dedup_uq
    ON messages (conversation_id, sender_user_id, client_msg_id);

-- +goose Down
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversation_participants;
DROP TABLE IF EXISTS conversations;
