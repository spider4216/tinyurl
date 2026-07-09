APP_NAME=tinyurl

DATABASE_DSN?=postgres://postgres:postgres@localhost:5432/tinyurl

run-slice:
	go run ./cmd/shortener -l debug

run-file:
	go run ./cmd/shortener -l debug -f ./store.json

run-file-audit:
	go run ./cmd/shortener -l debug -f ./store.json --audit-file ./audit.json --audit-url http://127.0.0.1

run-pgx:
	DATABASE_DSN=${DATABASE_DSN} go run ./cmd/shortener -l debug

test:
	go test ./...

test-v:
	go test ./... -v

bench:
	go test -bench=. -benchmem -benchtime=100ms ./...

lint:
	golangci-lint run

fmt:
	golangci-lint fmt

build:
	go build -o ./cmd/shortener ./cmd/shortener

migration-gen:
	migrate create -ext sql -dir ./migrations -seq $(name)

migrate-up:
	migrate -path ./migrations -database $(dsn) up $(ver)

migrate-down:
	migrate -path ./migrations -database $(dsn) down $(ver)

migrate-force:
	migrate -path ./migrations -database $(dsn) force $(ver)

doc:
	godoc -http=:8080 -play

doc-open:
	firefox http://127.0.0.1:8080/pkg/github.com/spider4216/tinyurl/?m=all

mcheck:
	go run ./cmd/staticclient ./...