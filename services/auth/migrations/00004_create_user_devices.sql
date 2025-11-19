-- +goose Up
CREATE TABLE user_devices (
      id UUID PRIMARY KEY,
      user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      device_token VARCHAR(255),
      platform VARCHAR(32),
      created_at TIMESTAMP NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE user_devices;
