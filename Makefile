.PHONY: all build run test clean proto deps lint release

# Binary name
BINARY=inceptor
# Build directory
BUILD_DIR=bin
# Version (override with VERSION=x.y.z)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
# LDFLAGS
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

all: deps web build

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Build the web dashboard
web:
	cd web && npm install && npm run generate
	rm -rf internal/api/rest/static/* 2>/dev/null || true
	mkdir -p internal/api/rest/static
	cp -r web/.output/public/* internal/api/rest/static/

# Build the binary
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/inceptor

# Build for multiple platforms
build-all: web
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/inceptor
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-arm64 ./cmd/inceptor
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/inceptor
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/inceptor
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe ./cmd/inceptor

# Release (requires VERSION to be set)
release: build-all
	@echo "Built release $(VERSION)"
	@ls -la $(BUILD_DIR)/

# Run the application
run:
	$(GORUN) ./cmd/inceptor

# Run with config file
run-config:
	$(GORUN) ./cmd/inceptor -config ./configs/config.yaml

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Generate proto files
proto:
	protoc --go_out=. --go-grpc_out=. api/proto/*.proto

# Lint the code
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Create data directories
init-dirs:
	mkdir -p data/crashes

# Development setup
dev-setup: deps init-dirs
	cp configs/config.example.yaml configs/config.yaml

# Docker build
docker-build:
	docker build -t inceptor:latest .

# Docker run
docker-run:
	docker run -p 8080:8080 -p 9090:9090 -v $(PWD)/data:/app/data inceptor:latest

# Docker compose
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

# Dashboard build (Nuxt)
dashboard-install:
	cd web && npm install

dashboard-dev:
	cd web && npm run dev

dashboard-build:
	cd web && npm run build

# Flutter SDK
flutter-sdk-deps:
	cd sdk/flutter && flutter pub get

flutter-sdk-test:
	cd sdk/flutter && flutter test

# Help
help:
	@echo "Available targets:"
	@echo "  all            - Download dependencies and build"
	@echo "  deps           - Download dependencies"
	@echo "  build          - Build the binary"
	@echo "  build-all      - Build for all platforms"
	@echo "  run            - Run the application"
	@echo "  run-config     - Run with config file"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  proto          - Generate proto files"
	@echo "  lint           - Run linter"
	@echo "  clean          - Clean build artifacts"
	@echo "  dev-setup      - Setup development environment"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  dashboard-dev  - Run dashboard in development mode"
	@echo "  dashboard-build- Build dashboard for production"
