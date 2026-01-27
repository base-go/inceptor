# Inceptor Flutter SDK

[![pub package](https://img.shields.io/pub/v/inceptor_flutter.svg)](https://pub.dev/packages/inceptor_flutter)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Flutter SDK for [Inceptor](https://github.com/base-go/inceptor) - a lightweight, self-hosted crash logging and error tracking service for mobile apps.

## Features

- **Automatic crash capture** - Catches Flutter errors and unhandled exceptions
- **Stack trace parsing** - Normalizes and symbolizes stack traces
- **Breadcrumbs** - Track user actions, navigation, and HTTP requests leading up to crashes
- **Offline support** - Queues crashes when offline, sends when connectivity returns
- **User context** - Associate crashes with user IDs and custom metadata
- **Multi-platform** - iOS, Android, Web, macOS, Windows, Linux

## Installation

Add to your `pubspec.yaml`:

```yaml
dependencies:
  inceptor_flutter: ^1.0.1
```

Then run:
```bash
flutter pub get
```

## Quick Start

### 1. Initialize in main.dart

```dart
import 'dart:async';
import 'package:flutter/material.dart';
import 'package:inceptor_flutter/inceptor_flutter.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize Inceptor
  await Inceptor.init(
    endpoint: 'https://inceptor.yourdomain.com',
    apiKey: 'your-api-key-from-dashboard',
    environment: 'production',
  );

  // Capture Flutter framework errors
  FlutterError.onError = Inceptor.recordFlutterError;

  // Capture async errors
  runZonedGuarded(() {
    runApp(MyApp());
  }, Inceptor.recordError);
}
```

### 2. Add Navigation Tracking (Optional)

```dart
MaterialApp(
  navigatorObservers: [InceptorNavigatorObserver()],
  home: HomeScreen(),
);
```

## Configuration Options

```dart
await Inceptor.init(
  // Required
  endpoint: 'https://inceptor.yourdomain.com',
  apiKey: 'your-api-key',

  // Optional
  appId: 'my-app',                    // App identifier (auto-detected if not set)
  environment: 'production',          // Environment tag (default: 'production')
  debug: false,                       // Enable debug logging
  maxBreadcrumbs: 50,                 // Max breadcrumbs to keep (default: 50)
  timeout: 30000,                     // API timeout in milliseconds (default: 30000)
  enableOfflineQueue: true,           // Queue crashes when offline (default: true)
  tags: {                             // Custom tags attached to all reports
    'app_variant': 'free',
    'feature_flags': 'dark_mode',
  },
);
```

## Manual Error Capture

Capture handled exceptions with context:

```dart
try {
  await someRiskyOperation();
} catch (e, stackTrace) {
  await Inceptor.captureException(
    e,
    stackTrace: stackTrace,
    context: {
      'operation': 'payment_processing',
      'amount': 99.99,
    },
  );
}
```

Capture messages/logs:

```dart
await Inceptor.captureMessage(
  'User attempted checkout with empty cart',
  level: 'warning',
  context: {'user_id': '123'},
);
```

## Breadcrumbs

Breadcrumbs help you understand what happened before a crash:

```dart
// Navigation breadcrumbs (automatic with InceptorNavigatorObserver)
Inceptor.addNavigationBreadcrumb(
  from: '/home',
  to: '/checkout',
);

// HTTP request breadcrumbs
Inceptor.addHttpBreadcrumb(
  method: 'POST',
  url: 'https://api.example.com/orders',
  statusCode: 201,
);

// User action breadcrumbs
Inceptor.addUserBreadcrumb(
  action: 'button_tap',
  target: 'checkout_button',
  data: {'cart_items': 3},
);

// Custom breadcrumbs
Inceptor.addBreadcrumb(
  InceptorBreadcrumb.custom(
    type: 'state',
    category: 'cart',
    message: 'Item added to cart',
    data: {'product_id': 'SKU123', 'quantity': 2},
  ),
);
```

## User Context

Associate crashes with users:

```dart
// Set user after login
Inceptor.setUser('user_12345');

// Add custom metadata
Inceptor.setMetadata('subscription_tier', 'premium');
Inceptor.setMetadata('account_age_days', 365);

// Clear on logout
Inceptor.setUser(null);
Inceptor.clearMetadata();
```

## HTTP Client Integration

Track HTTP requests as breadcrumbs with a wrapper:

```dart
import 'package:http/http.dart' as http;

Future<http.Response> trackedRequest(
  String method,
  Uri url, {
  Map<String, String>? headers,
  Object? body,
}) async {
  try {
    final response = await http.Request(method, url)
      ..headers.addAll(headers ?? {})
      ..body = body?.toString() ?? '';

    final streamedResponse = await http.Client().send(response);
    final httpResponse = await http.Response.fromStream(streamedResponse);

    Inceptor.addHttpBreadcrumb(
      method: method,
      url: url.toString(),
      statusCode: httpResponse.statusCode,
    );

    return httpResponse;
  } catch (e) {
    Inceptor.addHttpBreadcrumb(
      method: method,
      url: url.toString(),
      reason: e.toString(),
    );
    rethrow;
  }
}
```

## Server Setup

Inceptor requires a self-hosted server. Deploy with [BasePod](https://github.com/base-go/basepod):

```bash
# Clone and deploy
git clone https://github.com/base-go/inceptor
cd inceptor
bp deploy
```

Or run with Docker:

```bash
docker run -d -p 8080:8080 -v inceptor-data:/app/data ghcr.io/base-go/inceptor
```

Get your API key from the Inceptor dashboard after creating an app.

## Example App

See the [example](https://github.com/base-go/inceptor/tree/main/sdk/flutter/example) directory for a complete Flutter app demonstrating all features.

## API Reference

### Inceptor (Static Methods)

| Method | Description |
|--------|-------------|
| `init()` | Initialize the SDK |
| `recordFlutterError()` | Handler for FlutterError.onError |
| `recordError()` | Handler for runZonedGuarded |
| `captureException()` | Manually capture an exception |
| `captureMessage()` | Capture a message/log |
| `setUser()` | Set or clear user ID |
| `setMetadata()` | Add custom metadata |
| `removeMetadata()` | Remove a metadata key |
| `clearMetadata()` | Clear all metadata |
| `addBreadcrumb()` | Add a custom breadcrumb |
| `addNavigationBreadcrumb()` | Add navigation breadcrumb |
| `addHttpBreadcrumb()` | Add HTTP request breadcrumb |
| `addUserBreadcrumb()` | Add user action breadcrumb |
| `clearBreadcrumbs()` | Clear all breadcrumbs |

## Troubleshooting

### Crashes not appearing in dashboard

1. Check that `endpoint` and `apiKey` are correct
2. Enable `debug: true` to see console logs
3. Verify network connectivity
4. Check server logs for errors

### Offline queue not working

Ensure `enableOfflineQueue: true` (default) and the app has been initialized before crashes occur.

## Contributing

Contributions welcome! Please read our [contributing guidelines](https://github.com/base-go/inceptor/blob/main/CONTRIBUTING.md).

## License

MIT License - see [LICENSE](LICENSE) for details.

## Links

- [GitHub Repository](https://github.com/base-go/inceptor)
- [Issue Tracker](https://github.com/base-go/inceptor/issues)
- [Changelog](https://github.com/base-go/inceptor/blob/main/sdk/flutter/CHANGELOG.md)
