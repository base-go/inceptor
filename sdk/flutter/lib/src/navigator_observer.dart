import 'package:flutter/widgets.dart';

import 'inceptor.dart';

/// Navigator observer that automatically tracks navigation as breadcrumbs.
///
/// Add this to your [MaterialApp] or [Navigator] to automatically record
/// navigation events.
///
/// Example:
/// ```dart
/// MaterialApp(
///   navigatorObservers: [InceptorNavigatorObserver()],
///   // ...
/// );
/// ```
class InceptorNavigatorObserver extends NavigatorObserver {
  @override
  void didPush(Route<dynamic> route, Route<dynamic>? previousRoute) {
    super.didPush(route, previousRoute);
    if (Inceptor.isInitialized) {
      Inceptor.addNavigationBreadcrumb(
        from: _getRouteName(previousRoute),
        to: _getRouteName(route),
      );
    }
  }

  @override
  void didPop(Route<dynamic> route, Route<dynamic>? previousRoute) {
    super.didPop(route, previousRoute);
    if (Inceptor.isInitialized) {
      Inceptor.addNavigationBreadcrumb(
        from: _getRouteName(route),
        to: _getRouteName(previousRoute),
      );
    }
  }

  @override
  void didReplace({Route<dynamic>? newRoute, Route<dynamic>? oldRoute}) {
    super.didReplace(newRoute: newRoute, oldRoute: oldRoute);
    if (Inceptor.isInitialized) {
      Inceptor.addNavigationBreadcrumb(
        from: _getRouteName(oldRoute),
        to: _getRouteName(newRoute),
      );
    }
  }

  @override
  void didRemove(Route<dynamic> route, Route<dynamic>? previousRoute) {
    super.didRemove(route, previousRoute);
    if (Inceptor.isInitialized) {
      Inceptor.addNavigationBreadcrumb(
        from: _getRouteName(route),
        to: _getRouteName(previousRoute),
      );
    }
  }

  String _getRouteName(Route<dynamic>? route) {
    if (route == null) return '/';
    return route.settings.name ?? route.runtimeType.toString();
  }
}
