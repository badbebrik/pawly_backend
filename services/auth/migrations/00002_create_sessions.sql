-- +goose Up
CREATE TABLE sessions (
      id UUID PRIMARY KEY,
      user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      refresh_token_hash VARCHAR(255) NOT NULL,
      expires_at TIMESTAMP NOT NULL,
      ip_address VARCHAR(64),
      created_at TIMESTAMP NOT NULL DEFAULT now(),
      updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE sessions;
