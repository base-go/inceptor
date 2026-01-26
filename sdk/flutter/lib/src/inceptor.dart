import 'dart:async';
import 'dart:collection';
import 'dart:convert';
import 'dart:io';

import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:device_info_plus/device_info_plus.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:http/http.dart' as http;
import 'package:package_info_plus/package_info_plus.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'config.dart';
import 'models/breadcrumb.dart';
import 'models/crash_report.dart';
import 'models/stack_frame.dart';

/// Main Inceptor SDK class
class Inceptor {
  static Inceptor? _instance;
  static bool _initialized = false;

  late InceptorConfig _config;
  late http.Client _httpClient;

  String? _appVersion;
  String? _platform;
  String? _osVersion;
  String? _deviceModel;
  String? _userId;
  Map<String, dynamic> _globalMetadata = {};

  final Queue<InceptorBreadcrumb> _breadcrumbs = Queue();
  final List<InceptorCrashReport> _offlineQueue = [];

  Inceptor._();

  /// Get the singleton instance
  static Inceptor get instance {
    if (_instance == null) {
      throw StateError(
          'Inceptor not initialized. Call Inceptor.init() first.');
    }
    return _instance!;
  }

  /// Check if Inceptor is initialized
  static bool get isInitialized => _initialized;

  /// Initialize the Inceptor SDK
  ///
  /// Must be called before any other Inceptor methods.
  ///
  /// Example:
  /// ```dart
  /// await Inceptor.init(
  ///   endpoint: 'https://your-server.com',
  ///   apiKey: 'your-api-key',
  ///   environment: 'production',
  /// );
  /// ```
  static Future<void> init({
    required String endpoint,
    required String apiKey,
    String? appId,
    String environment = 'production',
    bool debug = false,
    int maxBreadcrumbs = 50,
    int timeout = 30000,
    bool enableOfflineQueue = true,
    Map<String, String>? tags,
  }) async {
    if (_initialized) {
      _log('Inceptor already initialized');
      return;
    }

    _instance = Inceptor._();
    _instance!._config = InceptorConfig(
      endpoint: endpoint,
      apiKey: apiKey,
      appId: appId,
      environment: environment,
      debug: debug,
      maxBreadcrumbs: maxBreadcrumbs,
      timeout: timeout,
      enableOfflineQueue: enableOfflineQueue,
      tags: tags,
    );

    _instance!._httpClient = http.Client();

    // Gather device info
    await _instance!._gatherDeviceInfo();

    // Load offline queue
    if (enableOfflineQueue) {
      await _instance!._loadOfflineQueue();
      // Try to flush any queued crashes
      _instance!._flushOfflineQueue();
    }

    _initialized = true;
    _log('Inceptor initialized');
  }

  /// Record a Flutter error (use as FlutterError.onError handler)
  ///
  /// Example:
  /// ```dart
  /// FlutterError.onError = Inceptor.recordFlutterError;
  /// ```
  static void recordFlutterError(FlutterErrorDetails details) {
    if (!_initialized) return;

    instance._recordError(
      details.exception,
      details.stack ?? StackTrace.current,
      context: {
        'flutter_error': true,
        'library': details.library,
        if (details.context != null) 'context': details.context.toString(),
      },
    );
  }

  /// Record an error (use with runZonedGuarded)
  ///
  /// Example:
  /// ```dart
  /// runZonedGuarded(() {
  ///   runApp(MyApp());
  /// }, Inceptor.recordError);
  /// ```
  static void recordError(dynamic error, StackTrace stackTrace) {
    if (!_initialized) return;
    instance._recordError(error, stackTrace);
  }

  /// Manually record an error with optional context
  ///
  /// Example:
  /// ```dart
  /// try {
  ///   // risky operation
  /// } catch (e, stackTrace) {
  ///   Inceptor.captureException(e, stackTrace: stackTrace, context: {
  ///     'user_action': 'checkout',
  ///   });
  /// }
  /// ```
  static Future<InceptorCrashResponse?> captureException(
    dynamic exception, {
    StackTrace? stackTrace,
    Map<String, dynamic>? context,
  }) async {
    if (!_initialized) return null;
    return instance._recordError(
      exception,
      stackTrace ?? StackTrace.current,
      context: context,
    );
  }

