.PHONY: run build test fmt vet

run:
	go run ./cmd/server

build:
	go build -o bin/flashy-server ./cmd/server

test:
	go test ./... -cover

fmt:
	gofmt -w .

vet:
	go vet ./...
