.PHONY: lint test build run clean

lint:
	golangci-lint run ./...

test:
	go test -v -race -cover ./...

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

clean:
	rm -rf bin/

