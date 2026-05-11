-include .env
export

.PHONY: run run-worker lint test migrate-up migrate-down migrate-create build swagger

# Requer: go install github.com/swaggo/swag/cmd/swag@latest (swag no PATH).
# Gera docs/ para go build funcionar sem passo extra; commitar docs/ após mudanças nas rotas.
swagger:
	swag init -g main.go -d cmd/api,internal/auth,internal/user,internal/order,internal/platform/httperr,internal/platform/pagination -o docs

DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

run:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

lint:
	golangci-lint run ./...

test:
	go test ./...

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down

migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=your_migration_name" && exit 1)
	migrate create -ext sql -dir migrations -seq $(name)

build:
	mkdir -p bin
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker
