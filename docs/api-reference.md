# API Reference

Inceptor provides a REST API for crash submission, retrieval, and management.

## Authentication

All API requests require authentication via the `X-API-Key` header:

```bash
curl -H "X-API-Key: your-api-key" https://your-server.com/api/v1/crashes
```

There are two types of API keys:

| Type | Purpose | Scope |
|------|---------|-------|
| **App API Key** | Submit crashes for a specific app | Crash submission, view own app's data |
| **Admin API Key** | Full system access | Manage apps, alerts, view all data |

## Base URL

All API endpoints are prefixed with `/api/v1`.

---

## Health Check

### GET /health

Check if the server is running.

**Authentication**: None required

**Response**:
```json
{
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

---

## Crashes

### POST /api/v1/crashes

Submit a crash report.

**Authentication**: App API Key

**Request Body**:
```json
{
  "app_version": "1.0.0",
  "platform": "flutter",
  "os_version": "Android 14",
  "device_model": "Pixel 8",
  "error_type": "FormatException",
  "error_message": "Invalid date format in input",
  "stack_trace": [
    {
      "file_name": "date_parser.dart",
      "line_number": 42,
      "column_number": 8,
      "method_name": "parse",
      "class_name": "DateParser",
      "native": false
    }
  ],
  "user_id": "user-123",
  "environment": "production",
  "metadata": {
    "screen": "checkout",
    "cart_items": 3
  },
  "breadcrumbs": [
    {
      "timestamp": "2024-01-15T10:29:55Z",
      "type": "navigation",
      "category": "navigation",
      "message": "Navigated to /checkout",
      "level": "info"
    }
  ]
}
```

**Required Fields**:
- `app_version` - Application version string
- `platform` - Platform identifier (flutter, ios, android, web)
- `error_type` - Exception/error class name
- `error_message` - Error description
- `stack_trace` - Array of stack frames

**Response** (201 Created):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "group_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "fingerprint": "a1b2c3d4e5f6g7h8",
  "is_new_group": true
}
```

---

### GET /api/v1/crashes

List crashes with optional filters.

**Authentication**: App API Key (own app) or Admin API Key (all apps)

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `app_id` | string | Filter by app ID |
| `group_id` | string | Filter by crash group |
| `platform` | string | Filter by platform |
| `environment` | string | Filter by environment |
| `error_type` | string | Filter by error type |
| `user_id` | string | Filter by user ID |
| `search` | string | Search in error message |
| `from` | datetime | Start date (RFC3339) |
| `to` | datetime | End date (RFC3339) |
| `limit` | int | Max results (default: 50) |
| `offset` | int | Pagination offset |

**Response**:
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "app_id": "app-123",
      "app_version": "1.0.0",
      "platform": "flutter",
      "error_type": "FormatException",
      "error_message": "Invalid date format",
      "fingerprint": "a1b2c3d4e5f6g7h8",
      "group_id": "group-456",
      "environment": "production",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 150,
  "limit": 50,
  "offset": 0
}
```

---

### GET /api/v1/crashes/:id

Get a single crash with full details.

**Authentication**: App API Key (own app) or Admin API Key

**Response**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "app_id": "app-123",
  "app_version": "1.0.0",
  "platform": "flutter",
  "os_version": "Android 14",
  "device_model": "Pixel 8",
  "error_type": "FormatException",
  "error_message": "Invalid date format in input",
  "stack_trace": [...],
  "fingerprint": "a1b2c3d4e5f6g7h8",
  "group_id": "group-456",
  "user_id": "user-123",
  "environment": "production",
  "created_at": "2024-01-15T10:30:00Z",
  "metadata": {...},
  "breadcrumbs": [...]
}
```

---

### DELETE /api/v1/crashes/:id

Delete a crash.

**Authentication**: App API Key (own app) or Admin API Key

**Response** (200 OK):
```json
{
  "message": "Crash deleted"
}
```

---

## Crash Groups

### GET /api/v1/groups

List crash groups.

**Authentication**: App API Key (own app) or Admin API Key

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `app_id` | string | Filter by app ID |
| `status` | string | Filter by status (open, resolved, ignored) |
| `error_type` | string | Filter by error type |
| `search` | string | Search in error message |
| `sort_by` | string | Sort field (last_seen, first_seen, occurrence_count) |
| `sort_order` | string | Sort direction (asc, desc) |
| `limit` | int | Max results (default: 50) |
| `offset` | int | Pagination offset |

