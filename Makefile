APP_NAME=tinyurl

DATABASE_DSN?=postgres://postgres:postgres@localhost:5432/tinyurl

run-slice:
	go run ./cmd/shortener -l debug

run-file:
	go run ./cmd/shortener -l debug -f ./store.json

run-pgx:
	DATABASE_DSN=${DATABASE_DSN} go run ./cmd/shortener -l debug

test:
	go test ./...

test-v:
	go test ./... -v

lint:
	golangci-lint run

fmt:
	golangci-lint fmt

build:
	go build -o ./cmd/shortener ./cmd/shortener