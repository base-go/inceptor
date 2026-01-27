# Alerting Guide

Inceptor supports multiple notification channels to keep you informed about crashes in your applications.

## Overview

Alerts are triggered when:
- A new crash group is created (first occurrence of an error)
- A crash is added to an existing group

Each alert is configured per-app, allowing different notification settings for different applications.

## Alert Types

### Webhook

Send HTTP POST requests to your own endpoint.

**Configuration**:
```json
{
  "app_id": "your-app-id",
  "type": "webhook",
  "config": {
    "url": "https://your-server.com/crash-webhook",
    "headers": {
      "Authorization": "Bearer your-token",
      "X-Custom-Header": "value"
    }
  },
  "enabled": true
}
```

**Payload sent to webhook**:
```json
{
  "event_type": "new_group",
  "timestamp": "2024-01-15T10:30:00Z",
  "app": {
    "id": "app-123",
    "name": "My App"
  },
  "crash": {
    "id": "crash-456",
    "error_type": "FormatException",
    "error_message": "Invalid date format",
    "platform": "flutter",
    "app_version": "1.0.0",
    "environment": "production"
  },
  "group": {
    "id": "group-789",
    "fingerprint": "a1b2c3d4",
    "occurrence_count": 1,
    "first_seen": "2024-01-15T10:30:00Z"
  },
  "is_new_group": true
}
```

**Event Types**:
- `new_group` - First occurrence of a crash pattern
- `new_crash` - New crash in existing group

### Email

Send email notifications via SMTP.

**Server Configuration** (in `config.yaml`):
```yaml
alerts:
  smtp:
    host: "smtp.gmail.com"
    port: 587
    username: "your-email@gmail.com"
    password: "your-app-password"
    from: "crashes@yourapp.com"
```

**Alert Configuration**:
```json
{
  "app_id": "your-app-id",
  "type": "email",
  "config": {
    "to": ["dev-team@example.com", "oncall@example.com"],
    "subject_prefix": "[Crash Alert]"
  },
  "enabled": true
}
```

**Email Format**:
```
Subject: [Crash Alert] New FormatException in My App

New crash group detected in My App (production)

Error Type: FormatException
Message: Invalid date format

Platform: flutter
App Version: 1.0.0
First Seen: 2024-01-15 10:30:00 UTC

View in Dashboard: https://your-server.com/groups/group-789
```

### Slack

Send notifications to Slack channels.

**Server Configuration** (in `config.yaml`):
```yaml
alerts:
  slack:
    webhook_url: "https://hooks.slack.com/services/T00/B00/xxxx"
```

**Alert Configuration**:
```json
{
  "app_id": "your-app-id",
  "type": "slack",
  "config": {
    "channel": "#crash-alerts"
  },
  "enabled": true
}
```

**Slack Message**:
```
🚨 New Crash in My App

*FormatException*
> Invalid date format

Platform: flutter | Version: 1.0.0 | Env: production

<https://your-server.com/groups/group-789|View in Dashboard>
```

## Creating Alerts

### Via API

```bash
curl -X POST http://localhost:8080/api/v1/alerts \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-admin-key" \
  -d '{
    "app_id": "your-app-id",
    "type": "webhook",
    "config": {
      "url": "https://your-webhook.com/endpoint"
    },
    "enabled": true
  }'
```

### Via Dashboard

1. Navigate to **Settings** → **Alerts**
2. Click **Add Alert**
3. Select the app and alert type
4. Configure the alert settings
5. Click **Save**

## Managing Alerts

### List Alerts

```bash
# All alerts
curl http://localhost:8080/api/v1/alerts \
  -H "X-API-Key: your-admin-key"

# Filter by app
curl "http://localhost:8080/api/v1/alerts?app_id=your-app-id" \
  -H "X-API-Key: your-admin-key"
```

### Delete Alert

```bash
curl -X DELETE http://localhost:8080/api/v1/alerts/alert-id \
  -H "X-API-Key: your-admin-key"
```

