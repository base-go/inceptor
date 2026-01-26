import 'breadcrumb.dart';
import 'stack_frame.dart';

/// Represents a crash report to be sent to the server
class InceptorCrashReport {
  /// App version string
  final String appVersion;

  /// Platform (ios, android, web, etc.)
  final String platform;

  /// OS version
  final String? osVersion;

  /// Device model
  final String? deviceModel;

  /// Error type/class name
  final String errorType;

  /// Error message
  final String errorMessage;

  /// Stack trace frames
  final List<InceptorStackFrame> stackTrace;

  /// Optional user identifier
  final String? userId;

  /// Environment (production, staging, development)
  final String environment;

  /// Custom metadata
  final Map<String, dynamic>? metadata;

  /// Breadcrumbs leading up to the crash
  final List<InceptorBreadcrumb>? breadcrumbs;

  const InceptorCrashReport({
    required this.appVersion,
    required this.platform,
    this.osVersion,
    this.deviceModel,
    required this.errorType,
    required this.errorMessage,
    required this.stackTrace,
    this.userId,
    required this.environment,
    this.metadata,
    this.breadcrumbs,
  });

  /// Convert to JSON for API submission
  Map<String, dynamic> toJson() {
    return {
      'app_version': appVersion,
      'platform': platform,
      if (osVersion != null) 'os_version': osVersion,
      if (deviceModel != null) 'device_model': deviceModel,
      'error_type': errorType,
      'error_message': errorMessage,
      'stack_trace': stackTrace.map((f) => f.toJson()).toList(),
      if (userId != null) 'user_id': userId,
      'environment': environment,
      if (metadata != null) 'metadata': metadata,
      if (breadcrumbs != null)
        'breadcrumbs': breadcrumbs!.map((b) => b.toJson()).toList(),
    };
  }

  /// Create from an exception and stack trace
  factory InceptorCrashReport.fromError({
    required dynamic error,
    required StackTrace stackTrace,
    required String appVersion,
    required String platform,
    String? osVersion,
    String? deviceModel,
    String? userId,
    required String environment,
    Map<String, dynamic>? metadata,
    List<InceptorBreadcrumb>? breadcrumbs,
  }) {
    return InceptorCrashReport(
      appVersion: appVersion,
      platform: platform,
      osVersion: osVersion,
      deviceModel: deviceModel,
      errorType: error.runtimeType.toString(),
      errorMessage: error.toString(),
      stackTrace: InceptorStackFrame.parseStackTrace(stackTrace),
      userId: userId,
      environment: environment,
      metadata: metadata,
      breadcrumbs: breadcrumbs,
    );
  }

  /// Create from JSON (for offline queue restoration)
  factory InceptorCrashReport.fromJson(Map<String, dynamic> json) {
    return InceptorCrashReport(
      appVersion: json['app_version'] as String,
      platform: json['platform'] as String,
      osVersion: json['os_version'] as String?,
      deviceModel: json['device_model'] as String?,
      errorType: json['error_type'] as String,
      errorMessage: json['error_message'] as String,
      stackTrace: (json['stack_trace'] as List<dynamic>)
          .map((e) => InceptorStackFrame.fromJson(e as Map<String, dynamic>))
          .toList(),
      userId: json['user_id'] as String?,
      environment: json['environment'] as String,
      metadata: json['metadata'] as Map<String, dynamic>?,
      breadcrumbs: json['breadcrumbs'] != null
          ? (json['breadcrumbs'] as List<dynamic>)
              .map(
                  (e) => InceptorBreadcrumb.fromJson(e as Map<String, dynamic>))
              .toList()
          : null,
    );
  }

  @override
  String toString() {
    return 'InceptorCrashReport(errorType: $errorType, errorMessage: $errorMessage)';
  }
}

/// Response from submitting a crash report
class InceptorCrashResponse {
  /// The crash ID assigned by the server
  final String id;

  /// The group ID this crash belongs to
  final String groupId;

  /// The fingerprint of this crash
  final String fingerprint;

  /// Whether this crash created a new group
  final bool isNewGroup;

  const InceptorCrashResponse({
    required this.id,
    required this.groupId,
    required this.fingerprint,
    required this.isNewGroup,
  });

  factory InceptorCrashResponse.fromJson(Map<String, dynamic> json) {
    return InceptorCrashResponse(
      id: json['id'] as String,
      groupId: json['group_id'] as String,
      fingerprint: json['fingerprint'] as String,
      isNewGroup: json['is_new_group'] as bool? ?? false,
    );
  }

  @override
  String toString() {
    return 'InceptorCrashResponse(id: $id, groupId: $groupId, isNewGroup: $isNewGroup)';
  }
}
