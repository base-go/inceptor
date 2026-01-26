import 'package:flutter_test/flutter_test.dart';
import 'package:inceptor_flutter/inceptor_flutter.dart';

void main() {
  group('InceptorCrashReport', () {
    test('fromError creates report from exception', () {
      final error = FormatException('Invalid input');
      final stackTrace = StackTrace.fromString('''
#0      MyClass.method (package:app/main.dart:10:5)
''');

      final report = InceptorCrashReport.fromError(
        error: error,
        stackTrace: stackTrace,
        appVersion: '1.0.0+1',
        platform: 'android',
        osVersion: 'Android 14',
        deviceModel: 'Pixel 8',
        userId: 'user_123',
        environment: 'production',
        metadata: {'screen': 'checkout'},
        breadcrumbs: [
          InceptorBreadcrumb.navigation(from: '/home', to: '/checkout'),
        ],
      );

      expect(report.errorType, 'FormatException');
      expect(report.errorMessage, contains('Invalid input'));
      expect(report.appVersion, '1.0.0+1');
      expect(report.platform, 'android');
      expect(report.osVersion, 'Android 14');
      expect(report.deviceModel, 'Pixel 8');
      expect(report.userId, 'user_123');
      expect(report.environment, 'production');
      expect(report.metadata?['screen'], 'checkout');
      expect(report.breadcrumbs?.length, 1);
      expect(report.stackTrace.isNotEmpty, true);
    });

    test('toJson produces correct format', () {
      final report = InceptorCrashReport(
        appVersion: '1.0.0',
        platform: 'ios',
        osVersion: 'iOS 17',
        deviceModel: 'iPhone 15',
        errorType: 'Exception',
        errorMessage: 'Test error',
        stackTrace: [
          InceptorStackFrame(
            fileName: 'package:app/main.dart',
            lineNumber: 10,
            methodName: 'main',
          ),
        ],
        userId: 'user_456',
        environment: 'staging',
        metadata: {'key': 'value'},
        breadcrumbs: [],
      );

      final json = report.toJson();

      expect(json['app_version'], '1.0.0');
      expect(json['platform'], 'ios');
      expect(json['os_version'], 'iOS 17');
      expect(json['device_model'], 'iPhone 15');
      expect(json['error_type'], 'Exception');
      expect(json['error_message'], 'Test error');
      expect(json['user_id'], 'user_456');
      expect(json['environment'], 'staging');
      expect(json['metadata'], {'key': 'value'});
      expect(json['stack_trace'], isA<List>());
      expect(json['breadcrumbs'], isA<List>());
    });

    test('toJson omits null optional fields', () {
      final report = InceptorCrashReport(
        appVersion: '1.0.0',
        platform: 'ios',
        errorType: 'Exception',
        errorMessage: 'Test error',
        stackTrace: [],
        environment: 'production',
      );

      final json = report.toJson();

      expect(json.containsKey('os_version'), false);
      expect(json.containsKey('device_model'), false);
      expect(json.containsKey('user_id'), false);
      expect(json.containsKey('metadata'), false);
      expect(json.containsKey('breadcrumbs'), false);
    });

    test('fromJson reconstructs report correctly', () {
      final json = {
        'app_version': '1.0.0',
        'platform': 'android',
        'os_version': 'Android 14',
        'device_model': 'Pixel 8',
        'error_type': 'FormatException',
        'error_message': 'Invalid format',
        'stack_trace': [
          {
            'file_name': 'package:app/main.dart',
            'line_number': 42,
            'method_name': 'process',
            'native': false,
          }
        ],
        'user_id': 'user_789',
        'environment': 'production',
        'metadata': {'screen': 'home'},
        'breadcrumbs': [
          {
            'timestamp': '2024-01-15T10:30:00.000Z',
            'type': 'navigation',
            'category': 'navigation',
            'message': 'Navigated',
            'level': 'info',
          }
        ],
      };

      final report = InceptorCrashReport.fromJson(json);

      expect(report.appVersion, '1.0.0');
      expect(report.platform, 'android');
      expect(report.osVersion, 'Android 14');
      expect(report.deviceModel, 'Pixel 8');
      expect(report.errorType, 'FormatException');
      expect(report.errorMessage, 'Invalid format');
      expect(report.userId, 'user_789');
      expect(report.environment, 'production');
      expect(report.metadata?['screen'], 'home');
      expect(report.stackTrace.length, 1);
      expect(report.stackTrace[0].methodName, 'process');
      expect(report.breadcrumbs?.length, 1);
    });

    test('roundtrip toJson/fromJson preserves data', () {
      final original = InceptorCrashReport(
        appVersion: '2.0.0+5',
        platform: 'web',
        osVersion: 'Chrome 120',
        deviceModel: 'Desktop',
        errorType: 'TypeError',
        errorMessage: 'null is not an object',
        stackTrace: [
          InceptorStackFrame(
            fileName: 'package:app/service.dart',
            lineNumber: 100,
            columnNumber: 15,
            methodName: 'fetchData',
            className: 'ApiService',
            native: false,
          ),
        ],
        userId: 'test_user',
        environment: 'development',
        metadata: {'feature_flag': true, 'count': 42},
        breadcrumbs: [
          InceptorBreadcrumb.http(
            method: 'GET',
            url: 'https://api.test.com',
            statusCode: 500,
          ),
        ],
      );

      final json = original.toJson();
      final restored = InceptorCrashReport.fromJson(json);

      expect(restored.appVersion, original.appVersion);
      expect(restored.platform, original.platform);
      expect(restored.errorType, original.errorType);
      expect(restored.errorMessage, original.errorMessage);
      expect(restored.stackTrace.length, original.stackTrace.length);
      expect(restored.stackTrace[0].className, 'ApiService');
      expect(restored.breadcrumbs?.length, original.breadcrumbs?.length);
    });
  });

  group('InceptorCrashResponse', () {
    test('fromJson parses response', () {
      final json = {
        'id': 'crash-123',
        'group_id': 'group-456',
        'fingerprint': 'abc123def456',
        'is_new_group': true,
      };

      final response = InceptorCrashResponse.fromJson(json);

      expect(response.id, 'crash-123');
      expect(response.groupId, 'group-456');
      expect(response.fingerprint, 'abc123def456');
      expect(response.isNewGroup, true);
    });

    test('fromJson handles missing is_new_group', () {
      final json = {
        'id': 'crash-123',
        'group_id': 'group-456',
        'fingerprint': 'abc123def456',
      };

      final response = InceptorCrashResponse.fromJson(json);

      expect(response.isNewGroup, false);
    });
  });
}
