.PHONY: lint test build run clean smoke compose-up compose-down migrate-up migrate-down migrate-status

-include .env
export

POSTGRES_USER ?= postgres
POSTGRES_PORT_HOST ?= 5433
POSTGRES_DB ?= url_shortener

DSN ?= "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT_HOST)/$(POSTGRES_DB)?sslmode=disable"

lint:
	golangci-lint run ./...

test:
	go test -v -race -cover ./...

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

smoke:
	./scripts/smoke-tests.sh

compose-up: build
	docker-compose up -d --build

compose-down:
	docker-compose down

migrate-up:
	goose -dir migrations postgres $(DSN) up

migrate-down:
	goose -dir migrations postgres $(DSN) down

migrate-status:
	goose -dir migrations postgres $(DSN) status

clean:
	rm -rf bin/
