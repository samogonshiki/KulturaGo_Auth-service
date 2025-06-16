CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
  id            BIGSERIAL PRIMARY KEY,
  email         TEXT UNIQUE,
  password_hash BYTEA,
  provider      TEXT DEFAULT 'local',
  provider_id   TEXT,
  created_at    TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE refresh_tokens (
  token       UUID PRIMARY KEY,
  user_id     BIGINT REFERENCES users(id) ON DELETE CASCADE,
  expires_at  TIMESTAMPTZ NOT NULL
);