# Flutter SDK Documentation

The Inceptor Flutter SDK provides automatic crash reporting and error tracking for Flutter applications.

## Installation

Add the SDK to your `pubspec.yaml`:

```yaml
dependencies:
  inceptor_flutter:
    git:
      url: https://github.com/base-go/inceptor.git
      path: sdk/flutter
```

Or if using a local copy:

```yaml
dependencies:
  inceptor_flutter:
    path: ../inceptor/sdk/flutter
```

## Quick Start

Initialize Inceptor in your `main.dart`:

```dart
import 'dart:async';
import 'package:flutter/material.dart';
import 'package:inceptor_flutter/inceptor_flutter.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize Inceptor
  await Inceptor.init(
    endpoint: 'https://your-inceptor-server.com',
    apiKey: 'ink_your-api-key',
    environment: 'production',
  );

  // Capture Flutter framework errors
  FlutterError.onError = Inceptor.recordFlutterError;

  // Capture async/isolate errors
  runZonedGuarded(() {
    runApp(MyApp());
  }, Inceptor.recordError);
}
```

## Initialization Options

```dart
await Inceptor.init(
  // Required
  endpoint: 'https://your-server.com',  // Your Inceptor server URL
  apiKey: 'ink_...',                     // Your app's API key

  // Optional
  appId: 'com.example.myapp',            // Custom app identifier
  environment: 'production',              // Environment (production, staging, dev)
  debug: false,                           // Enable debug logging
  maxBreadcrumbs: 50,                    // Max breadcrumbs to store
  timeout: 30000,                        // Request timeout (ms)
  enableOfflineQueue: true,              // Queue crashes when offline
  tags: {                                // Global tags for all crashes
    'region': 'eu-west',
  },
);
```

## Error Capturing

### Automatic Error Capture

The SDK captures errors automatically when properly initialized:

```dart
// Flutter framework errors
FlutterError.onError = Inceptor.recordFlutterError;

// Async/zone errors
runZonedGuarded(() {
  runApp(MyApp());
}, Inceptor.recordError);
```

### Manual Error Capture

Capture caught exceptions manually:

```dart
try {
  await riskyOperation();
} catch (e, stackTrace) {
  await Inceptor.captureException(
    e,
    stackTrace: stackTrace,
    context: {
      'operation': 'data_sync',
      'user_action': 'button_press',
    },
  );
}
```

### Capture Messages

Send non-exception events:

```dart
await Inceptor.captureMessage(
  'User completed onboarding',
  level: 'info',
  context: {
    'steps_completed': 5,
    'time_taken_seconds': 120,
  },
);
```

## User Identification

Associate crashes with users:

```dart
// Set the current user
Inceptor.setUser('user-123');

// Clear user on logout
Inceptor.setUser(null);
```

## Metadata

Add contextual information to all crashes:

```dart
// Set global metadata (persists across crashes)
Inceptor.setMetadata('subscription_tier', 'premium');
Inceptor.setMetadata('feature_flags', ['new_checkout', 'dark_mode']);

// Remove specific metadata
Inceptor.removeMetadata('subscription_tier');

// Clear all metadata
Inceptor.clearMetadata();
```

## Breadcrumbs

Breadcrumbs create a trail of events leading up to a crash, making debugging easier.

### Navigation Breadcrumbs

Track screen transitions:

```dart
// Manual navigation breadcrumb
Inceptor.addNavigationBreadcrumb(
  from: '/home',
  to: '/checkout',
);
```

**Automatic Navigation Tracking**: Use the built-in navigator observer:

```dart
MaterialApp(
  navigatorObservers: [
    InceptorNavigatorObserver(),
  ],
  // ...
);
```

### HTTP Breadcrumbs

Track API calls:

```dart
Inceptor.addHttpBreadcrumb(
  method: 'POST',
  url: 'https://api.example.com/orders',
  statusCode: 201,
);

// For errors
Inceptor.addHttpBreadcrumb(
  method: 'GET',
  url: 'https://api.example.com/user',
  statusCode: 500,
  reason: 'Internal server error',
);
```

### User Action Breadcrumbs

Track user interactions:

```dart
Inceptor.addUserBreadcrumb(
  action: 'button_tap',
  target: 'add_to_cart_button',
  data: {
    'product_id': 'prod-123',
    'quantity': 2,
  },
);
```

### Custom Breadcrumbs

For complete control:

```dart
Inceptor.addBreadcrumb(
  InceptorBreadcrumb(
    type: 'custom',
    category: 'auth',
    message: 'User logged in successfully',
    level: 'info',
    data: {
      'method': 'google_oauth',
      'first_login': true,
    },
  ),
);
```

### Clear Breadcrumbs

```dart
Inceptor.clearBreadcrumbs();
```

## Offline Support

