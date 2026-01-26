/// Inceptor Flutter SDK
///
/// A Flutter SDK for the Inceptor crash logging service.
/// Automatically captures and reports crashes and errors from your Flutter application.
///
/// ## Quick Start
///
/// ```dart
/// import 'package:inceptor_flutter/inceptor_flutter.dart';
///
/// void main() async {
///   WidgetsFlutterBinding.ensureInitialized();
///
///   await Inceptor.init(
///     endpoint: 'https://your-server.com',
///     apiKey: 'your-api-key',
///   );
///
///   // Capture Flutter errors
///   FlutterError.onError = Inceptor.recordFlutterError;
///
///   // Capture async errors
///   runZonedGuarded(() {
///     runApp(MyApp());
///   }, Inceptor.recordError);
/// }
/// ```
library inceptor_flutter;

export 'src/inceptor.dart';
export 'src/navigator_observer.dart';
export 'src/models/crash_report.dart';
export 'src/models/breadcrumb.dart';
export 'src/models/stack_frame.dart';
export 'src/config.dart';
