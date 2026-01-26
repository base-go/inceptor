import 'package:flutter_test/flutter_test.dart';
import 'package:inceptor_flutter/inceptor_flutter.dart';

void main() {
  group('InceptorConfig', () {
    test('has correct defaults', () {
      final config = InceptorConfig(
        endpoint: 'https://example.com',
        apiKey: 'test-key',
      );

      expect(config.endpoint, 'https://example.com');
      expect(config.apiKey, 'test-key');
      expect(config.appId, isNull);
      expect(config.environment, 'production');
      expect(config.debug, false);
      expect(config.maxBreadcrumbs, 50);
      expect(config.captureScreenshots, false);
      expect(config.timeout, 30000);
      expect(config.enableOfflineQueue, true);
      expect(config.maxOfflineQueueSize, 100);
      expect(config.tags, isNull);
    });

    test('accepts custom values', () {
      final config = InceptorConfig(
        endpoint: 'https://custom.com',
        apiKey: 'custom-key',
        appId: 'my-app',
        environment: 'staging',
        debug: true,
        maxBreadcrumbs: 100,
        captureScreenshots: true,
        timeout: 60000,
        enableOfflineQueue: false,
        maxOfflineQueueSize: 50,
        tags: {'version': '1.0'},
      );

      expect(config.endpoint, 'https://custom.com');
      expect(config.apiKey, 'custom-key');
      expect(config.appId, 'my-app');
      expect(config.environment, 'staging');
      expect(config.debug, true);
      expect(config.maxBreadcrumbs, 100);
      expect(config.captureScreenshots, true);
      expect(config.timeout, 60000);
      expect(config.enableOfflineQueue, false);
      expect(config.maxOfflineQueueSize, 50);
      expect(config.tags, {'version': '1.0'});
    });

    test('copyWith creates new instance with changes', () {
      final original = InceptorConfig(
        endpoint: 'https://example.com',
        apiKey: 'test-key',
        environment: 'production',
      );

      final modified = original.copyWith(
        environment: 'staging',
        debug: true,
      );

      // Original unchanged
      expect(original.environment, 'production');
      expect(original.debug, false);

      // New instance has changes
      expect(modified.environment, 'staging');
      expect(modified.debug, true);

      // Unmodified fields copied
      expect(modified.endpoint, 'https://example.com');
      expect(modified.apiKey, 'test-key');
    });

    test('copyWith preserves all fields when not changed', () {
      final original = InceptorConfig(
        endpoint: 'https://example.com',
        apiKey: 'test-key',
        appId: 'app-id',
        environment: 'staging',
        debug: true,
        maxBreadcrumbs: 75,
        captureScreenshots: true,
        timeout: 45000,
        enableOfflineQueue: false,
        maxOfflineQueueSize: 25,
        tags: {'tag': 'value'},
      );

      final copy = original.copyWith();

      expect(copy.endpoint, original.endpoint);
      expect(copy.apiKey, original.apiKey);
      expect(copy.appId, original.appId);
      expect(copy.environment, original.environment);
      expect(copy.debug, original.debug);
      expect(copy.maxBreadcrumbs, original.maxBreadcrumbs);
      expect(copy.captureScreenshots, original.captureScreenshots);
      expect(copy.timeout, original.timeout);
      expect(copy.enableOfflineQueue, original.enableOfflineQueue);
      expect(copy.maxOfflineQueueSize, original.maxOfflineQueueSize);
      expect(copy.tags, original.tags);
    });
  });
}
