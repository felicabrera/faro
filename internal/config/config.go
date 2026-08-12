// Package config loads the log service's configuration from the environment.
//
// The signing key has no default and is never generated at startup. A log that
// invents a key when it cannot find one would come up healthy and produce
// checkpoints that nobody can verify against the published key, which is a worse
// failure than not starting.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the faro-log runtime configuration.
type Config struct {
	// Addr is the listen address, e.g. ":2025".
	Addr string
	// StorageDir is the root of the POSIX-backed log.
	StorageDir string
	// SigningKey is the private note key for checkpoints. Supplied through the
	// environment rather than a flag so it does not appear in the process list.
	SigningKey string
	// CheckpointInterval is how often a tree head is published.
	CheckpointInterval time.Duration
	// BatchMaxAge bounds how long an entry waits to be sequenced.
	BatchMaxAge time.Duration
	// BatchMaxSize bounds how many entries are sequenced together.
	BatchMaxSize uint
	// ReadHeaderTimeout bounds how long a client may take to send its headers.
	ReadHeaderTimeout time.Duration
	// ShutdownTimeout bounds graceful shutdown.
	ShutdownTimeout time.Duration
	// CORSOrigin is the single origin allowed to read the log from a browser,
	// i.e. the audit explorer. Empty disables CORS entirely.
	CORSOrigin string
	// LogLevel is one of debug, info, warn, error.
	LogLevel string
}

// Load reads the configuration from the environment and applies defaults.
func Load() (*Config, error) {
	c := &Config{
		Addr:               env("FARO_ADDR", ":2025"),
		StorageDir:         env("FARO_STORAGE_DIR", "./data/log"),
		SigningKey:         os.Getenv("FARO_SIGNING_KEY"),
		CheckpointInterval: envDuration("FARO_CHECKPOINT_INTERVAL", 10*time.Second),
		BatchMaxAge:        envDuration("FARO_BATCH_MAX_AGE", time.Second),
		BatchMaxSize:       envUint("FARO_BATCH_MAX_SIZE", 256),
		ReadHeaderTimeout:  envDuration("FARO_READ_HEADER_TIMEOUT", 10*time.Second),
		ShutdownTimeout:    envDuration("FARO_SHUTDOWN_TIMEOUT", 15*time.Second),
		CORSOrigin:         os.Getenv("FARO_CORS_ORIGIN"),
		LogLevel:           env("FARO_LOG_LEVEL", "info"),
	}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) validate() error {
	if c.SigningKey == "" {
		return fmt.Errorf("config: FARO_SIGNING_KEY is required; generate one with 'faro-verify keygen'")
	}
	if c.StorageDir == "" {
		return fmt.Errorf("config: FARO_STORAGE_DIR must not be empty")
	}
	if c.BatchMaxSize == 0 {
		return fmt.Errorf("config: FARO_BATCH_MAX_SIZE must be greater than zero")
	}
	switch strings.ToLower(c.LogLevel) {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("config: FARO_LOG_LEVEL must be one of debug, info, warn, error; got %q", c.LogLevel)
	}
	return nil
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func envUint(key string, def uint) uint {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return def
	}
	return uint(n)
}
