package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/flakerimi/inceptor/internal/core"
	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1) // SQLite only supports one writer
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	repo := &SQLiteRepository{db: db}
	if err := repo.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return repo, nil
}

func (r *SQLiteRepository) Migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS apps (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			api_key_hash TEXT UNIQUE NOT NULL,
			created_at DATETIME NOT NULL,
			retention_days INTEGER DEFAULT 30
		)`,
		`CREATE TABLE IF NOT EXISTS crash_groups (
			id TEXT PRIMARY KEY,
			app_id TEXT NOT NULL,
			fingerprint TEXT NOT NULL,
			error_type TEXT,
			error_message TEXT,
			first_seen DATETIME NOT NULL,
			last_seen DATETIME NOT NULL,
			occurrence_count INTEGER DEFAULT 1,
			status TEXT DEFAULT 'open',
			assigned_to TEXT,
			notes TEXT,
			FOREIGN KEY (app_id) REFERENCES apps(id),
			UNIQUE(app_id, fingerprint)
		)`,
		`CREATE TABLE IF NOT EXISTS crashes (
			id TEXT PRIMARY KEY,
			app_id TEXT NOT NULL,
			app_version TEXT,
			platform TEXT,
			os_version TEXT,
			device_model TEXT,
			error_type TEXT,
			error_message TEXT,
			fingerprint TEXT NOT NULL,
			group_id TEXT,
			user_id TEXT,
			environment TEXT,
			created_at DATETIME NOT NULL,
			log_file_path TEXT,
			metadata TEXT,
			FOREIGN KEY (app_id) REFERENCES apps(id),
			FOREIGN KEY (group_id) REFERENCES crash_groups(id)
		)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id TEXT PRIMARY KEY,
			app_id TEXT NOT NULL,
			type TEXT NOT NULL,
			config TEXT,
			enabled INTEGER DEFAULT 1,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (app_id) REFERENCES apps(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_crashes_app_id ON crashes(app_id)`,
		`CREATE INDEX IF NOT EXISTS idx_crashes_group_id ON crashes(group_id)`,
		`CREATE INDEX IF NOT EXISTS idx_crashes_created_at ON crashes(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_crashes_fingerprint ON crashes(fingerprint)`,
		`CREATE INDEX IF NOT EXISTS idx_crash_groups_app_id ON crash_groups(app_id)`,
		`CREATE INDEX IF NOT EXISTS idx_crash_groups_fingerprint ON crash_groups(app_id, fingerprint)`,
		`CREATE INDEX IF NOT EXISTS idx_crash_groups_status ON crash_groups(status)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
	}

	for _, migration := range migrations {
		if _, err := r.db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

// App operations
func (r *SQLiteRepository) CreateApp(ctx context.Context, app *core.App) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO apps (id, name, api_key_hash, created_at, retention_days) VALUES (?, ?, ?, ?, ?)`,
		app.ID, app.Name, app.APIKeyHash, app.CreatedAt, app.RetentionDays,
	)
	return err
}

func (r *SQLiteRepository) GetApp(ctx context.Context, id string) (*core.App, error) {
	app := &core.App{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, api_key_hash, created_at, retention_days FROM apps WHERE id = ?`, id,
	).Scan(&app.ID, &app.Name, &app.APIKeyHash, &app.CreatedAt, &app.RetentionDays)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return app, err
}

func (r *SQLiteRepository) GetAppByAPIKey(ctx context.Context, apiKeyHash string) (*core.App, error) {
	app := &core.App{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, api_key_hash, created_at, retention_days FROM apps WHERE api_key_hash = ?`, apiKeyHash,
	).Scan(&app.ID, &app.Name, &app.APIKeyHash, &app.CreatedAt, &app.RetentionDays)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return app, err
}

func (r *SQLiteRepository) ListApps(ctx context.Context) ([]*core.App, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, api_key_hash, created_at, retention_days FROM apps ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []*core.App
	for rows.Next() {
		app := &core.App{}
		if err := rows.Scan(&app.ID, &app.Name, &app.APIKeyHash, &app.CreatedAt, &app.RetentionDays); err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	return apps, rows.Err()
}

func (r *SQLiteRepository) UpdateApp(ctx context.Context, app *core.App) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE apps SET name = ?, retention_days = ? WHERE id = ?`,
		app.Name, app.RetentionDays, app.ID,
	)
	return err
}

