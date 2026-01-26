package rest

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/flakerimi/inceptor/internal/core"
	"github.com/flakerimi/inceptor/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler holds dependencies for REST handlers
type Handler struct {
	repo      storage.Repository
	fileStore storage.FileStore
	grouper   *core.Grouper
	alerter   *core.AlertManager
}

// NewHandler creates a new Handler
func NewHandler(repo storage.Repository, fileStore storage.FileStore, alerter *core.AlertManager) *Handler {
	return &Handler{
		repo:      repo,
		fileStore: fileStore,
		grouper:   core.NewGrouper(),
		alerter:   alerter,
	}
}

// Health check
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().UTC()})
}

// SubmitCrash handles crash report submission
func (h *Handler) SubmitCrash(c *gin.Context) {
	app := GetApp(c)
	if app == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid app context"})
		return
	}

	var submission core.CrashSubmission
	if err := c.ShouldBindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Create crash object
	crash := &core.Crash{
		ID:           uuid.New().String(),
		AppID:        app.ID,
		AppVersion:   submission.AppVersion,
		Platform:     submission.Platform,
		OSVersion:    submission.OSVersion,
		DeviceModel:  submission.DeviceModel,
		ErrorType:    submission.ErrorType,
		ErrorMessage: submission.ErrorMessage,
		StackTrace:   submission.StackTrace,
		UserID:       submission.UserID,
		Environment:  submission.Environment,
		CreatedAt:    time.Now().UTC(),
		Metadata:     submission.Metadata,
		Breadcrumbs:  submission.Breadcrumbs,
	}

	// Set default environment if not provided
	if crash.Environment == "" {
		crash.Environment = core.EnvironmentProduction
	}

	// Generate fingerprint
	crash.Fingerprint = h.grouper.GenerateFingerprint(crash)

	// Get or create group
	crash.GroupID = uuid.New().String() // Pre-generate in case new group needed
	group, isNewGroup, err := h.repo.GetOrCreateGroup(c.Request.Context(), crash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process crash group"})
		return
	}
	crash.GroupID = group.ID

	// Save full crash log to file
	logPath, err := h.fileStore.SaveCrashLog(c.Request.Context(), crash)
	if err != nil {
		// Log error but continue - file storage is secondary
		// log.Error().Err(err).Msg("Failed to save crash log file")
	} else {
		crash.LogFilePath = logPath
	}

	// Save crash to database
	if err := h.repo.CreateCrash(c.Request.Context(), crash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save crash"})
		return
	}

	// Send alert
	if h.alerter != nil {
		eventType := core.AlertEventNewCrash
		if isNewGroup {
			eventType = core.AlertEventNewGroup
		}
		h.alerter.Notify(core.AlertEvent{
			Type:       eventType,
			AppID:      app.ID,
			Crash:      crash,
			Group:      group,
			IsNewGroup: isNewGroup,
		})
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":           crash.ID,
		"group_id":     crash.GroupID,
		"fingerprint":  crash.Fingerprint,
		"is_new_group": isNewGroup,
	})
}

// GetCrash retrieves a single crash
func (h *Handler) GetCrash(c *gin.Context) {
	id := c.Param("id")

	crash, err := h.repo.GetCrash(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve crash"})
		return
	}

	if crash == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Crash not found"})
		return
	}

	// Check access
	app := GetApp(c)
	if app != nil && crash.AppID != app.ID && !IsAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Load full crash data from file if available
	if crash.LogFilePath != "" {
		if fullCrash, err := h.fileStore.GetCrashLog(c.Request.Context(), crash.LogFilePath); err == nil && fullCrash != nil {
			crash = fullCrash
		}
	}

	c.JSON(http.StatusOK, crash)
}

// ListCrashes lists crashes with filters
func (h *Handler) ListCrashes(c *gin.Context) {
	filter := storage.CrashFilter{
		AppID:       c.Query("app_id"),
		GroupID:     c.Query("group_id"),
		Platform:    c.Query("platform"),
		Environment: c.Query("environment"),
		ErrorType:   c.Query("error_type"),
		UserID:      c.Query("user_id"),
		Search:      c.Query("search"),
		Limit:       parseIntQuery(c, "limit", 50),
		Offset:      parseIntQuery(c, "offset", 0),
	}

	// Non-admin users can only see their own app's crashes
	app := GetApp(c)
	if app != nil {
		filter.AppID = app.ID
	}

	// Parse date filters
	if from := c.Query("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			filter.FromDate = &t
		}
	}
	if to := c.Query("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			filter.ToDate = &t
		}
	}

	crashes, total, err := h.repo.ListCrashes(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list crashes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   crashes,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// DeleteCrash deletes a crash
func (h *Handler) DeleteCrash(c *gin.Context) {
	id := c.Param("id")

	// Get crash first to verify ownership and get file path
	crash, err := h.repo.GetCrash(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve crash"})
		return
	}

	if crash == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Crash not found"})
		return
	}

	// Check access
	app := GetApp(c)
	if app != nil && crash.AppID != app.ID && !IsAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Delete from database
	if err := h.repo.DeleteCrash(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete crash"})
		return
	}

	// Delete file if exists
	if crash.LogFilePath != "" {
		h.fileStore.DeleteCrashLog(c.Request.Context(), crash.LogFilePath)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Crash deleted"})
}

// GetGroup retrieves a crash group
func (h *Handler) GetGroup(c *gin.Context) {
	id := c.Param("id")

	group, err := h.repo.GetGroup(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve group"})
		return
	}

	if group == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Check access
	app := GetApp(c)
	if app != nil && group.AppID != app.ID && !IsAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, group)
}

