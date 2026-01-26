# Inceptor

<p align="center">
  <strong>Self-hosted crash logging service for Flutter and mobile apps</strong>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#architecture">Architecture</a> •
  <a href="#api-reference">API Reference</a> •
  <a href="#flutter-sdk">Flutter SDK</a> •
  <a href="#dashboard">Dashboard</a> •
  <a href="#deployment">Deployment</a>
</p>

---

Inceptor is a lightweight, self-hosted crash logging and error tracking service designed primarily for Flutter applications but extensible to any platform. It provides crash ingestion via REST and gRPC, intelligent crash grouping, a modern web dashboard, alerting, and configurable retention policies.

## Features

| Feature | Description |
|---------|-------------|
| **Crash Collection** | Receive crash reports via REST API or gRPC |
| **Smart Grouping** | Automatically group similar crashes using fingerprinting algorithm |
| **Web Dashboard** | Modern Nuxt 3 + Nuxt UI dashboard for viewing and managing crashes |
| **Alerting** | Get notified via webhooks, email, or Slack when crashes occur |
| **Retention Policies** | Configurable per-app retention with automatic cleanup |
| **Flutter SDK** | Easy-to-use SDK with automatic error capture |
| **API Key Auth** | Secure per-app API keys with admin access control |
| **File + SQLite Storage** | Efficient hybrid storage with SQLite index and file-based logs |

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 18+ (for dashboard)
- Flutter 3.10+ (for SDK)

### Option 1: BasePod Deployment (Recommended)

```bash
# Clone the repository
git clone https://github.com/flakerimi/inceptor.git
cd inceptor

# Login to your BasePod server
bp login your-server.example.com

# Deploy
bp push
```

### Option 2: Manual Setup

```bash
# Clone and build
git clone https://github.com/flakerimi/inceptor.git
cd inceptor

# Install dependencies
go mod download

# Create config
cp configs/config.example.yaml configs/config.yaml
# Edit configs/config.yaml with your settings

# Run
go run cmd/inceptor/main.go
```

### Option 3: Docker

```bash
# Build
docker build -t inceptor:latest .

# Run
docker run -d \
  -p 8080:8080 \
  -p 9090:9090 \
  -v inceptor-data:/app/data \
  -e INCEPTOR_AUTH_ADMIN_KEY=your-secure-key \
  inceptor:latest
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Inceptor                                │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │  REST API    │  │  gRPC API    │  │   Web Dashboard      │  │
│  │  (Gin)       │  │  (protobuf)  │  │   (Nuxt 3)           │  │
│  └──────┬───────┘  └──────┬───────┘  └──────────┬───────────┘  │
│         │                 │                      │              │
│         └─────────────────┼──────────────────────┘              │
│                           ▼                                     │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                    Core Service Layer                      │ │
│  │  • Crash Processor (parsing, normalization)                │ │
│  │  • Crash Grouper (fingerprinting, deduplication)           │ │
│  │  • Alert Manager (webhooks, email, Slack)                  │ │
│  │  • Retention Manager (scheduled cleanup)                   │ │
│  └────────────────────────────────────────────────────────────┘ │
│                           │                                     │
│         ┌─────────────────┼─────────────────┐                  │
│         ▼                 ▼                 ▼                  │
│  ┌────────────┐  ┌──────────────┐  ┌────────────────┐         │
│  │  SQLite    │  │  File Store  │  │  Alert Queue   │         │
│  │  (index)   │  │  (raw logs)  │  │  (in-memory)   │         │
│  └────────────┘  └──────────────┘  └────────────────┘         │
└─────────────────────────────────────────────────────────────────┘
```

## Project Structure

