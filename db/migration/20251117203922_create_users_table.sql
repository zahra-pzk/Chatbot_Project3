-- +goose Up
-- +goose StatementBegin
CREATE TYPE role_type AS ENUM ('superadmin', 'admin', 'user', 'guest', 'system');

CREATE TABLE users (
    user_id             BIGSERIAL,
    user_external_id    UUID                PRIMARY KEY DEFAULT gen_random_uuid(),
    name                VARCHAR(255)        NOT NULL,
    username            VARCHAR(255)        UNIQUE,
    phone_number        VARCHAR(50)         UNIQUE,
    email               VARCHAR(255)        UNIQUE,
    password            TEXT,
    role                role_type           NOT NULL DEFAULT 'guest', 
    created_at  TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at  TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE INDEX idx_users_role ON users(role);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_role;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS role_type;
-- +goose StatementEnd
