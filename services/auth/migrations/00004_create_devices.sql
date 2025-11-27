-- +goose Up
CREATE TABLE devices (
      id UUID PRIMARY KEY,
      user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      device_id VARCHAR(255),
      platform int2,
      app_version VARCHAR(64),
      locale VARCHAR(10),
      fcm_token TEXT,
      is_active BOOLEAN NOT NULL DEFAULT TRUE,
      created_at TIMESTAMP NOT NULL DEFAULT NOW(),
      updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

      UNIQUE(user_id, device_id),
      UNIQUE(fcm_token)
);

CREATE INDEX idx_devices_user ON devices(user_id);
CREATE INDEX idx_devices_active ON devices(is_active);

-- +goose Down
DROP TABLE user_devices;
