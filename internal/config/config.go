package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Storage   StorageConfig   `mapstructure:"storage"`
	Retention RetentionConfig `mapstructure:"retention"`
	Alerts    AlertsConfig    `mapstructure:"alerts"`
	Auth      AuthConfig      `mapstructure:"auth"`
}

type ServerConfig struct {
	RESTPort      int    `mapstructure:"rest_port"`
	GRPCPort      int    `mapstructure:"grpc_port"`
	DashboardPort int    `mapstructure:"dashboard_port"`
	Host          string `mapstructure:"host"`
}

type StorageConfig struct {
	SQLitePath string `mapstructure:"sqlite_path"`
	LogsPath   string `mapstructure:"logs_path"`
}

type RetentionConfig struct {
	DefaultDays     int           `mapstructure:"default_days"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

type AlertsConfig struct {
	SMTP  SMTPConfig  `mapstructure:"smtp"`
	Slack SlackConfig `mapstructure:"slack"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

type SlackConfig struct {
	WebhookURL string `mapstructure:"webhook_url"`
}

type AuthConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	AdminKey string `mapstructure:"admin_key"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("server.rest_port", 8080)
	v.SetDefault("server.grpc_port", 9090)
	v.SetDefault("server.dashboard_port", 3000)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("storage.sqlite_path", "./data/inceptor.db")
	v.SetDefault("storage.logs_path", "./data/crashes")
	v.SetDefault("retention.default_days", 30)
	v.SetDefault("retention.cleanup_interval", "24h")
	v.SetDefault("auth.enabled", true)

	// Config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	// Environment variables
	v.SetEnvPrefix("INCEPTOR")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file (optional)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
