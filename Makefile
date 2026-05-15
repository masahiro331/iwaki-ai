.PHONY: test fmt vet build run-cli tidy

test:
	go test ./...

fmt:
	gofmt -s -w .

vet:
	go vet ./...

tidy:
	go mod tidy

build:
	go build -o bin/discord-ai-cli ./cmd/cli

run-cli:
	go run ./cmd/cli $(ARGS)
