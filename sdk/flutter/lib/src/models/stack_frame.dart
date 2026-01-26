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

    // Parse location - handle URIs like package:path/file.dart:line:col
    // Split from the end to find line and column numbers
    String fileName = 'unknown';
    int lineNumber = 0;
    int? columnNumber;

    // Find the last two colons which separate line:col
    final lastColon = locationPart.lastIndexOf(':');
    if (lastColon > 0) {
      final beforeLastColon = locationPart.substring(0, lastColon);
      final secondLastColon = beforeLastColon.lastIndexOf(':');

      if (secondLastColon > 0) {
        // We have file:line:col format
        final possibleCol = int.tryParse(locationPart.substring(lastColon + 1));
        final possibleLine =
            int.tryParse(beforeLastColon.substring(secondLastColon + 1));

        if (possibleLine != null) {
          fileName = beforeLastColon.substring(0, secondLastColon);
          lineNumber = possibleLine;
          columnNumber = possibleCol;
        } else {
          // Might be file:line format (no column)
          final possibleLineOnly =
              int.tryParse(locationPart.substring(lastColon + 1));
          if (possibleLineOnly != null) {
            fileName = beforeLastColon;
            lineNumber = possibleLineOnly;
          } else {
            fileName = locationPart;
          }
        }
      } else {
        // Might be file:line format
        final possibleLine =
            int.tryParse(locationPart.substring(lastColon + 1));
        if (possibleLine != null) {
          fileName = beforeLastColon;
          lineNumber = possibleLine;
        } else {
          fileName = locationPart;
        }
      }
    } else {
      fileName = locationPart;
    }

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

  /// Create from JSON
  factory InceptorStackFrame.fromJson(Map<String, dynamic> json) {
    return InceptorStackFrame(
      fileName: json['file_name'] as String,
      lineNumber: json['line_number'] as int,
      columnNumber: json['column_number'] as int?,
      methodName: json['method_name'] as String,
      className: json['class_name'] as String?,
      native: json['native'] as bool? ?? false,
    );
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
