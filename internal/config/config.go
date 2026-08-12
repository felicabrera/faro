// Package config loads the log service's configuration from the environment.
//
// The signing key has no default and is never generated at startup. A log that
// invents a key when it cannot find one would come up healthy and produce
// checkpoints that nobody can verify against the published key, which is a worse
// failure than not starting.
package config

import (
	"errors"
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
	// ReadTimeout bounds the whole request, headers and body. ReadHeaderTimeout
	// alone leaves a client free to dribble a body indefinitely.
	ReadTimeout time.Duration
	// WriteTimeout bounds how long a response may take to write.
	WriteTimeout time.Duration
	// IdleTimeout bounds how long a keep-alive connection may sit unused.
	IdleTimeout time.Duration
	// ShutdownTimeout bounds graceful shutdown.
	ShutdownTimeout time.Duration
	// CORSOrigin is the single origin allowed to read the log from a browser,
	// i.e. the audit explorer. Empty disables CORS entirely.
	CORSOrigin string
	// LogLevel is one of debug, info, warn, error.
	LogLevel string
}

// Load reads the configuration from the environment and applies defaults.
//
// A malformed value is an error, not a fallback. Silently substituting the
// default for an unparseable checkpoint interval means an operator who typed
// `30` instead of `30s` gets a log that looks configured and is not.
func Load() (*Config, error) {
	var errs []error
	dur := func(key string, def time.Duration) time.Duration {
		d, err := envDuration(key, def)
		if err != nil {
			errs = append(errs, err)
		}
		return d
	}
	num := func(key string, def uint) uint {
		n, err := envUint(key, def)
		if err != nil {
			errs = append(errs, err)
		}
		return n
	}

	c := &Config{
		Addr:               env("FARO_ADDR", ":2025"),
		StorageDir:         env("FARO_STORAGE_DIR", "./data/log"),
		SigningKey:         os.Getenv("FARO_SIGNING_KEY"),
		CheckpointInterval: dur("FARO_CHECKPOINT_INTERVAL", 10*time.Second),
		BatchMaxAge:        dur("FARO_BATCH_MAX_AGE", time.Second),
		BatchMaxSize:       num("FARO_BATCH_MAX_SIZE", 256),
		ReadHeaderTimeout:  dur("FARO_READ_HEADER_TIMEOUT", 10*time.Second),
		ReadTimeout:        dur("FARO_READ_TIMEOUT", 30*time.Second),
		WriteTimeout:       dur("FARO_WRITE_TIMEOUT", 60*time.Second),
		IdleTimeout:        dur("FARO_IDLE_TIMEOUT", 120*time.Second),
		ShutdownTimeout:    dur("FARO_SHUTDOWN_TIMEOUT", 15*time.Second),
		CORSOrigin:         os.Getenv("FARO_CORS_ORIGIN"),
		LogLevel:           env("FARO_LOG_LEVEL", "info"),
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
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

// envDuration parses a duration, rejecting anything malformed or non-positive.
// A zero or negative timeout disables the protection it was meant to provide, so
// it is refused rather than applied.
func envDuration(key string, def time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def, fmt.Errorf("config: %s=%q is not a duration (want a form like 10s or 1m): %w", key, v, err)
	}
	if d <= 0 {
		return def, fmt.Errorf("config: %s=%q must be positive; a non-positive value disables the limit", key, v)
	}
	return d, nil
}

func envUint(key string, def uint) (uint, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	n, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return def, fmt.Errorf("config: %s=%q is not a non-negative integer: %w", key, v, err)
	}
	return uint(n), nil
}
