-- +goose Up
CREATE TABLE users (
    id         VARCHAR(64) PRIMARY KEY,
    login      VARCHAR(255) NOT NULL UNIQUE,
    auth_hash  BYTEA NOT NULL,
    auth_salt  BYTEA NOT NULL,
    enc_salt   BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE users;
