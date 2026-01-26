# Inceptor Flutter SDK

Flutter SDK for [Inceptor](https://github.com/flakerimi/inceptor) - a self-hosted crash logging service.

## Features

- Automatic crash and error capture
- Stack trace parsing and normalization
- Breadcrumbs for tracking user actions
- Offline queue with automatic retry
- Multi-platform support (iOS, Android, Web, macOS, Windows, Linux)

## Installation

```yaml
dependencies:
  inceptor_flutter: ^1.0.0
```

## Quick Start

```dart
import 'dart:async';
import 'package:flutter/material.dart';
import 'package:inceptor_flutter/inceptor_flutter.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  await Inceptor.init(
    endpoint: 'https://your-inceptor-server.com',
    apiKey: 'your-api-key',
    environment: 'production',
  );

  FlutterError.onError = Inceptor.recordFlutterError;

  runZonedGuarded(() {
    runApp(MyApp());
  }, Inceptor.recordError);
}
```

## Navigation Tracking

Add the navigator observer to automatically track navigation:

```dart
MaterialApp(
  navigatorObservers: [InceptorNavigatorObserver()],
  // ...
);
```

## Manual Error Capture

```dart
try {
  await riskyOperation();
} catch (e, stackTrace) {
  await Inceptor.captureException(
    e,
    stackTrace: stackTrace,
    context: {'operation': 'checkout'},
  );
}
```

## Breadcrumbs

```dart
// Navigation
Inceptor.addNavigationBreadcrumb(from: '/home', to: '/checkout');

// HTTP requests
Inceptor.addHttpBreadcrumb(
  method: 'POST',
  url: 'https://api.example.com/orders',
  statusCode: 201,
);

// User actions
Inceptor.addUserBreadcrumb(
  action: 'button_press',
  target: 'checkout_button',
);

// Custom
Inceptor.addBreadcrumb(
  InceptorBreadcrumb.custom(
    type: 'state',
    category: 'redux',
    message: 'Cart updated',
    data: {'items': 5},
  ),
);
```

## User Context

```dart
Inceptor.setUser('user_123');
Inceptor.setMetadata('subscription', 'premium');

// Clear on logout
Inceptor.setUser(null);
Inceptor.clearMetadata();
```

## Configuration Options

```dart
await Inceptor.init(
  endpoint: 'https://your-server.com',  // Required
  apiKey: 'your-api-key',                // Required
  environment: 'production',             // Default: 'production'
  debug: false,                          // Enable debug logging
  maxBreadcrumbs: 50,                    // Max breadcrumbs to keep
  timeout: 30000,                        // API timeout in ms
  enableOfflineQueue: true,              // Queue crashes when offline
  tags: {'app_variant': 'free'},         // Custom tags for all reports
);
```

## License

MIT License - see [LICENSE](LICENSE) for details.