When enabled (`enableOfflineQueue: true`), the SDK queues crashes that fail to send and retries them when connectivity is restored.

The offline queue:
- Stores up to 100 crashes by default
- Persists across app restarts using SharedPreferences
- Automatically flushes when the device comes back online
- Checks connectivity before attempting to send

## Platform Support

The SDK collects platform-specific device information:

| Platform | Device Model | OS Version |
|----------|--------------|------------|
| Android | Manufacturer + Model | Android version |
| iOS | Model name | iOS version |
| Web | Browser name | Platform |
| macOS | Model | macOS version |
| Windows | Computer name | Windows version |
| Linux | Distribution name | Pretty name |

## Best Practices

### 1. Initialize Early

Initialize Inceptor as early as possible in your app lifecycle:

```dart
void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await Inceptor.init(...);  // Before runApp
  // ...
}
```

### 2. Use Meaningful Metadata

Add context that helps debugging:

```dart
// Good - specific and actionable
Inceptor.setMetadata('checkout_step', 'payment');
Inceptor.setMetadata('cart_total', 149.99);

// Avoid - too generic
Inceptor.setMetadata('data', someObject);
```

### 3. Structure Breadcrumbs Consistently

Use consistent categories and types:

```dart
// Define constants for consistency
class BreadcrumbCategories {
  static const auth = 'auth';
  static const checkout = 'checkout';
  static const api = 'api';
}
```

### 4. Don't Over-Report

Avoid capturing expected errors:

```dart
try {
  final user = await getUser();
} on UserNotFoundException {
  // Expected - don't report
  showUserNotFoundUI();
} catch (e, stack) {
  // Unexpected - report it
  Inceptor.captureException(e, stackTrace: stack);
  showGenericErrorUI();
}
```

### 5. Set User Context at Login

```dart
Future<void> login(String email, String password) async {
  final user = await authService.login(email, password);

  // Set user context for crash attribution
  Inceptor.setUser(user.id);
  Inceptor.setMetadata('user_email', user.email);
  Inceptor.setMetadata('account_type', user.accountType);
}

Future<void> logout() async {
  await authService.logout();
  Inceptor.setUser(null);
  Inceptor.clearMetadata();
}
```

## API Reference

### Inceptor Class

| Method | Description |
|--------|-------------|
| `init(...)` | Initialize the SDK |
| `isInitialized` | Check if SDK is initialized |
| `recordFlutterError(details)` | Record Flutter framework error |
| `recordError(error, stackTrace)` | Record zone/async error |
| `captureException(e, {...})` | Manually capture exception |
| `captureMessage(msg, {...})` | Capture a message |
| `setUser(userId)` | Set current user ID |
| `setMetadata(key, value)` | Set global metadata |
| `removeMetadata(key)` | Remove metadata key |
| `clearMetadata()` | Clear all metadata |
| `addBreadcrumb(crumb)` | Add custom breadcrumb |
| `addNavigationBreadcrumb(...)` | Add navigation breadcrumb |
| `addHttpBreadcrumb(...)` | Add HTTP breadcrumb |
| `addUserBreadcrumb(...)` | Add user action breadcrumb |
| `clearBreadcrumbs()` | Clear all breadcrumbs |

### InceptorBreadcrumb Class

```dart
InceptorBreadcrumb({
  required String type,
  required String category,
  required String message,
  String level = 'info',
  Map<String, dynamic>? data,
})

// Factory constructors
InceptorBreadcrumb.navigation({from, to})
InceptorBreadcrumb.http({method, url, statusCode, reason})
InceptorBreadcrumb.user({action, target, data})
```

### InceptorCrashResponse Class

Returned when a crash is successfully submitted:

```dart
class InceptorCrashResponse {
  final String id;          // Crash ID
  final String groupId;     // Group ID
  final String fingerprint; // Crash fingerprint
  final bool isNewGroup;    // Whether this created a new group
}
```

## Dependencies

The SDK uses these packages:
- `http` - HTTP client
- `device_info_plus` - Device information
- `package_info_plus` - App version info
- `shared_preferences` - Offline queue persistence
- `connectivity_plus` - Network status checking

## Troubleshooting

### Crashes Not Appearing

1. Check the API key is correct
2. Verify the endpoint URL (no trailing slash)
3. Enable debug mode: `debug: true`
4. Check network connectivity

### Device Info Not Collected

Ensure `WidgetsFlutterBinding.ensureInitialized()` is called before `Inceptor.init()`.

### Offline Queue Not Working

- Verify `enableOfflineQueue: true`
- Check that SharedPreferences has storage permissions
- The queue limit is 100 crashes by default

### Stack Traces Missing

For release builds, ensure you keep Flutter's stack traces:

```yaml
# pubspec.yaml
flutter:
  # Keep symbols in release builds
  deobfuscation: true
```