  /// Capture a message as a crash report
  static Future<InceptorCrashResponse?> captureMessage(
    String message, {
    String level = 'error',
    Map<String, dynamic>? context,
  }) async {
    if (!_initialized) return null;

    final report = InceptorCrashReport(
      appVersion: instance._appVersion ?? 'unknown',
      platform: instance._platform ?? 'flutter',
      osVersion: instance._osVersion,
      deviceModel: instance._deviceModel,
      errorType: 'Message',
      errorMessage: message,
      stackTrace: InceptorStackFrame.parseStackTrace(StackTrace.current),
      userId: instance._userId,
      environment: instance._config.environment,
      metadata: {
        'level': level,
        ...?context,
        ...instance._globalMetadata,
      },
      breadcrumbs: instance._breadcrumbs.toList(),
    );

    return instance._submitReport(report);
  }

  /// Set the current user ID
  static void setUser(String? userId) {
    if (!_initialized) return;
    instance._userId = userId;
  }

  /// Set global metadata that will be attached to all crash reports
  static void setMetadata(String key, dynamic value) {
    if (!_initialized) return;
    instance._globalMetadata[key] = value;
  }

  /// Remove a metadata key
  static void removeMetadata(String key) {
    if (!_initialized) return;
    instance._globalMetadata.remove(key);
  }

  /// Clear all metadata
  static void clearMetadata() {
    if (!_initialized) return;
    instance._globalMetadata.clear();
  }

  /// Add a breadcrumb
  static void addBreadcrumb(InceptorBreadcrumb breadcrumb) {
    if (!_initialized) return;

    instance._breadcrumbs.addLast(breadcrumb);
    while (instance._breadcrumbs.length > instance._config.maxBreadcrumbs) {
      instance._breadcrumbs.removeFirst();
    }
  }

  /// Add a navigation breadcrumb
  static void addNavigationBreadcrumb({
    required String from,
    required String to,
  }) {
    addBreadcrumb(InceptorBreadcrumb.navigation(from: from, to: to));
  }

  /// Add an HTTP breadcrumb
  static void addHttpBreadcrumb({
    required String method,
    required String url,
    int? statusCode,
    String? reason,
  }) {
    addBreadcrumb(InceptorBreadcrumb.http(
      method: method,
      url: url,
      statusCode: statusCode,
      reason: reason,
    ));
  }

  /// Add a user action breadcrumb
  static void addUserBreadcrumb({
    required String action,
    String? target,
    Map<String, dynamic>? data,
  }) {
    addBreadcrumb(InceptorBreadcrumb.user(
      action: action,
      target: target,
      data: data,
    ));
  }

  /// Clear all breadcrumbs
  static void clearBreadcrumbs() {
    if (!_initialized) return;
    instance._breadcrumbs.clear();
  }

  // Private methods

  Future<void> _gatherDeviceInfo() async {
    try {
      // Get package info
      final packageInfo = await PackageInfo.fromPlatform();
      _appVersion = '${packageInfo.version}+${packageInfo.buildNumber}';

      // Get device info
      final deviceInfo = DeviceInfoPlugin();

      if (kIsWeb) {
        _platform = 'web';
        final webInfo = await deviceInfo.webBrowserInfo;
        _osVersion = webInfo.platform;
        _deviceModel = webInfo.browserName.name;
      } else if (Platform.isAndroid) {
        _platform = 'android';
        final androidInfo = await deviceInfo.androidInfo;
        _osVersion = 'Android ${androidInfo.version.release}';
        _deviceModel = '${androidInfo.manufacturer} ${androidInfo.model}';
      } else if (Platform.isIOS) {
        _platform = 'ios';
        final iosInfo = await deviceInfo.iosInfo;
        _osVersion = '${iosInfo.systemName} ${iosInfo.systemVersion}';
        _deviceModel = iosInfo.model;
      } else if (Platform.isMacOS) {
        _platform = 'macos';
        final macInfo = await deviceInfo.macOsInfo;
        _osVersion = 'macOS ${macInfo.osRelease}';
        _deviceModel = macInfo.model;
      } else if (Platform.isWindows) {
        _platform = 'windows';
        final windowsInfo = await deviceInfo.windowsInfo;
        _osVersion = 'Windows ${windowsInfo.majorVersion}';
        _deviceModel = windowsInfo.computerName;
      } else if (Platform.isLinux) {
        _platform = 'linux';
        final linuxInfo = await deviceInfo.linuxInfo;
        _osVersion = linuxInfo.prettyName;
        _deviceModel = linuxInfo.name;
      }
    } catch (e) {
      _log('Failed to gather device info: $e');
      _platform = 'flutter';
    }
  }

