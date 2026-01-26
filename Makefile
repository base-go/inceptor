.PHONY: all build run test clean proto deps lint

# Binary name
BINARY=inceptor
# Build directory
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

all: deps build

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Build the binary
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY) ./cmd/inceptor

# Build for multiple platforms
build-all:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/inceptor
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-linux-arm64 ./cmd/inceptor
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/inceptor
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/inceptor
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe ./cmd/inceptor

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
