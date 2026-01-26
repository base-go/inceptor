/// Represents a breadcrumb - an event leading up to a crash
class InceptorBreadcrumb {
  /// Timestamp of the event
  final DateTime timestamp;

  /// Type of breadcrumb (navigation, http, user, log, etc.)
  final String type;

  /// Category for grouping breadcrumbs
  final String category;

  /// Human-readable message
  final String message;

  /// Additional data
  final Map<String, dynamic>? data;

  /// Log level (debug, info, warning, error)
  final String level;

  const InceptorBreadcrumb({
    required this.timestamp,
    required this.type,
    required this.category,
    required this.message,
    this.data,
    this.level = 'info',
  });

  /// Create a navigation breadcrumb
  factory InceptorBreadcrumb.navigation({
    required String from,
    required String to,
  }) {
    return InceptorBreadcrumb(
      timestamp: DateTime.now(),
      type: 'navigation',
      category: 'navigation',
      message: 'Navigated from $from to $to',
      data: {'from': from, 'to': to},
    );
  }

  /// Create an HTTP breadcrumb
  factory InceptorBreadcrumb.http({
    required String method,
    required String url,
    int? statusCode,
    String? reason,
  }) {
    return InceptorBreadcrumb(
      timestamp: DateTime.now(),
      type: 'http',
      category: 'http',
      message: '$method $url${statusCode != null ? ' ($statusCode)' : ''}',
      data: {
        'method': method,
        'url': url,
        if (statusCode != null) 'status_code': statusCode,
        if (reason != null) 'reason': reason,
      },
      level: statusCode != null && statusCode >= 400 ? 'error' : 'info',
    );
  }

  /// Create a user action breadcrumb
  factory InceptorBreadcrumb.user({
    required String action,
    String? target,
    Map<String, dynamic>? data,
  }) {
    return InceptorBreadcrumb(
      timestamp: DateTime.now(),
      type: 'user',
      category: 'user_action',
      message: target != null ? '$action on $target' : action,
      data: {
        'action': action,
        if (target != null) 'target': target,
        ...?data,
      },
    );
  }

  /// Create a log breadcrumb
  factory InceptorBreadcrumb.log({
    required String message,
    String level = 'info',
    Map<String, dynamic>? data,
  }) {
    return InceptorBreadcrumb(
      timestamp: DateTime.now(),
      type: 'log',
      category: 'console',
      message: message,
      data: data,
      level: level,
    );
  }

  /// Create a custom breadcrumb
  factory InceptorBreadcrumb.custom({
    required String type,
    required String category,
    required String message,
    Map<String, dynamic>? data,
    String level = 'info',
  }) {
    return InceptorBreadcrumb(
      timestamp: DateTime.now(),
      type: type,
      category: category,
      message: message,
      data: data,
      level: level,
    );
  }

  /// Convert to JSON for API submission
  Map<String, dynamic> toJson() {
    return {
      'timestamp': timestamp.toUtc().toIso8601String(),
      'type': type,
      'category': category,
      'message': message,
      if (data != null) 'data': data,
      'level': level,
    };
  }

  @override
  String toString() {
    return '[$level] $type/$category: $message';
  }
}
