.PHONY: test fmt vet build build-bot build-bot-linux-arm64 run-cli run-bot tidy

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

build-bot:
	go build -o bin/discord-ai-bot ./cmd/bot

# Cross-compile for Oracle Cloud A1 (Ampere Arm64) instances.
# CGO is disabled so the binary runs on a minimal Ubuntu image.
build-bot-linux-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
		go build -o bin/discord-ai-bot-linux-arm64 ./cmd/bot

run-cli:
	go run ./cmd/cli $(ARGS)

run-bot:
	go run ./cmd/bot
