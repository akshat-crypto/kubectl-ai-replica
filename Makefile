# MCP Servers Makefile
.PHONY: help build test clean install docker-build docker-run lint format

# Variables
BINARY_NAME=mcp-cli
BUILD_DIR=bin
DOCKER_IMAGE=mcp-servers/cli
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.Version=${VERSION}"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the CLI binary
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p ${BUILD_DIR}
	CGO_ENABLED=0 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} ./cmd/cli

build-all: ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	@mkdir -p ${BUILD_DIR}
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-linux-amd64 ./cmd/cli
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-darwin-amd64 ./cmd/cli
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME}-windows-amd64.exe ./cmd/cli

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

install: ## Install dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf ${BUILD_DIR}
	rm -f coverage.out coverage.html

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

format: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t ${DOCKER_IMAGE}:${VERSION} .
	docker tag ${DOCKER_IMAGE}:${VERSION} ${DOCKER_IMAGE}:latest

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -it --rm ${DOCKER_IMAGE}:latest

dev: ## Run in development mode
	@echo "Running in development mode..."
	go run ./cmd/cli

release: clean install test build-all ## Create release build
	@echo "Release build complete!"

.PHONY: setup
setup: ## Initial project setup
	@echo "Setting up project..."
	mkdir -p ${BUILD_DIR}
	mkdir -p configs
	mkdir -p logs
	cp configs/config.example.yaml configs/config.yaml
	@echo "Setup complete! Edit configs/config.yaml with your settings." 