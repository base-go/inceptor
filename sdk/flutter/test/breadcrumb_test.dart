import 'package:flutter_test/flutter_test.dart';
import 'package:inceptor_flutter/inceptor_flutter.dart';

void main() {
  group('InceptorBreadcrumb', () {
    test('navigation breadcrumb creates correct structure', () {
      final breadcrumb = InceptorBreadcrumb.navigation(
        from: '/home',
        to: '/settings',
      );

      expect(breadcrumb.type, 'navigation');
      expect(breadcrumb.category, 'navigation');
      expect(breadcrumb.message, 'Navigated from /home to /settings');
      expect(breadcrumb.data?['from'], '/home');
      expect(breadcrumb.data?['to'], '/settings');
      expect(breadcrumb.level, 'info');
    });

    test('http breadcrumb with success status', () {
      final breadcrumb = InceptorBreadcrumb.http(
        method: 'GET',
        url: 'https://api.example.com/users',
        statusCode: 200,
      );

      expect(breadcrumb.type, 'http');
      expect(breadcrumb.category, 'http');
      expect(breadcrumb.message, 'GET https://api.example.com/users (200)');
      expect(breadcrumb.data?['method'], 'GET');
      expect(breadcrumb.data?['status_code'], 200);
      expect(breadcrumb.level, 'info');
    });

    test('http breadcrumb with error status', () {
      final breadcrumb = InceptorBreadcrumb.http(
        method: 'POST',
        url: 'https://api.example.com/orders',
        statusCode: 500,
        reason: 'Internal Server Error',
      );

      expect(breadcrumb.level, 'error');
      expect(breadcrumb.data?['reason'], 'Internal Server Error');
    });

    test('user breadcrumb with target', () {
      final breadcrumb = InceptorBreadcrumb.user(
        action: 'clicked',
        target: 'submit_button',
        data: {'form_id': 'checkout'},
      );

      expect(breadcrumb.type, 'user');
      expect(breadcrumb.category, 'user_action');
      expect(breadcrumb.message, 'clicked on submit_button');
      expect(breadcrumb.data?['action'], 'clicked');
      expect(breadcrumb.data?['target'], 'submit_button');
      expect(breadcrumb.data?['form_id'], 'checkout');
    });

    test('log breadcrumb', () {
      final breadcrumb = InceptorBreadcrumb.log(
        message: 'User logged in',
        level: 'info',
        data: {'user_id': '123'},
      );

      expect(breadcrumb.type, 'log');
      expect(breadcrumb.category, 'console');
      expect(breadcrumb.message, 'User logged in');
      expect(breadcrumb.level, 'info');
    });

    test('custom breadcrumb', () {
      final breadcrumb = InceptorBreadcrumb.custom(
        type: 'state',
        category: 'redux',
        message: 'State changed',
        data: {'action': 'ADD_ITEM'},
        level: 'debug',
      );

      expect(breadcrumb.type, 'state');
      expect(breadcrumb.category, 'redux');
      expect(breadcrumb.message, 'State changed');
      expect(breadcrumb.level, 'debug');
    });

    test('toJson produces correct format', () {
      final timestamp = DateTime.utc(2024, 1, 15, 10, 30, 0);
      final breadcrumb = InceptorBreadcrumb(
        timestamp: timestamp,
        type: 'navigation',
        category: 'navigation',
        message: 'Test message',
        data: {'key': 'value'},
        level: 'info',
      );

      final json = breadcrumb.toJson();

      expect(json['timestamp'], '2024-01-15T10:30:00.000Z');
      expect(json['type'], 'navigation');
      expect(json['category'], 'navigation');
      expect(json['message'], 'Test message');
      expect(json['data'], {'key': 'value'});
      expect(json['level'], 'info');
    });

    test('fromJson reconstructs breadcrumb correctly', () {
      final json = {
        'timestamp': '2024-01-15T10:30:00.000Z',
        'type': 'navigation',
        'category': 'navigation',
        'message': 'Test message',
        'data': {'key': 'value'},
        'level': 'info',
      };

      final breadcrumb = InceptorBreadcrumb.fromJson(json);

      expect(breadcrumb.timestamp, DateTime.utc(2024, 1, 15, 10, 30, 0));
      expect(breadcrumb.type, 'navigation');
      expect(breadcrumb.category, 'navigation');
      expect(breadcrumb.message, 'Test message');
      expect(breadcrumb.data, {'key': 'value'});
      expect(breadcrumb.level, 'info');
    });
  });
}
