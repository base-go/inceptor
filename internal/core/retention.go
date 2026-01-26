package core

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// RetentionManager handles automatic cleanup of old crash data
type RetentionManager struct {
	repo        RetentionRepository
	fileStore   RetentionFileStore
	defaultDays int
	interval    time.Duration
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// RetentionRepository defines the database operations needed for retention
type RetentionRepository interface {
	ListApps(ctx context.Context) ([]*App, error)
	DeleteCrashesOlderThan(ctx context.Context, appID string, before time.Time) (int, error)
}

// RetentionFileStore defines the file operations needed for retention
type RetentionFileStore interface {
	DeleteOldLogs(ctx context.Context, appID string, before time.Time) (int, error)
}

// NewRetentionManager creates a new RetentionManager
func NewRetentionManager(repo RetentionRepository, fileStore RetentionFileStore, defaultDays int, interval time.Duration) *RetentionManager {
	ctx, cancel := context.WithCancel(context.Background())

	rm := &RetentionManager{
		repo:        repo,
		fileStore:   fileStore,
		defaultDays: defaultDays,
		interval:    interval,
		ctx:         ctx,
		cancel:      cancel,
	}

	return rm
}

// Start begins the retention cleanup worker
func (rm *RetentionManager) Start() {
	rm.wg.Add(1)
	go rm.worker()
	log.Info().Dur("interval", rm.interval).Msg("Retention manager started")
}

// Stop gracefully stops the retention manager
func (rm *RetentionManager) Stop() {
	rm.cancel()
	rm.wg.Wait()
	log.Info().Msg("Retention manager stopped")
}

// worker runs the periodic cleanup
func (rm *RetentionManager) worker() {
	defer rm.wg.Done()

	// Run immediately on start
	rm.cleanup()

	ticker := time.NewTicker(rm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.cleanup()
		}
	}
}

// cleanup performs the actual cleanup of old data
func (rm *RetentionManager) cleanup() {
	log.Info().Msg("Starting retention cleanup")
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(rm.ctx, 30*time.Minute)
	defer cancel()

	// Get all apps
	apps, err := rm.repo.ListApps(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list apps for retention cleanup")
		return
	}

	totalDBDeleted := 0
	totalFilesDeleted := 0

	for _, app := range apps {
		// Determine retention period for this app
		retentionDays := app.RetentionDays
		if retentionDays <= 0 {
			retentionDays = rm.defaultDays
		}

		cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

		// Delete from database
		dbDeleted, err := rm.repo.DeleteCrashesOlderThan(ctx, app.ID, cutoffDate)
		if err != nil {
			log.Error().Err(err).Str("app_id", app.ID).Msg("Failed to delete old crashes from database")
		} else {
			totalDBDeleted += dbDeleted
		}

		// Delete log files
		filesDeleted, err := rm.fileStore.DeleteOldLogs(ctx, app.ID, cutoffDate)
		if err != nil {
			log.Error().Err(err).Str("app_id", app.ID).Msg("Failed to delete old crash log files")
		} else {
			totalFilesDeleted += filesDeleted
		}

		if dbDeleted > 0 || filesDeleted > 0 {
			log.Info().
				Str("app_id", app.ID).
				Int("retention_days", retentionDays).
				Int("db_deleted", dbDeleted).
				Int("files_deleted", filesDeleted).
				Msg("Cleaned up old crashes for app")
		}
	}

	duration := time.Since(startTime)
	log.Info().
		Dur("duration", duration).
		Int("total_db_deleted", totalDBDeleted).
		Int("total_files_deleted", totalFilesDeleted).
		Msg("Retention cleanup completed")
}

// RunNow triggers an immediate cleanup (useful for testing or manual triggering)
func (rm *RetentionManager) RunNow() {
	go rm.cleanup()
}

// CleanupApp cleans up data for a specific app (useful when deleting an app)
func (rm *RetentionManager) CleanupApp(ctx context.Context, appID string) error {
	// Delete all crashes for this app
	_, err := rm.repo.DeleteCrashesOlderThan(ctx, appID, time.Now().Add(time.Hour))
	if err != nil {
		return err
	}

	// Delete all log files for this app
	_, err = rm.fileStore.DeleteOldLogs(ctx, appID, time.Now().Add(time.Hour))
	return err
}
