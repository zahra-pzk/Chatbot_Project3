-- +goose Up
-- +goose StatementBegin
CREATE TYPE account_status AS ENUM ('incomplete','awaiting_verification', 'verified', 'disapproved', 'suspended');

CREATE TABLE user_account (
    account_id          BIGSERIAL,
    account_external_id UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    user_external_id    UUID             NOT NULL,
    status              account_status   NOT NULL DEFAULT 'incomplete',
    birth_date          DATE,
    photos              TEXT[]           DEFAULT '{}',
    updated_at          TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    
    CONSTRAINT fk_account_user
        FOREIGN KEY (user_external_id) 
        REFERENCES users (user_external_id)
        ON DELETE RESTRICT,
        
    CONSTRAINT max_photos_check CHECK (array_length(photos, 1) <= 10),
    CONSTRAINT check_birth_date_max CHECK (birth_date <= updated_at::DATE),
    CONSTRAINT check_birth_date_min CHECK (birth_date >= (updated_at - INTERVAL '120 years')::DATE)
);

CREATE INDEX idx_account_status ON user_account(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_account_status;
ALTER TABLE user_account DROP CONSTRAINT IF EXISTS check_birth_date_max;
ALTER TABLE user_account DROP CONSTRAINT IF EXISTS check_birth_date_min;
DROP TABLE IF EXISTS user_account;
DROP TYPE IF EXISTS account_status;
-- +goose StatementEnd