func (r *SQLiteRepository) DeleteApp(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete alerts first
	if _, err := tx.ExecContext(ctx, `DELETE FROM alerts WHERE app_id = ?`, id); err != nil {
		return err
	}

	// Delete crashes
	if _, err := tx.ExecContext(ctx, `DELETE FROM crashes WHERE app_id = ?`, id); err != nil {
		return err
	}

	// Delete crash groups
	if _, err := tx.ExecContext(ctx, `DELETE FROM crash_groups WHERE app_id = ?`, id); err != nil {
		return err
	}

	// Delete app
	if _, err := tx.ExecContext(ctx, `DELETE FROM apps WHERE id = ?`, id); err != nil {
		return err
	}

	return tx.Commit()
}

// Crash operations
func (r *SQLiteRepository) CreateCrash(ctx context.Context, crash *core.Crash) error {
	metadata, _ := json.Marshal(crash.Metadata)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO crashes (id, app_id, app_version, platform, os_version, device_model, error_type, error_message, fingerprint, group_id, user_id, environment, created_at, log_file_path, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		crash.ID, crash.AppID, crash.AppVersion, crash.Platform, crash.OSVersion, crash.DeviceModel,
		crash.ErrorType, crash.ErrorMessage, crash.Fingerprint, crash.GroupID, crash.UserID,
		crash.Environment, crash.CreatedAt, crash.LogFilePath, string(metadata),
	)
	return err
}

func (r *SQLiteRepository) GetCrash(ctx context.Context, id string) (*core.Crash, error) {
	crash := &core.Crash{}
	var metadata string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, app_id, app_version, platform, os_version, device_model, error_type, error_message, fingerprint, group_id, user_id, environment, created_at, log_file_path, COALESCE(metadata, '{}')
		FROM crashes WHERE id = ?`, id,
	).Scan(&crash.ID, &crash.AppID, &crash.AppVersion, &crash.Platform, &crash.OSVersion,
		&crash.DeviceModel, &crash.ErrorType, &crash.ErrorMessage, &crash.Fingerprint,
		&crash.GroupID, &crash.UserID, &crash.Environment, &crash.CreatedAt, &crash.LogFilePath, &metadata)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(metadata), &crash.Metadata)
	return crash, nil
}

func (r *SQLiteRepository) ListCrashes(ctx context.Context, filter CrashFilter) ([]*core.Crash, int, error) {
	var conditions []string
	var args []interface{}

	if filter.AppID != "" {
		conditions = append(conditions, "app_id = ?")
		args = append(args, filter.AppID)
	}
	if filter.GroupID != "" {
		conditions = append(conditions, "group_id = ?")
		args = append(args, filter.GroupID)
	}
	if filter.Platform != "" {
		conditions = append(conditions, "platform = ?")
		args = append(args, filter.Platform)
	}
	if filter.Environment != "" {
		conditions = append(conditions, "environment = ?")
		args = append(args, filter.Environment)
	}
	if filter.ErrorType != "" {
		conditions = append(conditions, "error_type = ?")
		args = append(args, filter.ErrorType)
	}
	if filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID)
	}
	if filter.FromDate != nil {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, filter.FromDate)
	}
	if filter.ToDate != nil {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, filter.ToDate)
	}
	if filter.Search != "" {
		conditions = append(conditions, "(error_type LIKE ? OR error_message LIKE ?)")
		searchTerm := "%" + filter.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM crashes %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	query := fmt.Sprintf(
		`SELECT id, app_id, app_version, platform, os_version, device_model, error_type, error_message, fingerprint, group_id, user_id, environment, created_at, log_file_path, COALESCE(metadata, '{}')
		FROM crashes %s ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		whereClause,
	)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var crashes []*core.Crash
	for rows.Next() {
		crash := &core.Crash{}
		var metadata string
		if err := rows.Scan(&crash.ID, &crash.AppID, &crash.AppVersion, &crash.Platform, &crash.OSVersion,
			&crash.DeviceModel, &crash.ErrorType, &crash.ErrorMessage, &crash.Fingerprint,
			&crash.GroupID, &crash.UserID, &crash.Environment, &crash.CreatedAt, &crash.LogFilePath, &metadata); err != nil {
			return nil, 0, err
		}
		json.Unmarshal([]byte(metadata), &crash.Metadata)
		crashes = append(crashes, crash)
	}
	return crashes, total, rows.Err()
}

func (r *SQLiteRepository) DeleteCrash(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM crashes WHERE id = ?`, id)
	return err
}

func (r *SQLiteRepository) DeleteCrashesOlderThan(ctx context.Context, appID string, before time.Time) (int, error) {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM crashes WHERE app_id = ? AND created_at < ?`, appID, before,
	)
	if err != nil {
		return 0, err
	}
	count, _ := result.RowsAffected()
	return int(count), nil
}

