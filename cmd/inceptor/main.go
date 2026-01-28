package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flakerimi/inceptor/internal/api/rest"
	"github.com/flakerimi/inceptor/internal/auth"
	"github.com/flakerimi/inceptor/internal/config"
	"github.com/flakerimi/inceptor/internal/core"
	"github.com/flakerimi/inceptor/internal/storage"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Version is updated by scripts/release.sh
var version = "1.0.1"

func main() {
	// Parse flags
	configPath := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	log.Info().Msg("Starting Inceptor - Crash Logging Service")

	// Initialize storage
	repo, err := storage.NewSQLiteRepository(cfg.Storage.SQLitePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer repo.Close()

	fileStore, err := storage.NewLocalFileStore(cfg.Storage.LogsPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize file store")
	}

	// Initialize alert manager
	alerter := core.NewAlertManager(
		core.SMTPConfig{
			Host:     cfg.Alerts.SMTP.Host,
			Port:     cfg.Alerts.SMTP.Port,
			Username: cfg.Alerts.SMTP.Username,
			Password: cfg.Alerts.SMTP.Password,
			From:     cfg.Alerts.SMTP.From,
		},
		cfg.Alerts.Slack.WebhookURL,
	)
	defer alerter.Close()

	// Load existing alerts
	alerts, err := repo.ListAlerts(context.Background(), "")
	if err == nil {
		alerter.SetAlerts(alerts)
	}

	// Initialize retention manager
	retention := core.NewRetentionManager(
		repo,
		fileStore,
		cfg.Retention.DefaultDays,
		cfg.Retention.CleanupInterval,
	)
	retention.Start()
	defer retention.Stop()

	// Initialize auth manager
	passwordHash, _ := repo.GetSetting(context.Background(), "password_hash")
	authManager := auth.NewManager(passwordHash, func(hash string) {
		if err := repo.SetSetting(context.Background(), "password_hash", hash); err != nil {
			log.Error().Err(err).Msg("Failed to save password hash")
		}
	})

	// Initialize REST server
	restServer := rest.NewServer(repo, fileStore, alerter, authManager, cfg.Auth.AdminKey, version)

	// Start servers
	errChan := make(chan error, 2)

	// REST server
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.RESTPort)
		log.Info().Str("addr", addr).Msg("Starting REST API server")
		if err := restServer.Run(addr); err != nil {
			errChan <- fmt.Errorf("REST server error: %w", err)
		}
	}()

	// gRPC server (optional - uncomment when proto is compiled)
	/*
	go func() {
		grpcServer := grpc.NewServer(repo, fileStore, alerter, cfg.Auth.AdminKey)
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.GRPCPort)
		log.Info().Str("addr", addr).Msg("Starting gRPC server")
		if err := grpcServer.Run(addr); err != nil {
			errChan <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()
	*/

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Fatal().Err(err).Msg("Server error")
	case sig := <-sigChan:
		log.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
	}

	log.Info().Msg("Shutting down gracefully...")
}