```
inceptor/
├── cmd/
│   └── inceptor/
│       └── main.go                 # Application entry point
├── internal/
│   ├── api/
│   │   ├── rest/
│   │   │   ├── handler.go          # REST API handlers
│   │   │   ├── middleware.go       # Auth, CORS middleware
│   │   │   └── routes.go           # Route definitions
│   │   └── grpc/
│   │       └── server.go           # gRPC service implementation
│   ├── config/
│   │   └── config.go               # Configuration management (Viper)
│   ├── core/
│   │   ├── crash.go                # Data models (Crash, Group, App, Alert)
│   │   ├── grouper.go              # Fingerprinting algorithm
│   │   ├── alerter.go              # Alert system (webhook, email, Slack)
│   │   └── retention.go            # Retention policy & cleanup
│   └── storage/
│       ├── repository.go           # Storage interfaces
│       ├── sqlite.go               # SQLite implementation
│       └── filestore.go            # File-based log storage
├── api/
│   └── proto/
│       └── crash.proto             # gRPC service definitions
├── web/                            # Nuxt 3 Dashboard
│   ├── nuxt.config.ts
│   ├── app.vue
│   ├── layouts/
│   │   └── default.vue
│   ├── pages/
│   │   ├── index.vue               # Dashboard home
│   │   ├── crashes/
│   │   │   ├── index.vue           # Crash list
│   │   │   └── [id].vue            # Crash details
│   │   ├── groups/
│   │   │   └── index.vue           # Crash groups
│   │   ├── apps/
│   │   │   └── index.vue           # App management
│   │   └── settings.vue            # Settings & alerts
│   ├── composables/
│   │   └── useApi.ts               # API client composable
│   └── types/
│       └── index.ts                # TypeScript types
├── sdk/
│   └── flutter/                    # Flutter SDK
│       ├── lib/
│       │   ├── inceptor.dart       # Main export
│       │   └── src/
│       │       ├── inceptor.dart   # Core SDK implementation
│       │       ├── config.dart     # Configuration
│       │       └── models/
│       │           ├── crash_report.dart
│       │           ├── breadcrumb.dart
│       │           └── stack_frame.dart
│       ├── example/
│       │   └── lib/main.dart       # Example Flutter app
│       └── pubspec.yaml
├── configs/
│   └── config.example.yaml         # Example configuration
├── basepod.yaml                    # BasePod deployment config
├── Dockerfile                      # Multi-stage Docker build
├── Makefile                        # Build commands
├── go.mod
└── go.sum
```

## Configuration

### Environment Variables

All configuration can be set via environment variables with the `INCEPTOR_` prefix:

| Variable | Default | Description |
|----------|---------|-------------|
| `INCEPTOR_SERVER_REST_PORT` | `8080` | REST API port |
| `INCEPTOR_SERVER_GRPC_PORT` | `9090` | gRPC port |
| `INCEPTOR_SERVER_HOST` | `0.0.0.0` | Bind address |
| `INCEPTOR_STORAGE_SQLITE_PATH` | `./data/inceptor.db` | SQLite database path |
| `INCEPTOR_STORAGE_LOGS_PATH` | `./data/crashes` | Crash log files path |
| `INCEPTOR_RETENTION_DEFAULT_DAYS` | `30` | Default retention period |
| `INCEPTOR_RETENTION_CLEANUP_INTERVAL` | `24h` | Cleanup job interval |
| `INCEPTOR_AUTH_ENABLED` | `true` | Enable authentication |
| `INCEPTOR_AUTH_ADMIN_KEY` | `` | Admin API key |
| `INCEPTOR_ALERTS_SMTP_HOST` | `` | SMTP server host |
| `INCEPTOR_ALERTS_SLACK_WEBHOOK_URL` | `` | Slack webhook URL |

### Configuration File

```yaml
# configs/config.yaml
server:
  rest_port: 8080
  grpc_port: 9090
  host: "0.0.0.0"

storage:
  sqlite_path: "./data/inceptor.db"
  logs_path: "./data/crashes"

retention:
  default_days: 30
  cleanup_interval: "24h"

alerts:
  smtp:
    host: "smtp.example.com"
    port: 587
    username: ""
    password: ""
    from: "inceptor@example.com"
  slack:
    webhook_url: ""

auth:
  enabled: true
  admin_key: "your-secure-admin-key"
```

## API Reference

### Authentication

All API endpoints require the `X-API-Key` header:

```bash
curl -H "X-API-Key: your-api-key" https://your-server.com/api/v1/crashes
```

- **App API Key**: For crash submission and viewing app-specific data
- **Admin API Key**: For managing apps, alerts, and viewing all data

### Endpoints

