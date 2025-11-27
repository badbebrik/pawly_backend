-- +goose Up
CREATE TABLE sessions (
      id UUID PRIMARY KEY,
      user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      refresh_token_hash VARCHAR(255) NOT NULL,
      expires_at TIMESTAMP NOT NULL,
      is_revoked bool NOT NULL DEFAULT FALSE,
      created_at TIMESTAMP NOT NULL DEFAULT now(),
      updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_active ON sessions(user_id, is_revoked);

-- +goose Down
DROP TABLE sessions;
