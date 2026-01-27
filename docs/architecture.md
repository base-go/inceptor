# Architecture

This document describes Inceptor's system architecture and component design.

## System Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                           Inceptor                                  │
├─────────────────────────────────────────────────────────────────────┤
│  ┌───────────────┐  ┌───────────────┐  ┌─────────────────────────┐ │
│  │   REST API    │  │   gRPC API    │  │     Web Dashboard       │ │
│  │  (Gin HTTP)   │  │  (protobuf)   │  │  (Embedded Nuxt SPA)    │ │
│  └───────┬───────┘  └───────┬───────┘  └───────────┬─────────────┘ │
│          │                  │                      │               │
│          └──────────────────┼──────────────────────┘               │
│                             ▼                                      │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │                   Core Service Layer                         │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │  │
│  │  │   Grouper   │  │   Alerter   │  │ Retention Manager   │  │  │
│  │  │ (fingerprint)│ │ (notify)    │  │ (cleanup)           │  │  │
│  │  └─────────────┘  └─────────────┘  └─────────────────────┘  │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                             │                                      │
│          ┌──────────────────┼──────────────────┐                  │
│          ▼                  ▼                  ▼                  │
│  ┌─────────────┐   ┌──────────────┐   ┌────────────────┐         │
│  │   SQLite    │   │  File Store  │   │  Alert Queue   │         │
│  │  (metadata) │   │  (raw logs)  │   │  (in-memory)   │         │
│  └─────────────┘   └──────────────┘   └────────────────┘         │
└─────────────────────────────────────────────────────────────────────┘
```

## Components

### API Layer

**REST API** (`internal/api/rest/`)

The primary API interface built with Gin framework:
- Handles crash submissions from mobile/web clients
- Provides CRUD operations for crashes, groups, apps, and alerts
- Serves the embedded web dashboard
- Implements API key authentication middleware

**gRPC API** (`internal/api/grpc/`)

Optional high-performance interface for:
- High-volume crash submission
- Streaming multiple crashes in a single connection
- Internal service-to-service communication

### Core Services

**Grouper** (`internal/core/grouper.go`)

Responsible for crash fingerprinting and grouping:

```go
type Grouper struct {
    FrameLimit int  // Number of frames to use (default: 5)
}

func (g *Grouper) GenerateFingerprint(crash *Crash) string
```

The fingerprinting algorithm:
1. Takes the error type (e.g., `FormatException`)
2. Extracts the top N stack frames (default: 5)
3. Normalizes each frame:
   - Removes line numbers (they change between builds)
   - Removes closure/lambda IDs
   - Strips generic type parameters
   - Extracts just the filename (no path)
   - Removes build hashes from filenames
4. Creates SHA256 hash of combined normalized data
5. Returns first 16 characters as fingerprint

This ensures similar crashes (same error type, same code path) are grouped together even if they occur on different lines or in different builds.

**Alert Manager** (`internal/core/alerter.go`)

Handles notifications for crash events:

```go
type AlertManager struct {
    alerts    []*core.Alert
    smtp      SMTPConfig
    slackURL  string
    queue     chan AlertEvent
}
```

Features:
- Webhook notifications (POST to configured URL)
- Email via SMTP
- Slack via incoming webhook
- Async processing via channel-based queue
- Configurable per-app alert rules

**Retention Manager** (`internal/core/retention.go`)

Enforces data retention policies:

```go
type RetentionManager struct {
    repo            storage.Repository
    fileStore       storage.FileStore
    defaultDays     int
    cleanupInterval time.Duration
}
```

- Runs as a background goroutine
- Checks each app's retention policy
- Deletes crashes older than retention period
- Cleans up both database records and log files
- Configurable cleanup interval (default: 24h)

### Storage Layer

**Repository Interface** (`internal/storage/repository.go`)

Defines the contract for all storage operations:

```go
type Repository interface {
    // Crashes
    CreateCrash(ctx context.Context, crash *core.Crash) error
    GetCrash(ctx context.Context, id string) (*core.Crash, error)
    ListCrashes(ctx context.Context, filter CrashFilter) ([]*core.Crash, int, error)
    DeleteCrash(ctx context.Context, id string) error
    DeleteCrashesOlderThan(ctx context.Context, appID string, before time.Time) (int, error)

    // Groups
    GetOrCreateGroup(ctx context.Context, crash *core.Crash) (*core.CrashGroup, bool, error)
    // ... more methods
}
```

**SQLite Implementation** (`internal/storage/sqlite.go`)

Uses `modernc.org/sqlite` (pure Go, no CGO) for:
- Crash metadata and indexes
- Crash groups
- App configurations
- Alert rules
- Settings

Schema highlights:

```sql
-- Crashes indexed by app, group, and creation time
CREATE TABLE crashes (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    group_id TEXT,
    created_at DATETIME,
    -- ... more fields
);
CREATE INDEX idx_crashes_app_id ON crashes(app_id);
CREATE INDEX idx_crashes_group_id ON crashes(group_id);
CREATE INDEX idx_crashes_created_at ON crashes(created_at);

-- Groups for crash aggregation
CREATE TABLE crash_groups (
    id TEXT PRIMARY KEY,
    fingerprint TEXT UNIQUE NOT NULL,
    occurrence_count INTEGER DEFAULT 1,
    status TEXT DEFAULT 'open',
    -- ...
);
```

**File Store** (`internal/storage/filestore.go`)

Stores full crash payloads as JSON files:

```
data/crashes/
└── {app_id}/
    └── {YYYY-MM-DD}/
        └── {crash_id}.json