#### Crashes

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/crashes` | Submit a crash report |
| `GET` | `/api/v1/crashes` | List crashes (with filters) |
| `GET` | `/api/v1/crashes/:id` | Get crash details |
| `DELETE` | `/api/v1/crashes/:id` | Delete a crash |

**Submit Crash Request:**

```json
{
  "app_version": "1.0.0+1",
  "platform": "android",
  "os_version": "Android 14",
  "device_model": "Pixel 8",
  "error_type": "FormatException",
  "error_message": "Invalid JSON format",
  "stack_trace": [
    {
      "file_name": "package:myapp/services/api.dart",
      "line_number": 42,
      "method_name": "parseResponse",
      "class_name": "ApiService"
    }
  ],
  "user_id": "user_123",
  "environment": "production",
  "metadata": {
    "screen": "checkout",
    "cart_items": 3
  },
  "breadcrumbs": [
    {
      "timestamp": "2024-01-15T10:30:00Z",
      "type": "navigation",
      "category": "navigation",
      "message": "Navigated to /checkout",
      "level": "info"
    }
  ]
}
```

**Response:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "group_id": "660e8400-e29b-41d4-a716-446655440001",
  "fingerprint": "a1b2c3d4e5f6g7h8",
  "is_new_group": false
}
```

#### Crash Groups

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/groups` | List crash groups |
| `GET` | `/api/v1/groups/:id` | Get group details |
| `PATCH` | `/api/v1/groups/:id` | Update group (status, notes) |

**Update Group:**

```json
{
  "status": "resolved",
  "assigned_to": "john@example.com",
  "notes": "Fixed in v1.0.1"
}
```

#### Apps (Admin Only)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/apps` | Create a new app |
| `GET` | `/api/v1/apps` | List all apps |
| `GET` | `/api/v1/apps/:id` | Get app details |
| `GET` | `/api/v1/apps/:id/stats` | Get app statistics |

**Create App:**

```json
{
  "name": "My Flutter App",
  "retention_days": 30
}
```

**Response (save the API key!):**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "My Flutter App",
  "api_key": "ink_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "created_at": "2024-01-15T10:00:00Z",
  "retention_days": 30
}
```

#### Alerts (Admin Only)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/alerts` | Create an alert |
| `GET` | `/api/v1/alerts` | List alerts |
| `DELETE` | `/api/v1/alerts/:id` | Delete an alert |

**Create Webhook Alert:**

```json
{
  "app_id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "webhook",
  "config": {
    "url": "https://example.com/webhook",
    "conditions": {
      "on_new_group": true
    }
  },
  "enabled": true
}
```

## Flutter SDK

### Installation

Add to your `pubspec.yaml`:

```yaml
dependencies:
  inceptor_flutter:
    git:
      url: https://github.com/flakerimi/inceptor.git
      path: sdk/flutter
```

### Basic Setup

```dart
import 'dart:async';
import 'package:flutter/material.dart';
import 'package:inceptor_flutter/inceptor.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize Inceptor
  await Inceptor.init(
    endpoint: 'https://your-server.com',
    apiKey: 'ink_your-api-key',
    environment: 'production',
    debug: false,
  );

  // Capture Flutter framework errors
  FlutterError.onError = Inceptor.recordFlutterError;

  // Capture async errors
  runZonedGuarded(() {
    runApp(MyApp());
  }, Inceptor.recordError);
}
```

### Manual Error Capture

```dart
try {
  await riskyOperation();
} catch (e, stackTrace) {
  await Inceptor.captureException(
    e,
    stackTrace: stackTrace,
    context: {
      'operation': 'checkout',
      'user_action': 'submit_order',
    },
  );
}
```

### Breadcrumbs

```dart
// Navigation tracking
Inceptor.addNavigationBreadcrumb(
  from: '/home',
  to: '/product/123',
);

// HTTP request tracking
Inceptor.addHttpBreadcrumb(
  method: 'POST',
  url: 'https://api.example.com/orders',
  statusCode: 201,
);

// User action tracking
Inceptor.addUserBreadcrumb(
  action: 'button_press',
  target: 'checkout_button',
  data: {'cart_total': 99.99},
);

// Custom breadcrumb
Inceptor.addBreadcrumb(
  InceptorBreadcrumb.custom(
    type: 'state',
    category: 'redux',
    message: 'Cart updated',
    data: {'items': 5},
  ),
);
```

