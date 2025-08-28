-- +goose Up
CREATE TABLE users (
    id UUID,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    name TEXT NOT NULL
);

-- +goose Down
DROP TABLE users;
