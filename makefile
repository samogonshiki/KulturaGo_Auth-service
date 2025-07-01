PROJECT   ?= auth-service
IMAGE      = kulturago/$(PROJECT)
LOG_DIR    = ./logs

.PHONY: dev stage prod build docker test swag dc-up

swag:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/auth/main.go -o api/docs

test:
	go test ./...

build:
	go build -v ./cmd/auth

docker:
	docker build -t $(IMAGE):latest .

dc-down:
	docker-compose down --volumes --remove-orphans
	rm -rf kafka-data


dc-up:
	docker compose up -d --build

mig:
	docker exec -i auth-service-postgres-1 psql \
          -U root \
          -d postgres \
          < ./db/migrations/0001_init.up.sql

dev: export LOG_LEVEL = debug
dev: export LOG_FILE  = $(LOG_DIR)/dev.log
dev: swag dc-up

stage: export LOG_LEVEL = warn
stage: swag dc-up

prod: export LOG_LEVEL = error
prod: swag dc-up