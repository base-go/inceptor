# Inceptor Documentation

Welcome to the Inceptor documentation. Inceptor is a self-hosted crash logging service designed primarily for Flutter applications but extensible to any platform.

## Overview

Inceptor provides a complete crash reporting solution with intelligent crash grouping, a web dashboard, alerting, and configurable retention policies. It runs entirely on your own infrastructure, giving you full control over your crash data.

## Documentation Index

| Document | Description |
|----------|-------------|
| [Getting Started](./getting-started.md) | Quick start guide for deploying Inceptor |
| [Architecture](./architecture.md) | System design and component overview |
| [API Reference](./api-reference.md) | Complete REST API documentation |
| [Flutter SDK](./flutter-sdk.md) | Integration guide for Flutter apps |
| [Configuration](./configuration.md) | All configuration options explained |
| [Deployment](./deployment.md) | Production deployment guide |
| [Alerting](./alerting.md) | Setting up notifications |

## Key Features

**Crash Ingestion**: Submit crash reports via REST API with full stack trace, device info, and custom metadata.

**Intelligent Grouping**: Similar crashes are automatically grouped using fingerprinting based on error type and normalized stack frames.

**Web Dashboard**: Built-in Nuxt 3 dashboard for viewing crashes, managing groups, and monitoring trends.

**Multi-Platform**: While optimized for Flutter, Inceptor works with any platform that can send HTTP requests.

**Retention Policies**: Automatic cleanup of old crash data based on configurable retention periods.

**Alerting**: Webhook, email (SMTP), and Slack notifications for new crashes and threshold breaches.

## Quick Start

```bash
# Clone the repository
git clone https://github.com/flakerimi/inceptor.git
cd inceptor

# Deploy with BasePod
bp push

# Or run locally with Go
go run cmd/inceptor/main.go
```

## Project Structure

```
inceptor/
├── cmd/inceptor/          # Application entry point
├── internal/
│   ├── api/rest/          # REST API handlers
│   ├── api/grpc/          # gRPC server (optional)
│   ├── core/              # Business logic
│   ├── storage/           # Database & file storage
│   └── config/            # Configuration management
├── web/                   # Nuxt 3 dashboard
├── sdk/flutter/           # Flutter SDK package
├── api/proto/             # gRPC protocol definitions
└── docs/                  # Documentation
```

## Requirements

- Go 1.22+ (for building from source)
- Node.js 20+ (for dashboard development)
- SQLite (embedded, no external dependencies)

## License

MIT License - See LICENSE file for details.
