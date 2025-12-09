.PHONY: run build test clean migrate

# Run the server
run:
	go run cmd/server/main.go

# Build the binary
build:
	go build -o bin/wallet-service cmd/server/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f wallet.db wallet.db-shm wallet.db-wal

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Create .env from example
setup:
	cp .env.example .env
	@echo "Please edit .env with your credentials"

# Development
dev:
	air

# Production build
prod-build:
	CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o bin/wallet-service cmd/server/main.go