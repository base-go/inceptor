package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

// Grouper handles crash fingerprinting and grouping logic
type Grouper struct {
	// Number of stack frames to use for fingerprinting
	FrameLimit int
}

// NewGrouper creates a new Grouper with default settings
func NewGrouper() *Grouper {
	return &Grouper{
		FrameLimit: 5,
	}
}

// GenerateFingerprint creates a unique fingerprint for a crash
// This is used to group similar crashes together
func (g *Grouper) GenerateFingerprint(crash *Crash) string {
	h := sha256.New()

	// Include error type
	h.Write([]byte(crash.ErrorType))
	h.Write([]byte("|"))

	// Include normalized stack frames
	frameCount := g.FrameLimit
	if len(crash.StackTrace) < frameCount {
		frameCount = len(crash.StackTrace)
	}

	for i := 0; i < frameCount; i++ {
		frame := crash.StackTrace[i]
		// Skip native/system frames for more consistent grouping
		if frame.Native {
			continue
		}

		normalized := g.normalizeFrame(frame)
		h.Write([]byte(normalized))
		h.Write([]byte("|"))
	}

	// Return first 16 characters of hex-encoded hash
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// normalizeFrame normalizes a stack frame for consistent fingerprinting
// Removes variable parts like line numbers, memory addresses, and closure IDs
func (g *Grouper) normalizeFrame(frame StackFrame) string {
	var parts []string

	// Include class name if present
	if frame.ClassName != "" {
		parts = append(parts, normalizeClassName(frame.ClassName))
	}

	// Include method name
	if frame.MethodName != "" {
		parts = append(parts, normalizeMethodName(frame.MethodName))
	}

	// Include file name (without path and extension variations)
	if frame.FileName != "" {
		parts = append(parts, normalizeFileName(frame.FileName))
	}

	return strings.Join(parts, ":")
}

// normalizeClassName normalizes a class name
func normalizeClassName(className string) string {
	// Remove generic type parameters
	re := regexp.MustCompile(`<[^>]+>`)
	className = re.ReplaceAllString(className, "")

	// Remove anonymous class indicators (e.g., $1, $anon)
	re = regexp.MustCompile(`\$\d+|\$anon\w*`)
	className = re.ReplaceAllString(className, "")

	return className
}

// normalizeMethodName normalizes a method name
func normalizeMethodName(methodName string) string {
	// Remove closure/lambda identifiers
	re := regexp.MustCompile(`_closure\d*|\$\d+|_\d+$`)
	methodName = re.ReplaceAllString(methodName, "")

	// Remove async state machine markers
	methodName = strings.TrimSuffix(methodName, "_async")

	return methodName
}

// normalizeFileName normalizes a file name
func normalizeFileName(fileName string) string {
	// Extract just the filename without path
	parts := strings.Split(fileName, "/")
	fileName = parts[len(parts)-1]
	parts = strings.Split(fileName, "\\")
	fileName = parts[len(parts)-1]

	// Remove query strings and hashes (for web)
	if idx := strings.Index(fileName, "?"); idx != -1 {
		fileName = fileName[:idx]
	}
	if idx := strings.Index(fileName, "#"); idx != -1 {
		fileName = fileName[:idx]
	}

	// Remove common build hashes
	re := regexp.MustCompile(`\.[a-f0-9]{8,}\.(js|dart|ts)$`)
	if re.MatchString(fileName) {
		fileName = re.ReplaceAllString(fileName, ".$1")
	}

	return fileName
}

// IsSimilar checks if two crashes are similar enough to be in the same group
func (g *Grouper) IsSimilar(crash1, crash2 *Crash) bool {
	return g.GenerateFingerprint(crash1) == g.GenerateFingerprint(crash2)
}

// ExtractErrorSummary creates a short summary of the error
func ExtractErrorSummary(crash *Crash) string {
	// Truncate error message to reasonable length
	message := crash.ErrorMessage
	if len(message) > 200 {
		message = message[:200] + "..."
	}
	return message
}

// GetTopFrame returns the most relevant stack frame (usually the first non-system frame)
func GetTopFrame(crash *Crash) *StackFrame {
	for i := range crash.StackTrace {
		frame := &crash.StackTrace[i]
		// Skip native/system frames
		if frame.Native {
			continue
		}
		// Skip common framework frames
		if isFrameworkFrame(frame) {
			continue
		}
		return frame
	}

	// Fall back to first frame
	if len(crash.StackTrace) > 0 {
		return &crash.StackTrace[0]
	}
	return nil
}

// isFrameworkFrame checks if a frame is from a common framework
func isFrameworkFrame(frame *StackFrame) bool {
	frameworkPatterns := []string{
		"dart:async",
		"dart:core",
		"package:flutter/",
		"java.lang.",
		"android.os.",
		"kotlinx.coroutines",
		"react-dom",
		"zone.js",
		"angular",
	}

	fullPath := frame.FileName
	if frame.ClassName != "" {
		fullPath = frame.ClassName
	}

	for _, pattern := range frameworkPatterns {
		if strings.Contains(fullPath, pattern) {
			return true
		}
	}

	return false
}

// ParseFlutterStackTrace parses a Flutter/Dart stack trace string into StackFrames
func ParseFlutterStackTrace(stackTrace string) []StackFrame {
	var frames []StackFrame

	lines := strings.Split(stackTrace, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		frame := parseFlutterFrame(line)
		if frame != nil {
			frames = append(frames, *frame)
		}
	}

	return frames
}

// parseFlutterFrame parses a single Flutter stack trace line
func parseFlutterFrame(line string) *StackFrame {
	// Flutter format: #0      methodName (package:path/file.dart:line:col)
	// Or: #0      className.methodName (package:path/file.dart:line:col)

	// Remove frame number prefix
	re := regexp.MustCompile(`^#\d+\s+`)
	line = re.ReplaceAllString(line, "")

	// Extract method and location
	parts := strings.SplitN(line, " (", 2)
	if len(parts) != 2 {
		return nil
	}

	methodPart := parts[0]
	locationPart := strings.TrimSuffix(parts[1], ")")

	frame := &StackFrame{}

	// Parse method/class
	if strings.Contains(methodPart, ".") {
		lastDot := strings.LastIndex(methodPart, ".")
		frame.ClassName = methodPart[:lastDot]
		frame.MethodName = methodPart[lastDot+1:]
	} else {
		frame.MethodName = methodPart
	}

	// Parse location (file:line:col)
	locParts := strings.Split(locationPart, ":")
	if len(locParts) >= 1 {
		frame.FileName = locParts[0]
	}
	if len(locParts) >= 2 {
		fmt.Sscanf(locParts[1], "%d", &frame.LineNumber)
	}
	if len(locParts) >= 3 {
		fmt.Sscanf(locParts[2], "%d", &frame.ColumnNumber)
	}

	// Check if it's a native/dart frame
	frame.Native = strings.HasPrefix(frame.FileName, "dart:")

	return frame
}
