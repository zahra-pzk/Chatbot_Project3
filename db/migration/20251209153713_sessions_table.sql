-- +goose Up
-- +goose StatementBegin

CREATE TABLE sessions (
    session_id          BIGSERIAL,
    session_external_id UUID                PRIMARY KEY DEFAULT gen_random_uuid(),
    user_agent          VARCHAR(255)        NOT NULL,
    username            VARCHAR(255)        NOT NULL,
    user_external_id    UUID                NOT NULL,
    is_blocked          BOOLEAN             NOT NULL DEFAULT false,
    client_ip           VARCHAR(255)        NOT NULL,
    refresh_token       VARCHAR             NOT NULL,
    expires_at          TIMESTAMPTZ         NOT NULL
);

ALTER TABLE "sessions" ADD FOREIGN KEY ("username") REFERENCES "users" ("username");
ALTER TABLE "sessions" ADD FOREIGN KEY ("user_external_id") REFERENCES "users" ("user_external_id");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
