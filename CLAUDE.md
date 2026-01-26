# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Inceptor is a self-hosted crash logging service for Flutter and mobile apps. It provides crash ingestion via REST/gRPC, intelligent crash grouping using fingerprinting, a Nuxt 3 web dashboard, and alerting (webhooks, email, Slack).

## Build Commands

```bash
make deps             # Download and tidy Go dependencies
make build            # Build binary to bin/inceptor
make run              # Run application
make test             # Run all tests
make test-coverage    # Run tests with HTML coverage report
make proto            # Generate gRPC code from .proto files (requires protoc)
make lint             # Run golangci-lint
make docker-build     # Build Docker image
make docker-run       # Run Docker container
```

**Dashboard (Nuxt 3):**
```bash
make dashboard-install  # npm install in web/
make dashboard-dev      # Run dev server on port 3000
make dashboard-build    # Production build
```

**Flutter SDK:**
```bash
make flutter-sdk-deps   # flutter pub get
make flutter-sdk-test   # Run Flutter tests
```

## Architecture

```
┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐
│  REST API    │  │  gRPC API    │  │   Web Dashboard      │
│  (Gin)       │  │  (protobuf)  │  │   (Nuxt 3)           │
└──────┬───────┘  └──────┬───────┘  └──────────┬───────────┘
       └─────────────────┼──────────────────────┘
                         ▼
┌────────────────────────────────────────────────────────────┐
│                    Core Service Layer                      │
│  • Crash Processor   • Crash Grouper (fingerprinting)     │
│  • Alert Manager     • Retention Manager (cleanup)        │
└────────────────────────────────────────────────────────────┘
       │                 │                 │
       ▼                 ▼                 ▼
┌────────────┐  ┌──────────────┐  ┌────────────────┐
│  SQLite    │  │  File Store  │  │  Alert Queue   │
│  (index)   │  │  (raw logs)  │  │  (in-memory)   │
└────────────┘  └──────────────┘  └────────────────┘
```

## Key Directories

- `cmd/inceptor/` - Application entry point and server initialization
- `internal/api/rest/` - REST handlers, middleware, routes (Gin)
- `internal/api/grpc/` - gRPC service implementation
- `internal/config/` - Viper-based configuration (YAML + env vars)
- `internal/core/` - Data models, fingerprinting, alerting, retention
- `internal/storage/` - Repository interface, SQLite and file store implementations
- `api/proto/` - Protocol Buffer definitions
- `web/` - Nuxt 3 dashboard
- `sdk/flutter/` - Flutter SDK

## Configuration

Configuration uses Viper with environment variable prefix `INCEPTOR_`:
- `INCEPTOR_SERVER_REST_PORT` (default: 8080)
- `INCEPTOR_SERVER_GRPC_PORT` (default: 9090)
- `INCEPTOR_STORAGE_SQLITE_PATH` (default: ./data/inceptor.db)
- `INCEPTOR_STORAGE_LOGS_PATH` (default: ./data/crashes)
- `INCEPTOR_AUTH_ADMIN_KEY` - Admin API key for management operations
- `INCEPTOR_AUTH_ENABLED` (default: true)

See `configs/config.example.yaml` for full configuration options.

## Key Technical Details

**Crash Fingerprinting** (`internal/core/grouper.go`): SHA256 hash of error type + top 5 non-native stack frames (normalized to remove line numbers, memory addresses, closure IDs). Results in 16-char hex fingerprint for grouping similar crashes.

**Authentication**: REST API uses `X-API-Key` header. App API keys for crash submission; admin key for management. Middleware chain: `APIKeyAuth()` → `AdminOnly()` for admin routes.

**Repository Pattern**: Storage abstraction in `internal/storage/repository.go` with SQLite (indexed queries) and LocalFileStore (full crash payloads as JSON) implementations.

**Retention**: Background goroutine in `internal/core/retention.go` runs on configurable schedule (default 24h) to delete old crashes per app retention settings.

## REST API Routes

```
/health, /ready - Health checks (no auth)

/api/v1/crashes       POST (app key), GET/DELETE (app or admin)
/api/v1/crashes/:id   GET/DELETE
/api/v1/groups        GET, PATCH (update status/notes)
/api/v1/apps          POST/GET (admin only)
/api/v1/apps/:id      GET, GET /stats
/api/v1/alerts        POST/GET/DELETE (admin only)
```

## Running Tests

```bash
make test                           # All tests
go test -v ./internal/core/...      # Specific package
go test -run TestGrouper ./...      # Specific test
```
