CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
                       id            BIGSERIAL PRIMARY KEY,
                       email         CITEXT  NOT NULL UNIQUE,
                       nickname      TEXT    NOT NULL UNIQUE,
                       password_hash BYTEA   NOT NULL,
                       provider      TEXT    NOT NULL DEFAULT 'local',
                       provider_id   TEXT,
                       created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE refresh_tokens (
                                token       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                expires_at  TIMESTAMPTZ NOT NULL
);