// ListGroups lists crash groups with filters
func (h *Handler) ListGroups(c *gin.Context) {
	filter := storage.GroupFilter{
		AppID:     c.Query("app_id"),
		Status:    c.Query("status"),
		ErrorType: c.Query("error_type"),
		Search:    c.Query("search"),
		SortBy:    c.DefaultQuery("sort_by", "last_seen"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
		Limit:     parseIntQuery(c, "limit", 50),
		Offset:    parseIntQuery(c, "offset", 0),
	}

	// Non-admin users can only see their own app's groups
	app := GetApp(c)
	if app != nil {
		filter.AppID = app.ID
	}

	groups, total, err := h.repo.ListGroups(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list groups"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   groups,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// UpdateGroup updates a crash group
func (h *Handler) UpdateGroup(c *gin.Context) {
	id := c.Param("id")

	group, err := h.repo.GetGroup(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve group"})
		return
	}

	if group == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Check access
	app := GetApp(c)
	if app != nil && group.AppID != app.ID && !IsAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var update struct {
		Status     *string `json:"status"`
		AssignedTo *string `json:"assigned_to"`
		Notes      *string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if update.Status != nil {
		group.Status = *update.Status
	}
	if update.AssignedTo != nil {
		group.AssignedTo = *update.AssignedTo
	}
	if update.Notes != nil {
		group.Notes = *update.Notes
	}

	if err := h.repo.UpdateGroup(c.Request.Context(), group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group"})
		return
	}

	c.JSON(http.StatusOK, group)
}

// CreateApp creates a new app
func (h *Handler) CreateApp(c *gin.Context) {
	var req struct {
		Name          string `json:"name" binding:"required"`
		RetentionDays int    `json:"retention_days"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Generate API key
	apiKey := generateSecureAPIKey()

	app := &core.App{
		ID:            uuid.New().String(),
		Name:          req.Name,
		APIKey:        apiKey, // Return to user only once
		APIKeyHash:    HashAPIKey(apiKey),
		CreatedAt:     time.Now().UTC(),
		RetentionDays: req.RetentionDays,
	}

	if app.RetentionDays <= 0 {
		app.RetentionDays = 30
	}

	if err := h.repo.CreateApp(c.Request.Context(), app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create app"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":             app.ID,
		"name":           app.Name,
		"api_key":        apiKey, // Only returned on creation
		"created_at":     app.CreatedAt,
		"retention_days": app.RetentionDays,
	})
}

// GetApp retrieves app info
func (h *Handler) GetApp(c *gin.Context) {
	id := c.Param("id")

	app, err := h.repo.GetApp(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve app"})
		return
	}

	if app == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "App not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             app.ID,
		"name":           app.Name,
		"created_at":     app.CreatedAt,
		"retention_days": app.RetentionDays,
	})
}

// ListApps lists all apps (admin only)
func (h *Handler) ListApps(c *gin.Context) {
	apps, err := h.repo.ListApps(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list apps"})
		return
	}

	// Don't expose API key hashes
	result := make([]gin.H, len(apps))
	for i, app := range apps {
		result[i] = gin.H{
			"id":             app.ID,
			"name":           app.Name,
			"created_at":     app.CreatedAt,
			"retention_days": app.RetentionDays,
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GetAppStats gets statistics for an app
func (h *Handler) GetAppStats(c *gin.Context) {
	id := c.Param("id")

	// Check access
	app := GetApp(c)
	if app != nil && app.ID != id && !IsAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	stats, err := h.repo.GetAppStats(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CreateAlert creates a new alert
func (h *Handler) CreateAlert(c *gin.Context) {
	var req struct {
		AppID   string                 `json:"app_id" binding:"required"`
		Type    string                 `json:"type" binding:"required"`
		Config  map[string]interface{} `json:"config"`
		Enabled bool                   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	alert := &core.Alert{
		ID:        uuid.New().String(),
		AppID:     req.AppID,
		Type:      req.Type,
		Config:    req.Config,
		Enabled:   req.Enabled,
		CreatedAt: time.Now().UTC(),
	}

	if err := h.repo.CreateAlert(c.Request.Context(), alert); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create alert"})
		return
	}

	// Update alerter
	if h.alerter != nil {
		h.alerter.AddAlert(alert)
	}

	c.JSON(http.StatusCreated, alert)
}

// ListAlerts lists alerts
func (h *Handler) ListAlerts(c *gin.Context) {
	appID := c.Query("app_id")

	// Non-admin users can only see their own app's alerts
	app := GetApp(c)
	if app != nil {
		appID = app.ID
	}

	alerts, err := h.repo.ListAlerts(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": alerts})
}

// DeleteAlert deletes an alert
func (h *Handler) DeleteAlert(c *gin.Context) {
	id := c.Param("id")

	if err := h.repo.DeleteAlert(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert deleted"})
}

// Helper functions
func parseIntQuery(c *gin.Context, key string, defaultVal int) int {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	if i, err := strconv.Atoi(val); err == nil {
		return i
	}
	return defaultVal
}

func generateSecureAPIKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "ink_" + hex.EncodeToString(b)[:32]
}
