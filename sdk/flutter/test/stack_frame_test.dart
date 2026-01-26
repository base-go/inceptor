import 'package:flutter_test/flutter_test.dart';
import 'package:inceptor_flutter/inceptor_flutter.dart';

void main() {
  group('InceptorStackFrame', () {
    test('parses standard Dart stack trace line', () {
      const line =
          '#0      MyClass.myMethod (package:myapp/src/my_class.dart:42:10)';
      final frame = InceptorStackFrame.fromString(line);

      expect(frame.className, 'MyClass');
      expect(frame.methodName, 'myMethod');
      expect(frame.fileName, 'package:myapp/src/my_class.dart');
      expect(frame.lineNumber, 42);
      expect(frame.columnNumber, 10);
      expect(frame.native, false);
    });

    test('parses function without class', () {
      const line = '#1      topLevelFunction (package:myapp/main.dart:10:5)';
      final frame = InceptorStackFrame.fromString(line);

      expect(frame.methodName, 'topLevelFunction');
      expect(frame.fileName, 'package:myapp/main.dart');
      expect(frame.lineNumber, 10);
    });

    test('identifies native frames', () {
      const line = '#2      _asyncRunCallback (dart:async/future.dart:100:10)';
      final frame = InceptorStackFrame.fromString(line);

      expect(frame.native, true);
      expect(frame.fileName.startsWith('dart:'), true);
    });

    test('parses nested class method', () {
      const line =
          '#0      _MyClassState._handleTap (package:app/widget.dart:50:5)';
      final frame = InceptorStackFrame.fromString(line);

      expect(frame.className, '_MyClassState');
      expect(frame.methodName, '_handleTap');
    });

    test('handles malformed line gracefully', () {
      const line = 'some random text';
      final frame = InceptorStackFrame.fromString(line);

      expect(frame.fileName, 'unknown');
      expect(frame.lineNumber, 0);
    });

    test('parseStackTrace extracts all frames', () {
      final stackTrace = StackTrace.fromString('''
#0      MyClass.method1 (package:app/a.dart:10:5)
#1      MyClass.method2 (package:app/b.dart:20:10)
#2      main (package:app/main.dart:5:3)
''');

      final frames = InceptorStackFrame.parseStackTrace(stackTrace);

      expect(frames.length, 3);
      expect(frames[0].methodName, 'method1');
      expect(frames[1].methodName, 'method2');
      expect(frames[2].methodName, 'main');
    });

    test('toJson produces correct format', () {
      final frame = InceptorStackFrame(
        fileName: 'package:app/main.dart',
        lineNumber: 42,
        columnNumber: 10,
        methodName: 'doSomething',
        className: 'MyClass',
        native: false,
      );

      final json = frame.toJson();

      expect(json['file_name'], 'package:app/main.dart');
      expect(json['line_number'], 42);
      expect(json['column_number'], 10);
      expect(json['method_name'], 'doSomething');
      expect(json['class_name'], 'MyClass');
      expect(json['native'], false);
    });

    test('fromJson reconstructs frame correctly', () {
      final json = {
        'file_name': 'package:app/main.dart',
        'line_number': 42,
        'column_number': 10,
        'method_name': 'doSomething',
        'class_name': 'MyClass',
        'native': false,
      };

      final frame = InceptorStackFrame.fromJson(json);

      expect(frame.fileName, 'package:app/main.dart');
      expect(frame.lineNumber, 42);
      expect(frame.columnNumber, 10);
      expect(frame.methodName, 'doSomething');
      expect(frame.className, 'MyClass');
      expect(frame.native, false);
    });
  });
}