// Crash group operations
func (r *SQLiteRepository) GetOrCreateGroup(ctx context.Context, crash *core.Crash) (*core.CrashGroup, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback()

	// Try to find existing group
	group := &core.CrashGroup{}
	err = tx.QueryRowContext(ctx,
		`SELECT id, app_id, fingerprint, error_type, error_message, first_seen, last_seen, occurrence_count, status, assigned_to, notes
		FROM crash_groups WHERE app_id = ? AND fingerprint = ?`,
		crash.AppID, crash.Fingerprint,
	).Scan(&group.ID, &group.AppID, &group.Fingerprint, &group.ErrorType, &group.ErrorMessage,
		&group.FirstSeen, &group.LastSeen, &group.OccurrenceCount, &group.Status, &group.AssignedTo, &group.Notes)

	if err == nil {
		// Group exists, update it
		_, err = tx.ExecContext(ctx,
			`UPDATE crash_groups SET last_seen = ?, occurrence_count = occurrence_count + 1 WHERE id = ?`,
			crash.CreatedAt, group.ID,
		)
		if err != nil {
			return nil, false, err
		}
		group.LastSeen = crash.CreatedAt
		group.OccurrenceCount++
		return group, false, tx.Commit()
	}

	if err != sql.ErrNoRows {
		return nil, false, err
	}

	// Create new group
	group = &core.CrashGroup{
		ID:              crash.GroupID,
		AppID:           crash.AppID,
		Fingerprint:     crash.Fingerprint,
		ErrorType:       crash.ErrorType,
		ErrorMessage:    crash.ErrorMessage,
		FirstSeen:       crash.CreatedAt,
		LastSeen:        crash.CreatedAt,
		OccurrenceCount: 1,
		Status:          string(core.GroupStatusOpen),
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO crash_groups (id, app_id, fingerprint, error_type, error_message, first_seen, last_seen, occurrence_count, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		group.ID, group.AppID, group.Fingerprint, group.ErrorType, group.ErrorMessage,
		group.FirstSeen, group.LastSeen, group.OccurrenceCount, group.Status,
	)
	if err != nil {
		return nil, false, err
	}

	return group, true, tx.Commit()
}

func (r *SQLiteRepository) GetGroup(ctx context.Context, id string) (*core.CrashGroup, error) {
	group := &core.CrashGroup{}
	var assignedTo, notes sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, app_id, fingerprint, error_type, error_message, first_seen, last_seen, occurrence_count, status, assigned_to, notes
		FROM crash_groups WHERE id = ?`, id,
	).Scan(&group.ID, &group.AppID, &group.Fingerprint, &group.ErrorType, &group.ErrorMessage,
		&group.FirstSeen, &group.LastSeen, &group.OccurrenceCount, &group.Status, &assignedTo, &notes)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	group.AssignedTo = assignedTo.String
	group.Notes = notes.String
	return group, err
}

func (r *SQLiteRepository) ListGroups(ctx context.Context, filter GroupFilter) ([]*core.CrashGroup, int, error) {
	var conditions []string
	var args []interface{}

	if filter.AppID != "" {
		conditions = append(conditions, "app_id = ?")
		args = append(args, filter.AppID)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.ErrorType != "" {
		conditions = append(conditions, "error_type = ?")
		args = append(args, filter.ErrorType)
	}
	if filter.Search != "" {
		conditions = append(conditions, "(error_type LIKE ? OR error_message LIKE ?)")
		searchTerm := "%" + filter.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	var total int
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM crash_groups %s", whereClause), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Determine sort
	sortBy := "last_seen"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	if filter.Limit == 0 {
		filter.Limit = 50
	}

	query := fmt.Sprintf(
		`SELECT id, app_id, fingerprint, error_type, error_message, first_seen, last_seen, occurrence_count, status, assigned_to, notes
		FROM crash_groups %s ORDER BY %s %s LIMIT ? OFFSET ?`,
		whereClause, sortBy, sortOrder,
	)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var groups []*core.CrashGroup
	for rows.Next() {
		group := &core.CrashGroup{}
		var assignedTo, notes sql.NullString
		if err := rows.Scan(&group.ID, &group.AppID, &group.Fingerprint, &group.ErrorType, &group.ErrorMessage,
			&group.FirstSeen, &group.LastSeen, &group.OccurrenceCount, &group.Status, &assignedTo, &notes); err != nil {
			return nil, 0, err
		}
		group.AssignedTo = assignedTo.String
		group.Notes = notes.String
		groups = append(groups, group)
	}
	return groups, total, rows.Err()
}

func (r *SQLiteRepository) UpdateGroupStatus(ctx context.Context, id string, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE crash_groups SET status = ? WHERE id = ?`, status, id)
	return err
}

