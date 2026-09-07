.PHONY: test vet build run lint ci

test:
	go test ./... -count=1 -race -timeout 5m

vet:
	go vet ./...

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

lint:
	golangci-lint run --timeout=5m

ci: vet test build