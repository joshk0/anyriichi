// Package config loads worker configuration from environment variables.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"
)

// TenhouMinPollInterval is the hard minimum imposed by Tenhou's raw-data usage rules.
// See docs/design/worker/tenhou-ingest.md#compliance.
const TenhouMinPollInterval = 20 * time.Minute

type Config struct {
	// DatabaseURL is the Postgres connection string (Supabase pooler, session mode). Optional until M1.
	DatabaseURL string
	// HTTPAddr serves /healthz.
	HTTPAddr string
	LogLevel slog.Level
	// TenhouPollInterval is clamped to TenhouMinPollInterval.
	TenhouPollInterval time.Duration
	// UserAgent sent to external sites. Generic on purpose: no personal contact details.
	UserAgent string
	// TenhouBackfill limits ingestion to hourly result files from this far back (index covers 7 days).
	TenhouBackfill time.Duration
	// TenhouFetchHououLogs queues full logs for every 鳳凰 game listed in scc files.
	TenhouFetchHououLogs bool
}

func Load() (Config, error) {
	c := Config{
		DatabaseURL:          os.Getenv("DATABASE_URL"),
		HTTPAddr:             envOr("HTTP_ADDR", ":8080"),
		TenhouPollInterval:   TenhouMinPollInterval,
		UserAgent:            envOr("USER_AGENT", "anyRiichi-worker/0.1"),
		TenhouBackfill:       48 * time.Hour,
		TenhouFetchHououLogs: os.Getenv("TENHOU_FETCH_HOUOU_LOGS") != "false",
	}
	if v := os.Getenv("TENHOU_BACKFILL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return c, fmt.Errorf("TENHOU_BACKFILL: %w", err)
		}
		c.TenhouBackfill = d
	}
	if err := c.LogLevel.UnmarshalText([]byte(envOr("LOG_LEVEL", "info"))); err != nil {
		return c, fmt.Errorf("LOG_LEVEL: %w", err)
	}
	if v := os.Getenv("TENHOU_POLL_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return c, fmt.Errorf("TENHOU_POLL_INTERVAL: %w", err)
		}
		c.TenhouPollInterval = d
	}
	if c.TenhouPollInterval < TenhouMinPollInterval {
		c.TenhouPollInterval = TenhouMinPollInterval
	}
	return c, nil
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