  Future<InceptorCrashResponse?> _recordError(
    dynamic error,
    StackTrace stackTrace, {
    Map<String, dynamic>? context,
  }) async {
    final report = InceptorCrashReport.fromError(
      error: error,
      stackTrace: stackTrace,
      appVersion: _appVersion ?? 'unknown',
      platform: _platform ?? 'flutter',
      osVersion: _osVersion,
      deviceModel: _deviceModel,
      userId: _userId,
      environment: _config.environment,
      metadata: {
        ...?context,
        ..._globalMetadata,
        if (_config.tags != null) ...?_config.tags,
      },
      breadcrumbs: _breadcrumbs.toList(),
    );

    return _submitReport(report);
  }

  Future<InceptorCrashResponse?> _submitReport(
      InceptorCrashReport report) async {
    try {
      final url = Uri.parse('${_config.endpoint}/api/v1/crashes');
      final response = await _httpClient.post(
        url,
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': _config.apiKey,
        },
        body: jsonEncode(report.toJson()),
      ).timeout(Duration(milliseconds: _config.timeout));

      if (response.statusCode >= 200 && response.statusCode < 300) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        _log('Crash reported: ${data['id']}');
        return InceptorCrashResponse.fromJson(data);
      } else {
        _log('Failed to report crash: ${response.statusCode}');
        _queueOffline(report);
        return null;
      }
    } catch (e) {
      _log('Error reporting crash: $e');
      _queueOffline(report);
      return null;
    }
  }

  void _queueOffline(InceptorCrashReport report) {
    if (!_config.enableOfflineQueue) return;

    _offlineQueue.add(report);
    if (_offlineQueue.length > _config.maxOfflineQueueSize) {
      _offlineQueue.removeAt(0);
    }
    _saveOfflineQueue();
  }

  Future<void> _saveOfflineQueue() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final data = _offlineQueue.map((r) => jsonEncode(r.toJson())).toList();
      await prefs.setStringList('inceptor_offline_queue', data);
    } catch (e) {
      _log('Failed to save offline queue: $e');
    }
  }

  Future<void> _loadOfflineQueue() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final data = prefs.getStringList('inceptor_offline_queue') ?? [];
      for (final item in data) {
        try {
          final json = jsonDecode(item) as Map<String, dynamic>;
          _offlineQueue.add(InceptorCrashReport.fromJson(json));
        } catch (e) {
          _log('Failed to parse queued crash: $e');
        }
      }
      _log('Loaded ${_offlineQueue.length} crashes from offline queue');
    } catch (e) {
      _log('Failed to load offline queue: $e');
    }
  }

  Future<void> _flushOfflineQueue() async {
    if (_offlineQueue.isEmpty) return;

    // Check connectivity before attempting to flush
    if (!await _hasConnectivity()) {
      _log('No connectivity, skipping offline queue flush');
      return;
    }

    _log('Flushing ${_offlineQueue.length} offline crashes');
    final queue = List<InceptorCrashReport>.from(_offlineQueue);
    _offlineQueue.clear();

    for (final report in queue) {
      await _submitReport(report);
    }

    await _saveOfflineQueue();
  }

  Future<bool> _hasConnectivity() async {
    try {
      final result = await Connectivity().checkConnectivity();
      // connectivity_plus 5.x returns ConnectivityResult, 6.x+ returns List
      if (result is List) {
        return (result as List).isNotEmpty &&
            !(result as List).contains(ConnectivityResult.none);
      }
      return result != ConnectivityResult.none;
    } catch (e) {
      _log('Failed to check connectivity: $e');
      return true; // Assume connected if check fails
    }
  }

  static void _log(String message) {
    if (_instance?._config.debug ?? false) {
      debugPrint('[Inceptor] $message');
    }
  }
}
