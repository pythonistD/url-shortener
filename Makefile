.PHONY: lint test build run clean smoke compose-up compose-down

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

clean:
	rm -rf bin/
