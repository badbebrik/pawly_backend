-- +goose Up
CREATE TABLE oauth_identities (
      id UUID PRIMARY KEY,
      user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      provider VARCHAR(32) NOT NULL,
      external_id VARCHAR(255) NOT NULL,
      email VARCHAR(255),
      created_at TIMESTAMP NOT NULL DEFAULT now(),
      updated_at TIMESTAMP NOT NULL DEFAULT now(),
      UNIQUE(provider, external_id)
);

-- +goose Down
DROP TABLE oauth_identities;
