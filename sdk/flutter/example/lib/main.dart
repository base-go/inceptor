import 'dart:async';

import 'package:flutter/material.dart';
import 'package:inceptor_flutter/inceptor_flutter.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize Inceptor
  await Inceptor.init(
    endpoint: 'http://localhost:8080', // Your Inceptor server
    apiKey: 'your-api-key', // Your app's API key
    environment: 'development',
    debug: true, // Enable debug logging
  );

  // Set up Flutter error handling
  FlutterError.onError = Inceptor.recordFlutterError;

  // Catch async errors
  runZonedGuarded(() {
    runApp(const MyApp());
  }, Inceptor.recordError);
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Inceptor Demo',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepPurple),
        useMaterial3: true,
      ),
      navigatorObservers: [InceptorNavigatorObserver()],
      home: const HomePage(),
    );
  }
}

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  @override
  void initState() {
    super.initState();

    // Set user context
    Inceptor.setUser('user_123');
    Inceptor.setMetadata('subscription', 'premium');

    // Add initial breadcrumb
    Inceptor.addBreadcrumb(
      InceptorBreadcrumb.custom(
        type: 'lifecycle',
        category: 'app',
        message: 'App started',
      ),
    );
  }

  void _triggerSyncError() {
    // This will be caught by FlutterError.onError
    throw Exception('Test sync error from button press');
  }

  Future<void> _triggerAsyncError() async {
    // Add breadcrumb before action
    Inceptor.addUserBreadcrumb(
      action: 'pressed',
      target: 'Async Error Button',
    );

    // Simulate async operation
    await Future.delayed(const Duration(milliseconds: 500));

    // This will be caught by runZonedGuarded
    throw Exception('Test async error');
  }

  Future<void> _triggerManualCapture() async {
    try {
      // Simulate some operation that fails
      final result = await _riskyOperation();
      debugPrint('Result: $result');
    } catch (e, stackTrace) {
      // Manually capture with context
      final response = await Inceptor.captureException(
        e,
        stackTrace: stackTrace,
        context: {
          'operation': 'risky_operation',
          'retry_count': 3,
        },
      );

      if (response != null) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              'Error captured! ID: ${response.id.substring(0, 8)}...',
            ),
          ),
        );
      }
    }
  }

  Future<int> _riskyOperation() async {
    await Future.delayed(const Duration(milliseconds: 100));
    throw FormatException('Invalid data format');
  }

  void _captureMessage() {
    Inceptor.captureMessage(
      'User reached important milestone',
      level: 'info',
      context: {'milestone': 'first_purchase'},
    );

    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Message captured!')),
    );
  }

  void _simulateHttpRequest() {
    // Simulate HTTP breadcrumb
    Inceptor.addHttpBreadcrumb(
      method: 'GET',
      url: 'https://api.example.com/users',
      statusCode: 200,
    );

    Inceptor.addHttpBreadcrumb(
      method: 'POST',
      url: 'https://api.example.com/orders',
      statusCode: 500,
      reason: 'Internal Server Error',
    );

    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('HTTP breadcrumbs added!')),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        title: const Text('Inceptor Demo'),
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          _buildInfoCard(),
          const SizedBox(height: 16),
          _buildSection(
            'Error Capture',
            [
              _buildButton(
                'Trigger Sync Error',
                Icons.error,
                Colors.red,
                _triggerSyncError,
              ),
              _buildButton(
                'Trigger Async Error',
                Icons.schedule,
                Colors.orange,
                _triggerAsyncError,
              ),
              _buildButton(
                'Manual Capture',
                Icons.pan_tool,
                Colors.blue,
                _triggerManualCapture,
              ),
              _buildButton(
                'Capture Message',
                Icons.message,
                Colors.green,
                _captureMessage,
              ),
            ],
          ),
          const SizedBox(height: 16),
          _buildSection(
            'Breadcrumbs',
            [
              _buildButton(
                'Add HTTP Breadcrumbs',
                Icons.http,
                Colors.purple,
                _simulateHttpRequest,
              ),
              _buildButton(
                'Navigate to Details',
                Icons.arrow_forward,
                Colors.teal,
                () => Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => const DetailsPage(),
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildInfoCard() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.info, color: Theme.of(context).colorScheme.primary),
                const SizedBox(width: 8),
                const Text(
                  'Inceptor SDK Demo',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            const Text(
              'This demo shows different ways to capture errors and add context to your crash reports.',
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSection(String title, List<Widget> children) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          title,
          style: const TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.bold,
          ),
        ),
        const SizedBox(height: 8),
        ...children,
      ],
    );
  }

  Widget _buildButton(
    String label,
    IconData icon,
    Color color,
    VoidCallback onPressed,
  ) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: ElevatedButton.icon(
        onPressed: onPressed,
        icon: Icon(icon),
        label: Text(label),
        style: ElevatedButton.styleFrom(
          backgroundColor: color,
          foregroundColor: Colors.white,
          minimumSize: const Size(double.infinity, 48),
        ),
      ),
    );
  }
}

class DetailsPage extends StatelessWidget {
  const DetailsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Details')),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Text('This is the details page'),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: () {
                Inceptor.addUserBreadcrumb(
                  action: 'pressed',
                  target: 'Error Button on Details',
                );
                throw Exception('Error from details page');
              },
              child: const Text('Trigger Error Here'),
            ),
          ],
        ),
      ),
    );
  }
}