```

Benefits:
- Full crash details preserved (including large stack traces)
- Easy to backup/archive
- No database bloat
- Date-based directory structure for efficient cleanup

### Web Dashboard

Built with Nuxt 3 and Nuxt UI:

```
web/
├── pages/
│   ├── index.vue          # Dashboard home
│   ├── crashes/
│   │   ├── index.vue      # Crash list
│   │   └── [id].vue       # Crash detail
│   ├── groups/
│   │   └── index.vue      # Group management
│   ├── apps/
│   │   └── index.vue      # App management
│   └── settings.vue       # Settings & alerts
├── composables/
│   └── useApi.ts          # API client composable
└── layouts/
    └── default.vue        # App shell
```

**Build Process**:

1. `npm run generate` produces static files in `web/.output/public/`
2. Static files are copied to `internal/api/rest/static/`
3. Go embeds the files using `embed.FS`
4. REST server serves them at the root path

## Data Flow

### Crash Submission Flow

```
Client App                    Inceptor Server
    │                              │
    │  POST /api/v1/crashes        │
    │  X-API-Key: ink_...          │
    │  {crash data}                │
    │─────────────────────────────>│
    │                              │
    │                    ┌─────────┴─────────┐
    │                    │  Auth Middleware  │
    │                    │  Validate API Key │
    │                    └─────────┬─────────┘
    │                              │
    │                    ┌─────────┴─────────┐
    │                    │     Handler       │
    │                    │  Parse request    │
    │                    └─────────┬─────────┘
    │                              │
    │                    ┌─────────┴─────────┐
    │                    │     Grouper       │
    │                    │ Generate fingerprint
    │                    └─────────┬─────────┘
    │                              │
    │                    ┌─────────┴─────────┐
    │                    │   Repository      │
    │                    │ Get/Create group  │
    │                    │ Save crash        │
    │                    └─────────┬─────────┘
    │                              │
    │                    ┌─────────┴─────────┐
    │                    │   File Store      │
    │                    │ Save full log     │
    │                    └─────────┬─────────┘
    │                              │
    │                    ┌─────────┴─────────┐
    │                    │   Alert Manager   │
    │                    │ Queue notification│
    │                    └─────────┬─────────┘
    │                              │
    │  {id, group_id, fingerprint} │
    │<─────────────────────────────│
    │                              │
```

### Retention Cleanup Flow

```
┌──────────────────────────────────────────────┐
│           Retention Manager                   │
│  (background goroutine, runs every 24h)      │
└──────────────────────┬───────────────────────┘
                       │
        ┌──────────────┴──────────────┐
        │  For each app:               │
        │  1. Get retention_days       │
        │  2. Calculate cutoff date    │
        └──────────────┬──────────────┘
                       │
        ┌──────────────┴──────────────┐
        │  Repository.DeleteOlderThan │
        │  Delete crashes from SQLite │
        └──────────────┬──────────────┘
                       │
        ┌──────────────┴──────────────┐
        │  FileStore.DeleteOldLogs    │
        │  Delete JSON files          │
        └─────────────────────────────┘
```

## Authentication Model

Inceptor uses API key authentication:

```
┌──────────────────────────────────────────────────────────────┐
│                    Authentication Flow                        │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Request                    Middleware                       │
│     │                          │                             │
│     │  X-API-Key: ink_...      │                             │
│     │─────────────────────────>│                             │
│     │                          │                             │
│     │                    Hash the key                        │
│     │                          │                             │
│     │                    Lookup in DB                        │
│     │                    (apps.api_key_hash)                 │
│     │                          │                             │
│     │                    ┌─────┴─────┐                       │
│     │                    │           │                       │
│     │              App Found    Admin Key?                   │
│     │                    │           │                       │
│     │              Set c.Set    Set c.Set                    │
│     │              ("app",     ("is_admin",                  │
│     │               app)        true)                        │
│     │                    │           │                       │
│     │                    └─────┬─────┘                       │
│     │                          │                             │
│     │                    Continue to                         │
│     │                    Handler                             │
│     │                          │                             │
└──────────────────────────────────────────────────────────────┘
```

## Deployment Architecture

### Single Instance (BasePod)

```
┌─────────────────────────────────────────┐
│              BasePod Host               │
│  ┌───────────────────────────────────┐  │
│  │         Inceptor Container        │  │
│  │  ┌──────────┐  ┌──────────────┐  │  │
│  │  │  Go App  │  │   SQLite DB  │  │  │
│  │  │  :8080   │  │  /app/data/  │  │  │
│  │  └──────────┘  └──────────────┘  │  │
│  │        │              │          │  │
│  │        └──────┬───────┘          │  │
│  │               │                  │  │
│  │       Persistent Volume          │  │
│  │       /app/data/                 │  │
│  └───────────────────────────────────┘  │
│                  │                      │
│                  ▼                      │
│            Load Balancer                │
│            (BasePod managed)            │
└─────────────────────────────────────────┘
```

### Production Recommendations

For production deployments:

1. **Backup**: Regularly backup `/app/data/` directory
2. **Monitoring**: Use the `/health` endpoint for health checks
3. **Reverse Proxy**: Add nginx/Caddy for:
   - TLS termination
   - Rate limiting
   - Request logging
4. **Alerts**: Configure webhook/email alerts for new crash groups
5. **Retention**: Set appropriate retention periods per app

## Technology Stack

| Component | Technology | Rationale |
|-----------|------------|-----------|
| Language | Go 1.22 | Performance, static binary, concurrency |
| HTTP | Gin | Fast, well-documented, middleware support |
| Database | SQLite (modernc.org) | Pure Go, no CGO, embedded |
| Config | Viper | Flexible, env vars + file support |
| Logging | zerolog | Fast, structured JSON logs |
| Frontend | Nuxt 3 + Nuxt UI | Vue-based, SSG, beautiful components |
| SDK | Dart/Flutter | Native Flutter integration |
