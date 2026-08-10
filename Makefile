APP_NAME=tinyurl

DATABASE_DSN?=postgres://postgres:postgres@localhost:5432/tinyurl
SIGN_KEY?=qwerty

run-slice:
	SIGN_KEY=${SIGN_KEY} go run ./cmd/shortener -l debug -b http://127.0.0.1:8080 -a 127.0.0.1:8080

run-slice-with-grpc:
	SIGN_KEY=${SIGN_KEY} go run ./cmd/shortener -l debug -b http://127.0.0.1:8080 -a 127.0.0.1:8080 -g :7777

run-slice-with-meta:
	SIGN_KEY=${SIGN_KEY} go run -ldflags "-X main.buildVersion=v1.0.0 -X 'main.buildDate=$(shell date +'%Y/%m/%d %H:%M:%S')' -X 'main.buildCommit=$(shell git rev-parse HEAD)'" ./cmd/shortener -l debug -b http://127.0.0.1:8080 -a 127.0.0.1:8080

run-slice-https:
	SIGN_KEY=${SIGN_KEY} go run ./cmd/shortener -l debug -s -b https://127.0.0.1:8080 -a 127.0.0.1:8080

run-file:
	SIGN_KEY=${SIGN_KEY} go run ./cmd/shortener -l debug -f ./store.json -b http://127.0.0.1:8080 -a 127.0.0.1:8080

run-file-with-config:
	SIGN_KEY=${SIGN_KEY} go run ./cmd/shortener -c ./config.json

run-file-audit:
	SIGN_KEY=${SIGN_KEY} go run ./cmd/shortener -l debug -f ./store.json --audit-file ./audit.json --audit-url http://127.0.0.1 -b http://127.0.0.1:8080 -a 127.0.0.1:8080

run-pgx:
	SIGN_KEY=${SIGN_KEY} DATABASE_DSN=${DATABASE_DSN} go run ./cmd/shortener -l debug -b http://127.0.0.1:8080 -a 127.0.0.1:8080

run-slice-subnet:
	SIGN_KEY=${SIGN_KEY} go run ./cmd/shortener -l debug -b http://127.0.0.1:8080 -a 127.0.0.1:8080 -t 192.168.1.0/24

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

genreset:
	go run ./cmd/reset

crt:
	go run ./cmd/cert

proto-gen:
	protoc -I=. -I=/opt/protoc/include --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --go_opt=default_api_level=API_OPAQUE proto/shortener.proto