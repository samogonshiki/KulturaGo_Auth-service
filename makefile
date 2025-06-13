PROJECT=auth-service

.PHONY: dev build docker test swag

dev: swag docker-compose-up

swag:
	swag init -g cmd/auth/main.go -o api/docs

test:
	go test ./...

build:
	go build -v ./cmd/auth

docker:
	docker build -t kulturago/$(PROJECT):latest .

docker-compose-up:
	docker compose up -d --build