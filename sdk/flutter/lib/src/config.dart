/// Configuration for the Inceptor SDK
class InceptorConfig {
  /// The Inceptor server endpoint URL
  final String endpoint;

  /// API key for authentication
  final String apiKey;

  /// Optional app ID (derived from package name if not provided)
  final String? appId;

  /// Current environment (production, staging, development)
  final String environment;

  /// Whether to enable debug logging
  final bool debug;

  /// Maximum breadcrumbs to keep
  final int maxBreadcrumbs;

  /// Whether to capture screenshots on crash (not implemented yet)
  final bool captureScreenshots;

  /// Timeout for API requests in milliseconds
  final int timeout;

  /// Whether to enable offline queue
  final bool enableOfflineQueue;

  /// Maximum number of crashes to queue offline
  final int maxOfflineQueueSize;

  /// Custom tags to add to all crash reports
  final Map<String, String>? tags;

  const InceptorConfig({
    required this.endpoint,
    required this.apiKey,
    this.appId,
    this.environment = 'production',
    this.debug = false,
    this.maxBreadcrumbs = 50,
    this.captureScreenshots = false,
    this.timeout = 30000,
    this.enableOfflineQueue = true,
    this.maxOfflineQueueSize = 100,
    this.tags,
  });

  /// Create a copy with updated values
  InceptorConfig copyWith({
    String? endpoint,
    String? apiKey,
    String? appId,
    String? environment,
    bool? debug,
    int? maxBreadcrumbs,
    bool? captureScreenshots,
    int? timeout,
    bool? enableOfflineQueue,
    int? maxOfflineQueueSize,
    Map<String, String>? tags,
  }) {
    return InceptorConfig(
      endpoint: endpoint ?? this.endpoint,
      apiKey: apiKey ?? this.apiKey,
      appId: appId ?? this.appId,
      environment: environment ?? this.environment,
      debug: debug ?? this.debug,
      maxBreadcrumbs: maxBreadcrumbs ?? this.maxBreadcrumbs,
      captureScreenshots: captureScreenshots ?? this.captureScreenshots,
      timeout: timeout ?? this.timeout,
      enableOfflineQueue: enableOfflineQueue ?? this.enableOfflineQueue,
      maxOfflineQueueSize: maxOfflineQueueSize ?? this.maxOfflineQueueSize,
      tags: tags ?? this.tags,
    );
  }
}
