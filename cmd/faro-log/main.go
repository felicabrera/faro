// Command faro-log runs the FARO transparency log.
//
// It exposes two things:
//
//   - POST /add, which appends an entry and returns the index assigned to it.
//   - The C2SP tlog-tiles read API, served straight off the filesystem that
//     Tessera's POSIX driver writes: /checkpoint, /tile/... and /tile/entries/...
//
// The read API is deliberately just static files. Every byte an auditor needs is
// on disk in a documented, standard format, so the log can be mirrored, archived
// or served by any web server without running this binary at all. A transparency
// log whose contents can only be read through its own API is one you have to
// trust; this one is not.
//
// Access control is asymmetric on purpose: writes will be authenticated (that
// work belongs with the ÁGORA integration), reads never are. The log is public.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/felicabrera/faro/internal/config"
	farolog "github.com/felicabrera/faro/internal/log"
	"github.com/felicabrera/faro/internal/version"
)

// maxEntryBytes bounds a submission. The tlog-tiles entry-bundle format
// length-prefixes each entry with a uint16, so an entry above 64 KiB could not
// be encoded into a bundle at all; refusing it here gives a clear error instead
// of a failure deep inside the storage layer.
const maxEntryBytes = 65535

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "faro-log: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger := newLogger(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	lg, err := farolog.New(ctx, farolog.Config{
		StorageDir:         cfg.StorageDir,
		SigningKey:         cfg.SigningKey,
		CheckpointInterval: cfg.CheckpointInterval,
		BatchMaxAge:        cfg.BatchMaxAge,
		BatchMaxSize:       cfg.BatchMaxSize,
	})
	if err != nil {
		return err
	}

	logger.Info("starting faro-log",
		"addr", cfg.Addr,
		"origin", lg.Origin,
		"storage", cfg.StorageDir,
		"build", version.Read().String())

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           routes(cfg, lg, logger),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("listening: %w", err)
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logger.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutting down the server: %w", err)
		}
		// Flush pending entries only after the listener is closed, so nothing is
		// accepted and then dropped.
		return lg.Close(shutdownCtx)
	}
}

func routes(cfg *config.Config, lg *farolog.Log, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /add", func(w http.ResponseWriter, r *http.Request) {
		entry, err := io.ReadAll(io.LimitReader(r.Body, maxEntryBytes+1))
		if err != nil {
			http.Error(w, "could not read the request body", http.StatusBadRequest)
			return
		}
		if len(entry) == 0 {
			http.Error(w, "empty entry", http.StatusBadRequest)
			return
		}
		if len(entry) > maxEntryBytes {
			http.Error(w, "entry exceeds the maximum size", http.StatusRequestEntityTooLarge)
			return
		}
		index, err := lg.Add(r.Context(), entry)
		if err != nil {
			logger.ErrorContext(r.Context(), "appending entry", "error", err)
			http.Error(w, "could not append the entry", http.StatusInternalServerError)
			return
		}
		if _, err := fmt.Fprintf(w, "%d", index); err != nil {
			logger.ErrorContext(r.Context(), "writing add response", "error", err)
		}
	})

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if _, err := fmt.Fprintf(w, `{"status":"ok","origin":%q,"build":%q}`+"\n",
			lg.Origin, version.Read().String()); err != nil {
			logger.ErrorContext(r.Context(), "writing health response", "error", err)
		}
	})

	// The tlog-tiles read API, served as static files.
	//
	// Cache headers follow the shape of the data rather than a blanket policy:
	// tiles and entry bundles are immutable once written, so they can be cached
	// forever, while the checkpoint is the one thing that changes and must never
	// be served stale.
	fs := http.FileServer(http.Dir(cfg.StorageDir))
	mux.Handle("GET /checkpoint", cacheControl("no-store", fs))
	mux.Handle("GET /tile/", cacheControl("public, max-age=31536000, immutable", fs))

	return corsForExplorer(cfg.CORSOrigin, mux)
}

func cacheControl(value string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", value)
		next.ServeHTTP(w, r)
	})
}

// corsForExplorer allows the audit explorer to read the log from a browser.
//
// Only the read paths get CORS headers, and only for one configured origin. The
// log is public, so this is not protecting the data; it is keeping /add off the
// list of things a hostile page can reach with the visitor's credentials.
func corsForExplorer(origin string, next http.Handler) http.Handler {
	if origin == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		next.ServeHTTP(w, r)
	})
}

func newLogger(level string) *slog.Logger {
	var l slog.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		l = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: l}))
}
