package config

import (
	"testing"
	"time"
)

func TestTenhouPollIntervalIsClamped(t *testing.T) {
	t.Setenv("TENHOU_POLL_INTERVAL", "1m")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.TenhouPollInterval != TenhouMinPollInterval {
		t.Fatalf("got %v, want %v", c.TenhouPollInterval, TenhouMinPollInterval)
	}
	t.Setenv("TENHOU_POLL_INTERVAL", "30m")
	c, _ = Load()
	if c.TenhouPollInterval != 30*time.Minute {
		t.Fatalf("got %v", c.TenhouPollInterval)
	}
}
