package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// AlertManager handles sending alerts when crashes occur
type AlertManager struct {
	alerts    []*Alert
	alertsMu  sync.RWMutex
	smtpCfg   SMTPConfig
	slackURL  string
	client    *http.Client
	queue     chan AlertEvent
	ctx       context.Context
	cancel    context.CancelFunc
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// AlertEvent represents an event that may trigger alerts
type AlertEvent struct {
	Type      AlertEventType
	AppID     string
	Crash     *Crash
	Group     *CrashGroup
	IsNewGroup bool
}

// AlertEventType defines types of alertable events
type AlertEventType string

const (
	AlertEventNewCrash    AlertEventType = "new_crash"
	AlertEventNewGroup    AlertEventType = "new_group"
	AlertEventThreshold   AlertEventType = "threshold"
)

// NewAlertManager creates a new AlertManager
func NewAlertManager(smtpCfg SMTPConfig, slackURL string) *AlertManager {
	ctx, cancel := context.WithCancel(context.Background())

	am := &AlertManager{
		alerts:   make([]*Alert, 0),
		smtpCfg:  smtpCfg,
		slackURL: slackURL,
		client:   &http.Client{Timeout: 10 * time.Second},
		queue:    make(chan AlertEvent, 100),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Start worker
	go am.worker()

	return am
}

// SetAlerts updates the list of configured alerts
func (am *AlertManager) SetAlerts(alerts []*Alert) {
	am.alertsMu.Lock()
	defer am.alertsMu.Unlock()
	am.alerts = alerts
}

// AddAlert adds a single alert configuration
func (am *AlertManager) AddAlert(alert *Alert) {
	am.alertsMu.Lock()
	defer am.alertsMu.Unlock()
	am.alerts = append(am.alerts, alert)
}

// Notify queues an alert event for processing
func (am *AlertManager) Notify(event AlertEvent) {
	select {
	case am.queue <- event:
	default:
		log.Warn().Msg("Alert queue full, dropping event")
	}
}

// Close shuts down the alert manager
func (am *AlertManager) Close() {
	am.cancel()
	close(am.queue)
}

// worker processes alert events
func (am *AlertManager) worker() {
	for {
		select {
		case <-am.ctx.Done():
			return
		case event, ok := <-am.queue:
			if !ok {
				return
			}
			am.processEvent(event)
		}
	}
}

// processEvent processes a single alert event
func (am *AlertManager) processEvent(event AlertEvent) {
	am.alertsMu.RLock()
	alerts := make([]*Alert, len(am.alerts))
	copy(alerts, am.alerts)
	am.alertsMu.RUnlock()

	for _, alert := range alerts {
		if !alert.Enabled {
			continue
		}

		if alert.AppID != "" && alert.AppID != event.AppID {
			continue
		}

		// Check if this alert type matches the event
		if !am.shouldAlert(alert, event) {
			continue
		}

		// Send the alert
		if err := am.sendAlert(alert, event); err != nil {
			log.Error().Err(err).Str("alert_id", alert.ID).Msg("Failed to send alert")
		}
	}
}

// shouldAlert checks if an alert should be triggered for an event
func (am *AlertManager) shouldAlert(alert *Alert, event AlertEvent) bool {
	// Get alert conditions from config
	conditions, _ := alert.Config["conditions"].(map[string]interface{})

	switch event.Type {
	case AlertEventNewGroup:
		// Alert on new crash groups
		if alertOnNew, ok := conditions["on_new_group"].(bool); ok && alertOnNew {
			return true
		}
	case AlertEventNewCrash:
		// Alert on every crash (usually not recommended)
		if alertOnCrash, ok := conditions["on_every_crash"].(bool); ok && alertOnCrash {
			return true
		}
	case AlertEventThreshold:
		// Alert when threshold exceeded (handled elsewhere)
		return true
	}

	// Check error type filter
	if errorTypes, ok := conditions["error_types"].([]interface{}); ok && len(errorTypes) > 0 {
		for _, et := range errorTypes {
			if etStr, ok := et.(string); ok && event.Crash != nil && event.Crash.ErrorType == etStr {
				return true
			}
		}
	}

	return false
}

// sendAlert sends an alert via the configured channel
func (am *AlertManager) sendAlert(alert *Alert, event AlertEvent) error {
	switch alert.Type {
	case "webhook":
		return am.sendWebhook(alert, event)
	case "email":
		return am.sendEmail(alert, event)
	case "slack":
		return am.sendSlack(alert, event)
	default:
		return fmt.Errorf("unknown alert type: %s", alert.Type)
	}
}

// sendWebhook sends a webhook notification
func (am *AlertManager) sendWebhook(alert *Alert, event AlertEvent) error {
	url, ok := alert.Config["url"].(string)
	if !ok || url == "" {
		return fmt.Errorf("webhook URL not configured")
	}

	payload := map[string]interface{}{
		"event_type": event.Type,
		"app_id":     event.AppID,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	if event.Crash != nil {
		payload["crash"] = map[string]interface{}{
			"id":            event.Crash.ID,
			"error_type":    event.Crash.ErrorType,
			"error_message": event.Crash.ErrorMessage,
			"platform":      event.Crash.Platform,
			"app_version":   event.Crash.AppVersion,
			"environment":   event.Crash.Environment,
		}
	}

	if event.Group != nil {
		payload["group"] = map[string]interface{}{
			"id":               event.Group.ID,
			"fingerprint":      event.Group.Fingerprint,
			"occurrence_count": event.Group.OccurrenceCount,
			"first_seen":       event.Group.FirstSeen,
			"last_seen":        event.Group.LastSeen,
		}
	}

	payload["is_new_group"] = event.IsNewGroup

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add custom headers if configured
	if headers, ok := alert.Config["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if vStr, ok := v.(string); ok {
				req.Header.Set(k, vStr)
			}
		}
	}

	resp, err := am.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// sendEmail sends an email notification
func (am *AlertManager) sendEmail(alert *Alert, event AlertEvent) error {
	to, ok := alert.Config["to"].(string)
	if !ok || to == "" {
		return fmt.Errorf("email recipient not configured")
	}

	if am.smtpCfg.Host == "" {
		return fmt.Errorf("SMTP not configured")
	}

	subject := fmt.Sprintf("[Inceptor] New crash in %s", event.AppID)
	if event.IsNewGroup {
		subject = fmt.Sprintf("[Inceptor] NEW ERROR in %s: %s", event.AppID, event.Crash.ErrorType)
	}

	body := fmt.Sprintf(`
New crash detected in your application.

App ID: %s
Error Type: %s
Error Message: %s
Platform: %s
App Version: %s
Environment: %s
Time: %s

Group ID: %s
Is New Group: %v
Occurrence Count: %d

View in dashboard: [your-dashboard-url]/crashes/%s
`,
		event.AppID,
		event.Crash.ErrorType,
		event.Crash.ErrorMessage,
		event.Crash.Platform,
		event.Crash.AppVersion,
		event.Crash.Environment,
		event.Crash.CreatedAt.Format(time.RFC3339),
		event.Group.ID,
		event.IsNewGroup,
		event.Group.OccurrenceCount,
		event.Crash.ID,
	)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		am.smtpCfg.From, to, subject, body)

	addr := fmt.Sprintf("%s:%d", am.smtpCfg.Host, am.smtpCfg.Port)

	var auth smtp.Auth
	if am.smtpCfg.Username != "" {
		auth = smtp.PlainAuth("", am.smtpCfg.Username, am.smtpCfg.Password, am.smtpCfg.Host)
	}

	return smtp.SendMail(addr, auth, am.smtpCfg.From, []string{to}, []byte(msg))
}

// sendSlack sends a Slack notification
func (am *AlertManager) sendSlack(alert *Alert, event AlertEvent) error {
	webhookURL := am.slackURL
	if url, ok := alert.Config["webhook_url"].(string); ok && url != "" {
		webhookURL = url
	}

	if webhookURL == "" {
		return fmt.Errorf("Slack webhook URL not configured")
	}

	color := "#ff0000" // Red for errors
	if event.IsNewGroup {
		color = "#ff6600" // Orange for new groups
	}

	title := fmt.Sprintf("Crash in %s", event.AppID)
	if event.IsNewGroup {
		title = fmt.Sprintf("🆕 NEW ERROR in %s", event.AppID)
	}

	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color":  color,
				"title":  title,
				"fields": []map[string]interface{}{
					{"title": "Error Type", "value": event.Crash.ErrorType, "short": true},
					{"title": "Platform", "value": event.Crash.Platform, "short": true},
					{"title": "App Version", "value": event.Crash.AppVersion, "short": true},
					{"title": "Environment", "value": event.Crash.Environment, "short": true},
					{"title": "Occurrences", "value": fmt.Sprintf("%d", event.Group.OccurrenceCount), "short": true},
				},
				"text":      event.Crash.ErrorMessage,
				"footer":    "Inceptor Crash Logger",
				"ts":        event.Crash.CreatedAt.Unix(),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := am.client.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to send Slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}
