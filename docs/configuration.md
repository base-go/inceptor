# Configuration Reference

Inceptor can be configured via YAML file, environment variables, or command-line flags.

## Configuration Methods

### Priority Order (highest to lowest)

1. Command-line flags
2. Environment variables
3. Configuration file
4. Default values

### Configuration File

By default, Inceptor looks for config in:
- `./configs/config.yaml`
- `/app/configs/config.yaml`

Specify a custom path:
```bash
./inceptor --config /path/to/config.yaml
```

### Environment Variables

All configuration can be set via environment variables with the prefix `INCEPTOR_`:

| Config Path | Environment Variable |
|-------------|---------------------|
| `server.rest_port` | `INCEPTOR_SERVER_REST_PORT` |
| `auth.admin_key` | `INCEPTOR_AUTH_ADMIN_KEY` |
| `storage.sqlite_path` | `INCEPTOR_STORAGE_SQLITE_PATH` |

---

## Full Configuration Reference

```yaml
# Server configuration
server:
  # Host to bind to (0.0.0.0 for all interfaces)
  host: "0.0.0.0"

  # REST API port
  rest_port: 8080

  # gRPC port (optional)
  grpc_port: 9090

  # Dashboard port (if serving separately)
  dashboard_port: 3000

# Storage configuration
storage:
  # Path to SQLite database file
  sqlite_path: "./data/inceptor.db"

  # Path to store crash log files
  logs_path: "./data/crashes"

# Data retention configuration
retention:
  # Default retention period in days
  # Crashes older than this are automatically deleted
  default_days: 30

  # How often to run cleanup (Go duration format)
  # Examples: "1h", "24h", "7d"
  cleanup_interval: "24h"

# Alert configuration
alerts:
  # SMTP configuration for email alerts
  smtp:
    host: "smtp.example.com"
    port: 587
    username: ""
    password: ""
    from: "inceptor@example.com"

  # Slack webhook for notifications
  slack:
    webhook_url: ""

# Authentication configuration
auth:
  # Enable/disable authentication
  # WARNING: Disabling authentication is not recommended
  enabled: true

  # Admin API key for full system access
  # Generate with: openssl rand -hex 32
  admin_key: "your-secure-admin-key-here"
```

---

## Configuration Details

### Server Settings

#### `server.host`

| Property | Value |
|----------|-------|
| Type | string |
| Default | `"0.0.0.0"` |
| Environment | `INCEPTOR_SERVER_HOST` |

The network interface to bind to:
- `"0.0.0.0"` - Listen on all interfaces (recommended for containers)
- `"127.0.0.1"` - Listen only on localhost
- `"192.168.1.100"` - Listen on specific IP

#### `server.rest_port`

| Property | Value |
|----------|-------|
| Type | integer |
| Default | `8080` |
| Environment | `INCEPTOR_SERVER_REST_PORT` |

Port for the REST API and web dashboard.

#### `server.grpc_port`

| Property | Value |
|----------|-------|
| Type | integer |
| Default | `9090` |
| Environment | `INCEPTOR_SERVER_GRPC_PORT` |

Port for gRPC service (if enabled).

---

### Storage Settings

#### `storage.sqlite_path`

| Property | Value |
|----------|-------|
| Type | string |
| Default | `"./data/inceptor.db"` |
| Environment | `INCEPTOR_STORAGE_SQLITE_PATH` |

Path to the SQLite database file. The directory must exist and be writable.

**Important**: For containerized deployments, use an absolute path and mount a persistent volume.

#### `storage.logs_path`

| Property | Value |
|----------|-------|
| Type | string |
| Default | `"./data/crashes"` |
| Environment | `INCEPTOR_STORAGE_LOGS_PATH` |

Directory for storing full crash log JSON files.

Structure:
```
{logs_path}/
└── {app_id}/
    └── {YYYY-MM-DD}/
        └── {crash_id}.json
```

---

### Retention Settings

#### `retention.default_days`

| Property | Value |
|----------|-------|
| Type | integer |
| Default | `30` |
| Environment | `INCEPTOR_RETENTION_DEFAULT_DAYS` |

Default number of days to keep crash data. Can be overridden per-app when creating the app.

**Per-App Override**: When creating an app, specify `retention_days`:
```json
{
  "name": "My App",
  "retention_days": 90
}
```

#### `retention.cleanup_interval`