func (r *SQLiteRepository) UpdateGroup(ctx context.Context, group *core.CrashGroup) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE crash_groups SET status = ?, assigned_to = ?, notes = ? WHERE id = ?`,
		group.Status, group.AssignedTo, group.Notes, group.ID,
	)
	return err
}

func (r *SQLiteRepository) IncrementGroupCount(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE crash_groups SET occurrence_count = occurrence_count + 1, last_seen = ? WHERE id = ?`,
		time.Now(), id,
	)
	return err
}

// Alert operations
func (r *SQLiteRepository) CreateAlert(ctx context.Context, alert *core.Alert) error {
	config, _ := json.Marshal(alert.Config)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO alerts (id, app_id, type, config, enabled, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		alert.ID, alert.AppID, alert.Type, string(config), alert.Enabled, alert.CreatedAt,
	)
	return err
}

func (r *SQLiteRepository) GetAlert(ctx context.Context, id string) (*core.Alert, error) {
	alert := &core.Alert{}
	var config string
	var enabled int
	err := r.db.QueryRowContext(ctx,
		`SELECT id, app_id, type, config, enabled, created_at FROM alerts WHERE id = ?`, id,
	).Scan(&alert.ID, &alert.AppID, &alert.Type, &config, &enabled, &alert.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	alert.Enabled = enabled == 1
	json.Unmarshal([]byte(config), &alert.Config)
	return alert, err
}

func (r *SQLiteRepository) ListAlerts(ctx context.Context, appID string) ([]*core.Alert, error) {
	query := `SELECT id, app_id, type, config, enabled, created_at FROM alerts`
	var args []interface{}
	if appID != "" {
		query += " WHERE app_id = ?"
		args = append(args, appID)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*core.Alert
	for rows.Next() {
		alert := &core.Alert{}
		var config string
		var enabled int
		if err := rows.Scan(&alert.ID, &alert.AppID, &alert.Type, &config, &enabled, &alert.CreatedAt); err != nil {
			return nil, err
		}
		alert.Enabled = enabled == 1
		json.Unmarshal([]byte(config), &alert.Config)
		alerts = append(alerts, alert)
	}
	return alerts, rows.Err()
}

func (r *SQLiteRepository) UpdateAlert(ctx context.Context, alert *core.Alert) error {
	config, _ := json.Marshal(alert.Config)
	enabled := 0
	if alert.Enabled {
		enabled = 1
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE alerts SET type = ?, config = ?, enabled = ? WHERE id = ?`,
		alert.Type, string(config), enabled, alert.ID,
	)
	return err
}

func (r *SQLiteRepository) DeleteAlert(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM alerts WHERE id = ?`, id)
	return err
}

// Stats
func (r *SQLiteRepository) GetAppStats(ctx context.Context, appID string) (*core.CrashStats, error) {
	stats := &core.CrashStats{AppID: appID}

	// Total crashes
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM crashes WHERE app_id = ?`, appID).Scan(&stats.TotalCrashes)

	// Total groups
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM crash_groups WHERE app_id = ?`, appID).Scan(&stats.TotalGroups)

	// Open groups
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM crash_groups WHERE app_id = ? AND status = 'open'`, appID).Scan(&stats.OpenGroups)

	// Crashes in time periods
	now := time.Now()
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM crashes WHERE app_id = ? AND created_at >= ?`,
		appID, now.Add(-24*time.Hour)).Scan(&stats.CrashesLast24h)
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM crashes WHERE app_id = ? AND created_at >= ?`,
		appID, now.Add(-7*24*time.Hour)).Scan(&stats.CrashesLast7d)
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM crashes WHERE app_id = ? AND created_at >= ?`,
		appID, now.Add(-30*24*time.Hour)).Scan(&stats.CrashesLast30d)

	// Top errors
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, error_type, error_message, occurrence_count FROM crash_groups
		WHERE app_id = ? ORDER BY occurrence_count DESC LIMIT 5`, appID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var summary core.ErrorSummary
			rows.Scan(&summary.GroupID, &summary.ErrorType, &summary.ErrorMessage, &summary.Count)
			stats.TopErrors = append(stats.TopErrors, summary)
		}
	}

	// Crash trend (last 30 days)
	rows, err = r.db.QueryContext(ctx,
		`SELECT DATE(created_at) as date, COUNT(*) as count FROM crashes
		WHERE app_id = ? AND created_at >= ? GROUP BY DATE(created_at) ORDER BY date`,
		appID, now.Add(-30*24*time.Hour))
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var point core.TrendPoint
			rows.Scan(&point.Date, &point.Count)
			stats.CrashTrend = append(stats.CrashTrend, point)
		}
	}

	return stats, nil
}

// Settings operations
func (r *SQLiteRepository) GetSetting(ctx context.Context, key string) (string, error) {
	var value string
	err := r.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (r *SQLiteRepository) SetSetting(ctx context.Context, key, value string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = ?`,
		key, value, value,
	)
	return err
}
