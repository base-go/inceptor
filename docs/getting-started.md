# Getting Started with Inceptor

This guide walks you through setting up Inceptor and reporting your first crash.

## Prerequisites

- Docker (for containerized deployment) or Go 1.22+ (for local development)
- A Flutter app (or any HTTP-capable client)

## Installation Options

### Option 1: BasePod Deployment (Recommended)

BasePod provides the simplest deployment experience:

```bash
# Clone the repository
git clone https://github.com/flakerimi/inceptor.git
cd inceptor

# Deploy to BasePod
bp push
```

Your Inceptor instance will be available at the URL provided by BasePod.

### Option 2: Docker

```bash
# Build the image
docker build -t inceptor .

# Run the container
docker run -d \
  -p 8080:8080 \
  -v inceptor-data:/app/data \
  -e INCEPTOR_AUTH_ADMIN_KEY=your-secure-key \
  inceptor
```

### Option 3: From Source

```bash
# Clone and build
git clone https://github.com/flakerimi/inceptor.git
cd inceptor

# Build the web dashboard
cd web && npm install && npm run generate && cd ..

# Copy static files
cp -r web/.output/public internal/api/rest/static/

# Build and run
go build -o inceptor ./cmd/inceptor
./inceptor
```

## Initial Configuration

Create a configuration file or use environment variables:

```yaml
# configs/config.yaml
server:
  rest_port: 8080
  host: "0.0.0.0"

storage:
  sqlite_path: "./data/inceptor.db"
  logs_path: "./data/crashes"

auth:
  enabled: true
  admin_key: "your-secure-admin-key"

retention:
  default_days: 30
```

Or use environment variables:

```bash
export INCEPTOR_SERVER_REST_PORT=8080
export INCEPTOR_AUTH_ADMIN_KEY=your-secure-admin-key
export INCEPTOR_STORAGE_SQLITE_PATH=/app/data/inceptor.db
```

## Register Your First App

Register an app to get an API key:

```bash
curl -X POST http://localhost:8080/api/v1/apps \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-admin-key" \
  -d '{"name": "My Flutter App", "retention_days": 30}'
```

Response:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "My Flutter App",
  "api_key": "ink_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "retention_days": 30
}
```

**Important**: Save the `api_key` immediately - it's only shown once!

## Integrate with Flutter

Add the Inceptor SDK to your Flutter app:

```yaml
# pubspec.yaml
dependencies:
  inceptor_flutter:
    path: ../inceptor/sdk/flutter
```

Initialize in your app:

```dart
import 'package:inceptor_flutter/inceptor_flutter.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  await Inceptor.init(
    endpoint: 'https://your-inceptor-server.com',
    apiKey: 'ink_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6',
    environment: 'production',
  );

  // Capture Flutter errors
  FlutterError.onError = Inceptor.recordFlutterError;

  // Capture async errors
  runZonedGuarded(() {
    runApp(MyApp());
  }, Inceptor.recordError);
}
```

## Submit a Test Crash

You can test your setup with a curl request:

```bash
curl -X POST http://localhost:8080/api/v1/crashes \
  -H "Content-Type: application/json" \
  -H "X-API-Key: ink_your-app-api-key" \
  -d '{
    "app_version": "1.0.0",
    "platform": "flutter",
    "os_version": "Android 14",
    "device_model": "Pixel 8",
    "error_type": "FormatException",
    "error_message": "Invalid date format",
    "stack_trace": [
      {
        "file_name": "date_parser.dart",
        "line_number": 42,
        "method_name": "parse",
        "class_name": "DateParser"
      }
    ],
    "environment": "production"
  }'
```

## Access the Dashboard

Open your browser and navigate to:

```
http://localhost:8080
```

Log in with your admin credentials to view crashes, manage groups, and configure alerts.

## Next Steps

- [Configure alerting](./alerting.md) to get notified of new crashes
- [Read the API reference](./api-reference.md) for advanced usage
- [Explore the Flutter SDK](./flutter-sdk.md) for all available features
