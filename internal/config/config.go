package config

import (
	"os"
	"strconv"
	"time"
)

const (
	defaultPort          = ":8080"
	defaultDBPath        = "meshdns.db"
	defaultProbeInterval = 60 * time.Second
	defaultProbeTimeout  = 5 * time.Second
	defaultWorkers       = 8
)

type Config struct {
	Port          string
	DBPath        string
	ProbeInterval time.Duration
	ProbeTimeout  time.Duration
	Workers       int
}

func Load() Config {
	port := envString("PORT", defaultPort)
	// Render sets PORT without colon; MESHDNS_PORT with colon takes precedence
	if mp := os.Getenv("MESHDNS_PORT"); mp != "" {
		port = mp
	} else if port != "" && port[0] != ':' {
		port = ":" + port
	}
	return Config{
		Port:          port,
		DBPath:        envString("MESHDNS_DB", defaultDBPath),
		ProbeInterval: envDuration("MESHDNS_PROBE_INTERVAL", defaultProbeInterval),
		ProbeTimeout:  envDuration("MESHDNS_PROBE_TIMEOUT", defaultProbeTimeout),
		Workers:       envInt("MESHDNS_WORKERS", defaultWorkers),
	}
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