### Enable/Disable Alert

Alerts can be temporarily disabled without deleting them:

```bash
curl -X PATCH http://localhost:8080/api/v1/alerts/alert-id \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-admin-key" \
  -d '{"enabled": false}'
```

## Integration Examples

### Discord (via Webhook)

Discord accepts Slack-compatible webhooks:

```json
{
  "type": "webhook",
  "config": {
    "url": "https://discord.com/api/webhooks/xxx/yyy/slack",
    "headers": {}
  }
}
```

### PagerDuty

Use PagerDuty's Events API v2:

```json
{
  "type": "webhook",
  "config": {
    "url": "https://events.pagerduty.com/v2/enqueue",
    "headers": {
      "Content-Type": "application/json"
    }
  }
}
```

Then create a middleware service to transform the payload to PagerDuty format.

### Microsoft Teams

Teams requires a specific payload format. Use a webhook relay or create a custom endpoint:

```json
{
  "type": "webhook",
  "config": {
    "url": "https://your-relay.com/teams-webhook",
    "headers": {}
  }
}
```

### Opsgenie

```json
{
  "type": "webhook",
  "config": {
    "url": "https://api.opsgenie.com/v2/alerts",
    "headers": {
      "Authorization": "GenieKey your-api-key"
    }
  }
}
```

## Alert Payload Reference

### Full Webhook Payload

```typescript
interface AlertPayload {
  event_type: "new_group" | "new_crash";
  timestamp: string; // ISO 8601

  app: {
    id: string;
    name: string;
  };

  crash: {
    id: string;
    error_type: string;
    error_message: string;
    platform: string;
    app_version: string;
    os_version?: string;
    device_model?: string;
    environment: string;
    user_id?: string;
    created_at: string;
  };

  group: {
    id: string;
    fingerprint: string;
    occurrence_count: number;
    first_seen: string;
    last_seen: string;
    status: "open" | "resolved" | "ignored";
  };

  is_new_group: boolean;
}
```

## Best Practices

### 1. Don't Alert on Everything

For high-volume apps, alerting on every crash can cause alert fatigue. Consider:
- Only alerting on new crash groups
- Using webhook with your own filtering logic
- Setting up escalation rules in your incident management system

### 2. Use Different Channels for Different Apps

```
Production apps → PagerDuty + Slack #incidents
Staging apps → Slack #staging-alerts only
Dev apps → No alerts (check dashboard manually)
```

### 3. Include Context in Webhooks

When building a webhook receiver, extract useful context:
- `is_new_group: true` → Might be a new bug, investigate immediately
- `occurrence_count > 100` → High-impact issue
- `environment: production` → Needs immediate attention

### 4. Test Your Alerts

After setting up alerts, submit a test crash:

```bash
curl -X POST http://localhost:8080/api/v1/crashes \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-app-api-key" \
  -d '{
    "app_version": "1.0.0-test",
    "platform": "test",
    "error_type": "TestException",
    "error_message": "This is a test crash",
    "stack_trace": [],
    "environment": "development"
  }'
```

## Troubleshooting

### Alerts Not Sending

1. Check alert is enabled: `"enabled": true`
2. Verify app_id matches the app submitting crashes
3. Check server logs for errors
4. For webhooks, verify the endpoint is accessible from the server

### Email Not Arriving

1. Check spam folder
2. Verify SMTP credentials in config
3. Some providers require "app passwords" (e.g., Gmail with 2FA)
4. Check server logs for SMTP errors

### Slack Messages Not Appearing

1. Verify webhook URL is correct
2. Check the channel exists
3. Ensure the Slack app has permission to post to the channel
4. Test the webhook directly with curl

### Webhook Timeouts

Webhooks have a 10-second timeout. If your endpoint is slow:
- Return 200 immediately and process async
- Use a faster endpoint
- Check your endpoint's health
