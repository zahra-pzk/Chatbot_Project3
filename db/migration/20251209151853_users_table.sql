-- +goose Up
-- +goose StatementBegin
CREATE TYPE role_type AS ENUM ('superadmin', 'admin', 'user', 'guest', 'system');
CREATE TYPE account_status AS ENUM ('incomplete','awaiting_verification', 'verified', 'disapproved', 'suspended');

CREATE TABLE users (
    user_id             BIGSERIAL,
    user_external_id    UUID                PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name          VARCHAR(255)        NOT NULL,
    last_name           VARCHAR(255)        NOT NULL,
    username            VARCHAR(255)        UNIQUE,
    phone_number        VARCHAR(50)         UNIQUE,
    email               VARCHAR(255)        UNIQUE NOT NULL,
    hashed_password     TEXT,
    role                role_type           NOT NULL DEFAULT 'guest', 
    created_at  TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at  TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    status              account_status      NOT NULL DEFAULT 'incomplete',
    birth_date          DATE,
    photos              TEXT[]              DEFAULT '{}',
    CONSTRAINT max_photos_check             CHECK (array_length(photos, 1) <= 10),
    CONSTRAINT check_birth_date_max         CHECK (birth_date <= updated_at::DATE),
    CONSTRAINT check_birth_date_min         CHECK (birth_date >= (updated_at - INTERVAL '120 years')::DATE)
);

CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_account_status ON users(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_role;
DROP TYPE IF EXISTS role_type;
DROP INDEX IF EXISTS idx_account_status;
ALTER TABLE users DROP CONSTRAINT IF EXISTS check_birth_date_max;
ALTER TABLE users DROP CONSTRAINT IF EXISTS check_birth_date_min;
DROP TYPE IF EXISTS account_status;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
