#!/usr/bin/env bash
set -euo pipefail

for db in prod stage test; do
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-SQL
    DO \$\$
BEGIN
      IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = '$db') THEN
        CREATE DATABASE $db;
END IF;
END
\$\$;
SQL
done

psql --username "$POSTGRES_USER" <<-SQL
  DO \$\$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'auth_svc') THEN
CREATE ROLE auth_svc LOGIN PASSWORD 'authpass';
END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'content_svc') THEN
CREATE ROLE content_svc LOGIN PASSWORD 'contentpass';
END IF;
END
\$\$;
SQL

for db in prod stage test; do
psql --username "$POSTGRES_USER" --dbname "$db" <<-SQL
CREATE SCHEMA IF NOT EXISTS auth    AUTHORIZATION auth_svc;
CREATE SCHEMA IF NOT EXISTS content AUTHORIZATION content_svc;

GRANT USAGE ON SCHEMA auth    TO auth_svc;
    GRANT USAGE ON SCHEMA content TO content_svc;
SQL
done

echo "Databases & schemas created successfully."

if command -v mc >/dev/null 2>&1 && [[ -n "${MINIO_ENDPOINT:-}" ]]; then
  echo "Configuring MinIO buckets…"
  mc alias set kultura "$MINIO_ENDPOINT" "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" --api S3v4
  mc mb    kultura/kultura-content   || true
  mc mb    kultura/kultura-audio     || true
  mc policy set public kultura/kultura-content
  mc policy set public kultura/kultura-audio
  echo "MinIO buckets ready."
fi