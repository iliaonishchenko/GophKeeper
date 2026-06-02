-- +goose Up
CREATE TABLE items (
    id         VARCHAR(64) PRIMARY KEY,
    user_id    VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       SMALLINT NOT NULL,
    name       VARCHAR(255) NOT NULL,
    ciphertext BYTEA NOT NULL,
    metadata   TEXT NOT NULL DEFAULT '',
    version    BIGINT NOT NULL DEFAULT 1,
    updated_at BIGINT NOT NULL,
    deleted    BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_items_user_updated ON items (user_id, updated_at);

-- +goose Down
DROP TABLE items;
