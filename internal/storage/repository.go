package storage

import (
	"context"
	"time"

	"github.com/flakerimi/inceptor/internal/core"
)

// Repository defines the interface for all storage operations
type Repository interface {
	// Crash operations
	CreateCrash(ctx context.Context, crash *core.Crash) error
	GetCrash(ctx context.Context, id string) (*core.Crash, error)
	ListCrashes(ctx context.Context, filter CrashFilter) ([]*core.Crash, int, error)
	DeleteCrash(ctx context.Context, id string) error
	DeleteCrashesOlderThan(ctx context.Context, appID string, before time.Time) (int, error)

	// Crash group operations
	GetOrCreateGroup(ctx context.Context, crash *core.Crash) (*core.CrashGroup, bool, error)
	GetGroup(ctx context.Context, id string) (*core.CrashGroup, error)
	ListGroups(ctx context.Context, filter GroupFilter) ([]*core.CrashGroup, int, error)
	UpdateGroupStatus(ctx context.Context, id string, status string) error
	UpdateGroup(ctx context.Context, group *core.CrashGroup) error
	IncrementGroupCount(ctx context.Context, id string) error

	// App operations
	CreateApp(ctx context.Context, app *core.App) error
	GetApp(ctx context.Context, id string) (*core.App, error)
	GetAppByAPIKey(ctx context.Context, apiKeyHash string) (*core.App, error)
	ListApps(ctx context.Context) ([]*core.App, error)
	UpdateApp(ctx context.Context, app *core.App) error
	UpdateAppAPIKey(ctx context.Context, id string, newKeyHash string) error
	DeleteApp(ctx context.Context, id string) error
	GetAppStats(ctx context.Context, appID string) (*core.CrashStats, error)

	// Alert operations
	CreateAlert(ctx context.Context, alert *core.Alert) error
	GetAlert(ctx context.Context, id string) (*core.Alert, error)
	ListAlerts(ctx context.Context, appID string) ([]*core.Alert, error)
	UpdateAlert(ctx context.Context, alert *core.Alert) error
	DeleteAlert(ctx context.Context, id string) error

	// Settings
	GetSetting(ctx context.Context, key string) (string, error)
	SetSetting(ctx context.Context, key, value string) error

	// Lifecycle
	Close() error
	Migrate() error
}

// CrashFilter defines filters for listing crashes
type CrashFilter struct {
	AppID       string
	GroupID     string
	Platform    string
	Environment string
	ErrorType   string
	UserID      string
	FromDate    *time.Time
	ToDate      *time.Time
	Search      string
	Offset      int
	Limit       int
}

// GroupFilter defines filters for listing crash groups
type GroupFilter struct {
	AppID     string
	Status    string
	ErrorType string
	Search    string
	Offset    int
	Limit     int
	SortBy    string // first_seen, last_seen, occurrence_count
	SortOrder string // asc, desc
}

// FileStore defines the interface for file-based storage
type FileStore interface {
	// SaveCrashLog saves the full crash payload to a file
	SaveCrashLog(ctx context.Context, crash *core.Crash) (string, error)

	// GetCrashLog retrieves the full crash payload from a file
	GetCrashLog(ctx context.Context, filePath string) (*core.Crash, error)

	// DeleteCrashLog deletes a crash log file
	DeleteCrashLog(ctx context.Context, filePath string) error

	// DeleteOldLogs deletes all logs older than the specified date for an app
	DeleteOldLogs(ctx context.Context, appID string, before time.Time) (int, error)

	// GetStorageStats returns storage statistics
	GetStorageStats(ctx context.Context, appID string) (*StorageStats, error)
}

// StorageStats represents storage usage statistics
type StorageStats struct {
	TotalFiles int64 `json:"total_files"`
	TotalSize  int64 `json:"total_size_bytes"`
}
