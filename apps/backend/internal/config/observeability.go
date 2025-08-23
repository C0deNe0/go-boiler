package config

import (
	"fmt"
	"time"
)

type ObserveabilityConfig struct {
	ServiceName  string            `koanf:"service_name" validate:"required"`
	Environment  string            `koanf:"environment" validate:"required"`
	Logging      LoggingConfig     `koanf:"logging" validate:"required"`
	NewRelic     NewRelicConfig    `koanf:"new_relic" validate:"required"`
	HealthChecks HealthCheckConfig `koanf:"health_checks" validate:"required"`
}

type LoggingConfig struct {
	Level             string        `koanf:"level" validate:"required"`
	Format            string        `koanf:"format" validate:"required"`
	SlowQueryTreshold time.Duration `koanf:"slow_query_treshold"`
}

type NewRelicConfig struct {
	LicenseKey                string `koanf:"license_key" validate:"required"`
	AppLogForwardEnabled      bool   `koanf:"app_log_forward_enabled"`
	DistributedTracingEnabled bool   `koanf:"distributed_tracing_enabled"`
	DebugLogging              bool   `koanf:"debug_logging"`
}

type HealthCheckConfig struct {
	Enabled  bool          `koanf:"enabled"`
	Interval time.Duration `koanf:"interval" validate:"min=1s"`
	Timeout  time.Duration `koanf:"timeout" validate:"min=1s"`
	Checks   []string      `koanf:"checks"`
}

func DefaultObserveabilityConfig() *ObserveabilityConfig {
	return &ObserveabilityConfig{
		ServiceName: "boilerplate",
		Environment: "development",
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		NewRelic: NewRelicConfig{
			LicenseKey:                "",
			AppLogForwardEnabled:      true,
			DistributedTracingEnabled: true,
			DebugLogging:              true, //this is false by default, to avoid mixed log formats
		},
		HealthChecks: HealthCheckConfig{
			Enabled:  true,
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
			Checks:   []string{"database", "redis"},
		},
	}
}

func (c *ObserveabilityConfig) validate() error {
	if c.ServiceName == "" {
		return fmt.Errorf("service_name is required")
	}
	if c.Environment == "" {
		return fmt.Errorf("environment is required")
	}

	//validte the logging config

	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid logging level: %s(must be one of debug, info, warn, error)", c.Logging.Level)

	}

	//validate slow query treshold
	if c.Logging.SlowQueryTreshold < 0 {
		return fmt.Errorf("slow_query_treshold must be a positive duration")
	}

	return nil

}

func (c *ObserveabilityConfig) GetLogLevel() string {
	switch c.Environment {
	case "production":
		if c.Logging.Level == "" {
			return "info" //in production, we don't want debug logs
		}
	case "development":
		if c.Logging.Level == "" {
			return "debug"
		}
	}

	return c.Logging.Level
}

func (c *ObserveabilityConfig) IsProduction() bool {
	return c.Environment == "production"
}
