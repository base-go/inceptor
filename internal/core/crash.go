package core

import (
	"time"
)

// Crash represents a single crash report
type Crash struct {
	ID          string                 `json:"id"`
	AppID       string                 `json:"app_id"`
	AppVersion  string                 `json:"app_version"`
	Platform    string                 `json:"platform"` // ios, android, web, etc.
	OSVersion   string                 `json:"os_version"`
	DeviceModel string                 `json:"device_model"`
	ErrorType   string                 `json:"error_type"`
	ErrorMessage string               `json:"error_message"`
	StackTrace  []StackFrame           `json:"stack_trace"`
	Fingerprint string                 `json:"fingerprint"`
	GroupID     string                 `json:"group_id"`
	UserID      string                 `json:"user_id,omitempty"`
	Environment string                 `json:"environment"` // production, staging, dev
	CreatedAt   time.Time              `json:"created_at"`
	LogFilePath string                 `json:"log_file_path,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Breadcrumbs []Breadcrumb           `json:"breadcrumbs,omitempty"`
}

// StackFrame represents a single frame in a stack trace
type StackFrame struct {
	FileName   string `json:"file_name"`
	LineNumber int    `json:"line_number"`
	ColumnNumber int  `json:"column_number,omitempty"`
	MethodName string `json:"method_name"`
	ClassName  string `json:"class_name,omitempty"`
	Native     bool   `json:"native,omitempty"`
}

// Breadcrumb represents a user action or event leading up to a crash
type Breadcrumb struct {
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"` // navigation, http, user, log
	Category  string                 `json:"category"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Level     string                 `json:"level"` // debug, info, warning, error
}

// CrashGroup represents a group of similar crashes
type CrashGroup struct {
	ID              string    `json:"id"`
	AppID           string    `json:"app_id"`
	Fingerprint     string    `json:"fingerprint"`
	ErrorType       string    `json:"error_type"`
	ErrorMessage    string    `json:"error_message"`
	FirstSeen       time.Time `json:"first_seen"`
	LastSeen        time.Time `json:"last_seen"`
	OccurrenceCount int       `json:"occurrence_count"`
	Status          string    `json:"status"` // open, resolved, ignored
	AssignedTo      string    `json:"assigned_to,omitempty"`
	Notes           string    `json:"notes,omitempty"`
}

// App represents a registered application
type App struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	APIKey        string    `json:"api_key"`
	APIKeyHash    string    `json:"-"` // Stored in DB, not exposed
	CreatedAt     time.Time `json:"created_at"`
	RetentionDays int       `json:"retention_days"`
}

// Alert represents an alert configuration
type Alert struct {
	ID        string                 `json:"id"`
	AppID     string                 `json:"app_id"`
	Type      string                 `json:"type"` // webhook, email, slack
	Config    map[string]interface{} `json:"config"`
	Enabled   bool                   `json:"enabled"`
	CreatedAt time.Time              `json:"created_at"`
}

// CrashStats represents statistics for an app
type CrashStats struct {
	AppID           string         `json:"app_id"`
	TotalCrashes    int            `json:"total_crashes"`
	TotalGroups     int            `json:"total_groups"`
	OpenGroups      int            `json:"open_groups"`
	CrashesLast24h  int            `json:"crashes_last_24h"`
	CrashesLast7d   int            `json:"crashes_last_7d"`
	CrashesLast30d  int            `json:"crashes_last_30d"`
	TopErrors       []ErrorSummary `json:"top_errors"`
	CrashTrend      []TrendPoint   `json:"crash_trend"`
}

// ErrorSummary represents a summary of an error type
type ErrorSummary struct {
	GroupID      string `json:"group_id"`
	ErrorType    string `json:"error_type"`
	ErrorMessage string `json:"error_message"`
	Count        int    `json:"count"`
}

// TrendPoint represents a single point in a crash trend
type TrendPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// CrashSubmission represents the incoming crash report from clients
type CrashSubmission struct {
	AppVersion   string                 `json:"app_version" binding:"required"`
	Platform     string                 `json:"platform" binding:"required"`
	OSVersion    string                 `json:"os_version"`
	DeviceModel  string                 `json:"device_model"`
	ErrorType    string                 `json:"error_type" binding:"required"`
	ErrorMessage string                 `json:"error_message" binding:"required"`
	StackTrace   []StackFrame           `json:"stack_trace" binding:"required"`
	UserID       string                 `json:"user_id,omitempty"`
	Environment  string                 `json:"environment"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Breadcrumbs  []Breadcrumb           `json:"breadcrumbs,omitempty"`
}

// GroupStatus represents valid statuses for crash groups
type GroupStatus string

const (
	GroupStatusOpen     GroupStatus = "open"
	GroupStatusResolved GroupStatus = "resolved"
	GroupStatusIgnored  GroupStatus = "ignored"
)

// Platform constants
const (
	PlatformIOS     = "ios"
	PlatformAndroid = "android"
	PlatformWeb     = "web"
	PlatformDesktop = "desktop"
	PlatformFlutter = "flutter"
)

// Environment constants
const (
	EnvironmentProduction  = "production"
	EnvironmentStaging     = "staging"
	EnvironmentDevelopment = "development"
)