### User Context

```dart
// Set user identifier
Inceptor.setUser('user_123');

// Add metadata
Inceptor.setMetadata('subscription', 'premium');
Inceptor.setMetadata('feature_flags', {'dark_mode': true});

// Clear on logout
Inceptor.setUser(null);
Inceptor.clearMetadata();
```

### Navigator Observer

```dart
MaterialApp(
  navigatorObservers: [InceptorNavigatorObserver()],
  // ...
);
```

## Dashboard

The web dashboard is built with Nuxt 3 and Nuxt UI.

### Development

```bash
cd web
npm install
npm run dev
```

Open http://localhost:3000 and enter your admin API key.

### Features

- **Dashboard**: Overview with crash statistics and trends
- **Crashes**: List, search, and filter crash reports
- **Groups**: Manage crash groups (resolve, ignore, assign)
- **Apps**: Register and manage applications
- **Settings**: Configure alerts and API access

### Production Build

```bash
cd web
npm run build
npm run preview
```

## Deployment

### BasePod

The project includes `basepod.yaml` for easy deployment:

```bash
# Login to your server
bp login your-server.example.com

# Deploy
bp push

# View logs
bp logs inceptor

# Manage
bp stop inceptor
bp start inceptor
bp restart inceptor
```

### Docker Compose

```yaml
version: '3.8'

services:
  inceptor:
    build: .
    ports:
      - "8080:8080"
      - "9090:9090"
    volumes:
      - inceptor-data:/app/data
    environment:
      - INCEPTOR_AUTH_ADMIN_KEY=your-secure-key
      - INCEPTOR_ALERTS_SLACK_WEBHOOK_URL=https://hooks.slack.com/...
    restart: unless-stopped

volumes:
  inceptor-data:
```

### Systemd

```ini
# /etc/systemd/system/inceptor.service
[Unit]
Description=Inceptor Crash Logging Service
After=network.target

[Service]
Type=simple
User=inceptor
WorkingDirectory=/opt/inceptor
ExecStart=/opt/inceptor/bin/inceptor
Restart=always
RestartSec=5
Environment=INCEPTOR_AUTH_ADMIN_KEY=your-secure-key

[Install]
WantedBy=multi-user.target
```

## Development

### Build Commands

```bash
# Install dependencies
make deps

# Build binary
make build

# Run locally
make run

# Run tests
make test

# Build for all platforms
make build-all

# Generate gRPC code (requires protoc)
make proto
```

### Testing

```bash
# Run all tests
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Data Models

### Crash

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique crash ID |
| `app_id` | string | Associated app ID |
| `app_version` | string | App version string |
| `platform` | string | Platform (ios, android, web, flutter) |
| `os_version` | string | OS version |
| `device_model` | string | Device model |
| `error_type` | string | Exception/error type |
| `error_message` | string | Error message |
| `stack_trace` | array | Stack frames |
| `fingerprint` | string | Grouping fingerprint |
| `group_id` | string | Associated group ID |
| `user_id` | string | Optional user identifier |
| `environment` | string | Environment (production, staging, dev) |
| `created_at` | datetime | Timestamp |
| `metadata` | object | Custom metadata |
| `breadcrumbs` | array | Events leading to crash |

### Crash Group

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique group ID |
| `app_id` | string | Associated app ID |
| `fingerprint` | string | Unique fingerprint |
| `error_type` | string | Common error type |
| `error_message` | string | Representative message |
| `first_seen` | datetime | First occurrence |
| `last_seen` | datetime | Most recent occurrence |
| `occurrence_count` | int | Total occurrences |
| `status` | string | open, resolved, ignored |
| `assigned_to` | string | Assignee email |
| `notes` | string | Notes/comments |

## Crash Fingerprinting

Inceptor uses a fingerprinting algorithm to group similar crashes:

1. **Error Type**: The exception/error class name
2. **Top Stack Frames**: First 5 non-native frames (normalized)
3. **Normalization**: Removes line numbers, closure IDs, and build hashes

This ensures crashes from the same code path are grouped together, even across different app versions.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
