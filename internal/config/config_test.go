package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	cfg := Load()

	if cfg.Port != ":8080" {
		t.Fatalf("Port = %q, want %q", cfg.Port, ":8080")
	}
	if cfg.DBPath != "meshdns.db" {
		t.Fatalf("DBPath = %q, want %q", cfg.DBPath, "meshdns.db")
	}
	if cfg.ProbeInterval != 60*time.Second {
		t.Fatalf("ProbeInterval = %s, want %s", cfg.ProbeInterval, 60*time.Second)
	}
	if cfg.ProbeTimeout != 5*time.Second {
		t.Fatalf("ProbeTimeout = %s, want %s", cfg.ProbeTimeout, 5*time.Second)
	}
	if cfg.Workers != 8 {
		t.Fatalf("Workers = %d, want %d", cfg.Workers, 8)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("MESHDNS_PORT", ":9090")
	t.Setenv("MESHDNS_DB", "test.db")
	t.Setenv("MESHDNS_PROBE_INTERVAL", "30s")
	t.Setenv("MESHDNS_PROBE_TIMEOUT", "750ms")
	t.Setenv("MESHDNS_WORKERS", "16")

	cfg := Load()

	if cfg.Port != ":9090" {
		t.Fatalf("Port = %q, want %q", cfg.Port, ":9090")
	}
	if cfg.DBPath != "test.db" {
		t.Fatalf("DBPath = %q, want %q", cfg.DBPath, "test.db")
	}
	if cfg.ProbeInterval != 30*time.Second {
		t.Fatalf("ProbeInterval = %s, want %s", cfg.ProbeInterval, 30*time.Second)
	}
	if cfg.ProbeTimeout != 750*time.Millisecond {
		t.Fatalf("ProbeTimeout = %s, want %s", cfg.ProbeTimeout, 750*time.Millisecond)
	}
	if cfg.Workers != 16 {
		t.Fatalf("Workers = %d, want %d", cfg.Workers, 16)
	}
}

func TestLoadInvalidEnvFallsBackToDefaults(t *testing.T) {
	t.Setenv("MESHDNS_PROBE_INTERVAL", "invalid")
	t.Setenv("MESHDNS_PROBE_TIMEOUT", "soon")
	t.Setenv("MESHDNS_WORKERS", "many")

	cfg := Load()

	if cfg.ProbeInterval != 60*time.Second {
		t.Fatalf("ProbeInterval = %s, want %s", cfg.ProbeInterval, 60*time.Second)
	}
	if cfg.ProbeTimeout != 5*time.Second {
		t.Fatalf("ProbeTimeout = %s, want %s", cfg.ProbeTimeout, 5*time.Second)
	}
	if cfg.Workers != 8 {
		t.Fatalf("Workers = %d, want %d", cfg.Workers, 8)
	}
}
