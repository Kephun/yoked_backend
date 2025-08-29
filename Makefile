build:
	go build -o bin/server ./cmd/api

run:
	go run ./cmd/api

test:
	go test -v ./...

.PHONY: build run test