| Property | Value |
|----------|-------|
| Type | duration |
| Default | `"24h"` |
| Environment | `INCEPTOR_RETENTION_CLEANUP_INTERVAL` |

How often the retention manager runs cleanup. Uses Go duration format:
- `"1h"` - Every hour
- `"24h"` - Every day
- `"168h"` - Every week

---

### Alert Settings

#### SMTP Configuration

For email alerts:

```yaml
alerts:
  smtp:
    host: "smtp.gmail.com"      # SMTP server hostname
    port: 587                    # SMTP port (usually 587 for TLS)
    username: "your@email.com"  # SMTP username
    password: "app-password"     # SMTP password or app password
    from: "alerts@example.com"  # From address for sent emails
```

| Setting | Environment Variable |
|---------|---------------------|
| `host` | `INCEPTOR_ALERTS_SMTP_HOST` |
| `port` | `INCEPTOR_ALERTS_SMTP_PORT` |
| `username` | `INCEPTOR_ALERTS_SMTP_USERNAME` |
| `password` | `INCEPTOR_ALERTS_SMTP_PASSWORD` |
| `from` | `INCEPTOR_ALERTS_SMTP_FROM` |

#### Slack Configuration

```yaml
alerts:
  slack:
    webhook_url: "https://hooks.slack.com/services/T00/B00/xxxx"
```

| Setting | Environment Variable |
|---------|---------------------|
| `webhook_url` | `INCEPTOR_ALERTS_SLACK_WEBHOOK_URL` |

---

### Authentication Settings

#### `auth.enabled`

| Property | Value |
|----------|-------|
| Type | boolean |
| Default | `true` |
| Environment | `INCEPTOR_AUTH_ENABLED` |

Whether to require API key authentication.

**WARNING**: Setting this to `false` allows anyone to:
- Submit crashes
- View all crash data
- Create/delete apps and alerts

Only disable for development/testing.

#### `auth.admin_key`

| Property | Value |
|----------|-------|
| Type | string |
| Default | `""` (must be set) |
| Environment | `INCEPTOR_AUTH_ADMIN_KEY` |

The admin API key for full system access. Used to:
- Create and manage apps
- Configure alerts
- Access all crash data
- Access the dashboard

**Generate a secure key**:
```bash
openssl rand -hex 32
```

---

## Example Configurations

### Development

```yaml
server:
  host: "127.0.0.1"
  rest_port: 8080

storage:
  sqlite_path: "./data/dev.db"
  logs_path: "./data/crashes"

retention:
  default_days: 7
  cleanup_interval: "1h"

auth:
  enabled: false  # Only for local dev!
```

### Production (Docker)

```yaml
server:
  host: "0.0.0.0"
  rest_port: 8080

storage:
  sqlite_path: "/app/data/inceptor.db"
  logs_path: "/app/data/crashes"

retention:
  default_days: 90
  cleanup_interval: "24h"

alerts:
  smtp:
    host: "smtp.sendgrid.net"
    port: 587
    username: "apikey"
    password: "${SENDGRID_API_KEY}"
    from: "crashes@yourcompany.com"
  slack:
    webhook_url: "${SLACK_WEBHOOK_URL}"

auth:
  enabled: true
  admin_key: "${ADMIN_API_KEY}"
```

### Environment Variables Only

For containerized deployments, you can skip the config file entirely:

```bash
docker run -d \
  -e INCEPTOR_SERVER_HOST=0.0.0.0 \
  -e INCEPTOR_SERVER_REST_PORT=8080 \
  -e INCEPTOR_STORAGE_SQLITE_PATH=/app/data/inceptor.db \
  -e INCEPTOR_STORAGE_LOGS_PATH=/app/data/crashes \
  -e INCEPTOR_RETENTION_DEFAULT_DAYS=30 \
  -e INCEPTOR_AUTH_ENABLED=true \
  -e INCEPTOR_AUTH_ADMIN_KEY=your-secure-key \
  -v inceptor-data:/app/data \
  inceptor:latest
```

---

## Validation

On startup, Inceptor validates configuration:

| Check | Error If |
|-------|----------|
| `auth.admin_key` | Empty when `auth.enabled` is true |
| `storage.sqlite_path` | Directory doesn't exist or isn't writable |
| `storage.logs_path` | Directory doesn't exist or isn't writable |
| `server.rest_port` | Port is already in use |

Check the startup logs for any configuration warnings or errors.
