/// Represents a single frame in a stack trace
class InceptorStackFrame {
  /// The file name where the error occurred
  final String fileName;

  /// The line number
  final int lineNumber;

  /// The column number (if available)
  final int? columnNumber;

  /// The method/function name
  final String methodName;

  /// The class name (if available)
  final String? className;

  /// Whether this is a native/system frame
  final bool native;

  const InceptorStackFrame({
    required this.fileName,
    required this.lineNumber,
    this.columnNumber,
    required this.methodName,
    this.className,
    this.native = false,
  });

  /// Create from a Dart StackTrace line
  factory InceptorStackFrame.fromString(String line) {
    // Parse Dart stack trace format:
    // #0      methodName (package:path/file.dart:line:col)
    // #0      ClassName.methodName (package:path/file.dart:line:col)

    final trimmed = line.trim();

    // Remove frame number prefix
    final withoutNumber = trimmed.replaceFirst(RegExp(r'^#\d+\s+'), '');

    // Split method and location
    final parenIndex = withoutNumber.lastIndexOf('(');
    if (parenIndex == -1) {
      return InceptorStackFrame(
        fileName: 'unknown',
        lineNumber: 0,
        methodName: withoutNumber,
      );
    }

    final methodPart = withoutNumber.substring(0, parenIndex).trim();
    var locationPart =
        withoutNumber.substring(parenIndex + 1).replaceAll(')', '');

    // Parse method/class
    String? className;
    String methodName;
    if (methodPart.contains('.')) {
      final lastDot = methodPart.lastIndexOf('.');
      className = methodPart.substring(0, lastDot);
      methodName = methodPart.substring(lastDot + 1);
    } else {
      methodName = methodPart;
    }

    // Parse location
    final locParts = locationPart.split(':');
    final fileName = locParts.isNotEmpty ? locParts[0] : 'unknown';
    final lineNumber =
        locParts.length > 1 ? int.tryParse(locParts[1]) ?? 0 : 0;
    final columnNumber =
        locParts.length > 2 ? int.tryParse(locParts[2]) : null;

    return InceptorStackFrame(
      fileName: fileName,
      lineNumber: lineNumber,
      columnNumber: columnNumber,
      methodName: methodName,
      className: className,
      native: fileName.startsWith('dart:'),
    );
  }

  /// Parse a full StackTrace into frames
  static List<InceptorStackFrame> parseStackTrace(StackTrace stackTrace) {
    final lines = stackTrace.toString().split('\n');
    return lines
        .where((line) => line.trim().isNotEmpty && line.trim().startsWith('#'))
        .map((line) => InceptorStackFrame.fromString(line))
        .toList();
  }

  /// Convert to JSON for API submission
  Map<String, dynamic> toJson() {
    return {
      'file_name': fileName,
      'line_number': lineNumber,
      if (columnNumber != null) 'column_number': columnNumber,
      'method_name': methodName,
      if (className != null) 'class_name': className,
      'native': native,
    };
  }

  @override
  String toString() {
    final classPrefix = className != null ? '$className.' : '';
    final location = columnNumber != null
        ? '$fileName:$lineNumber:$columnNumber'
        : '$fileName:$lineNumber';
    return '$classPrefix$methodName ($location)';
  }
}
