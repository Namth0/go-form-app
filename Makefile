.PHONY: help build test test-verbose test-coverage clean docker-build docker-run docker-test docker-stop lint fmt

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Go commands
build: ## Build the application
	go build -o main main.go

test: ## Run tests
	go test ./...

test-verbose: ## Run tests with verbose output
	go test -v ./...

test-coverage: ## Run tests with coverage report
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race: ## Run tests with race detection
	go test -race ./...

bench: ## Run benchmarks
	go test -bench=. ./...

lint: ## Run linter
	golangci-lint run

fmt: ## Format code
	go fmt ./...

clean: ## Clean build artifacts
	rm -f main main.exe coverage.out coverage.html

# Docker commands
docker-build: ## Build Docker image
	cd docker && docker-compose build

docker-run: ## Run application with Docker Compose
	cd docker && docker-compose up

docker-run-detached: ## Run application with Docker Compose in background
	cd docker && docker-compose up -d

docker-test: ## Run tests in Docker container
	docker run --rm -v $(PWD):/app -w /app golang:1.23.4-alpine sh -c "apk add --no-cache git && go test -v ./..."

docker-stop: ## Stop Docker containers
	cd docker && docker-compose down

docker-clean: ## Clean Docker containers and images
	cd docker && docker-compose down -v
	docker system prune -f

# Development commands
dev: ## Run application locally
	go run main.go

watch: ## Run application with file watching (requires entr)
	find . -name "*.go" | entr -r go run main.go

# Production commands
docker-prod: ## Run in production mode with Nginx
	cd docker && docker-compose --profile production up -d

# Install dependencies for development
install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Quick development cycle
quick-test: fmt test ## Format code and run tests 