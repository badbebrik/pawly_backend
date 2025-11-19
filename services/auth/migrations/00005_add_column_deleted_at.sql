-- +goose Up

ALTER TABLE users
    ADD deleted_at TIMESTAMP DEFAULT NULL;

-- +goose Down
DROP TABLE users;