**Response**:
```json
{
  "data": [
    {
      "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
      "app_id": "app-123",
      "fingerprint": "a1b2c3d4e5f6g7h8",
      "error_type": "FormatException",
      "error_message": "Invalid date format",
      "first_seen": "2024-01-10T08:00:00Z",
      "last_seen": "2024-01-15T10:30:00Z",
      "occurrence_count": 47,
      "status": "open"
    }
  ],
  "total": 25,
  "limit": 50,
  "offset": 0
}
```

---

### GET /api/v1/groups/:id

Get a single crash group.

**Authentication**: App API Key (own app) or Admin API Key

---

### PATCH /api/v1/groups/:id

Update a crash group (change status, assign, add notes).

**Authentication**: App API Key (own app) or Admin API Key

**Request Body**:
```json
{
  "status": "resolved",
  "assigned_to": "developer@example.com",
  "notes": "Fixed in v1.0.1"
}
```

**Response**: Updated group object

---

## Apps (Admin Only)

### POST /api/v1/apps

Register a new application.

**Authentication**: Admin API Key

**Request Body**:
```json
{
  "name": "My Flutter App",
  "retention_days": 30
}
```

**Response** (201 Created):
```json
{
  "id": "app-123",
  "name": "My Flutter App",
  "api_key": "ink_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "created_at": "2024-01-15T10:00:00Z",
  "retention_days": 30
}
```

**Note**: The `api_key` is only returned on creation. Store it securely!

---

### GET /api/v1/apps

List all registered applications.

**Authentication**: Admin API Key

---

### GET /api/v1/apps/:id

Get application details.

**Authentication**: Admin API Key

---

### GET /api/v1/apps/:id/stats

Get crash statistics for an application.

**Authentication**: App API Key (own app) or Admin API Key

**Response**:
```json
{
  "app_id": "app-123",
  "total_crashes": 1250,
  "total_groups": 45,
  "open_groups": 12,
  "crashes_last_24h": 23,
  "crashes_last_7d": 156,
  "crashes_last_30d": 487,
  "top_errors": [
    {
      "group_id": "group-1",
      "error_type": "FormatException",
      "error_message": "Invalid date format",
      "count": 89
    }
  ],
  "crash_trend": [
    {"date": "2024-01-14", "count": 45},
    {"date": "2024-01-15", "count": 23}
  ]
}
```

---

## Alerts (Admin Only)

### POST /api/v1/alerts

Create an alert rule.

**Authentication**: Admin API Key

**Request Body**:
```json
{
  "app_id": "app-123",
  "type": "webhook",
  "config": {
    "url": "https://your-webhook.com/endpoint",
    "headers": {
      "Authorization": "Bearer token"
    }
  },
  "enabled": true
}
```

Alert types and their config:

**Webhook**:
```json
{
  "type": "webhook",
  "config": {
    "url": "https://...",
    "headers": {}
  }
}
```

**Email**:
```json
{
  "type": "email",
  "config": {
    "to": ["dev@example.com"],
    "subject_prefix": "[Crash Alert]"
  }
}
```

**Slack**:
```json
{
  "type": "slack",
  "config": {
    "channel": "#alerts"
  }
}
```

---

### GET /api/v1/alerts

List alerts.

**Authentication**: Admin API Key

**Query Parameters**:
- `app_id` - Filter by app ID

---

### DELETE /api/v1/alerts/:id

Delete an alert.

**Authentication**: Admin API Key

---

## Error Responses

All errors follow this format:

```json
{
  "error": "Error description",
  "details": "Additional information (optional)"
}
```

**Common HTTP Status Codes**:
| Code | Meaning |
|------|---------|
| 400 | Bad Request - Invalid request body |
| 401 | Unauthorized - Invalid or missing API key |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource doesn't exist |
| 500 | Internal Server Error |

---

## Rate Limiting

Inceptor does not enforce rate limits by default. Consider adding a reverse proxy (nginx, Caddy) for production rate limiting.

## Data Types

### Stack Frame
```typescript
{
  file_name: string;
  line_number: number;
  column_number?: number;
  method_name: string;
  class_name?: string;
  native?: boolean;
}
```

### Breadcrumb
```typescript
{
  timestamp: string; // RFC3339
  type: "navigation" | "http" | "user" | "log";
  category: string;
  message: string;
  data?: Record<string, any>;
  level: "debug" | "info" | "warning" | "error";
}
```
