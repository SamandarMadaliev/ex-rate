BINARY  := ex-rate
PORT    ?= 8080

.PHONY: run build test tidy swagger migrate-up migrate-down migrate-create

run:
	go run ./cmd/main.go

build:
	go build -o bin/$(BINARY) .

test:
	go test ./...

tidy:
	go mod tidy

swagger:
	swag init -g cmd/main.go -o docs --parseInternal --parseDependency

migrate-up:
	docker compose up migrate

migrate-down:
	docker compose run --rm --entrypoint sh migrate -c 'migrate -path=/migrations -database "postgres://$$DATABASE_USER:$$DATABASE_PASSWORD@$$DATABASE_HOST:$$DATABASE_PORT/$$DATABASE_NAME?sslmode=$$DATABASE_SSLMODE" down 1'

migrate-create:
	docker compose run --rm --entrypoint sh migrate -c 'migrate create -ext sql -dir /migrations -seq $(name)'

