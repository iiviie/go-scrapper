.PHONY: build run clean test deps

# Build the application
build:
	go build -o bin/scraper cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Run with hot reload (requires air: go install github.com/air-verse/air@latest)
dev:
	air

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf data/*.db

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Create necessary directories
setup:
	mkdir -p data
	mkdir -p bin

# Run all checks
check: fmt test lint
