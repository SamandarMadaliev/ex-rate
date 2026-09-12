BINARY  := ex-rate
PORT    ?= 8080

.PHONY: run build test

run:
	go run ./cmd/main.go

build:
	go build -o bin/$(BINARY) .

test:
	go test ./...

tidy:
	go mod tidy

