// Package log wires the Tessera append-only log that FARO is built on.
//
// FARO does not implement a Merkle tree. It configures
// [Tessera](https://github.com/transparency-dev/tessera), the transparency-log
// library that also backs production Certificate Transparency logs, and serves
// the result over the [C2SP tlog-tiles](https://c2sp.org/tlog-tiles) read API.
// Writing our own tree would mean asking auditors to trust a bespoke
// implementation of the one component whose correctness the whole system rests
// on; using the same library and wire format as CT means an auditor can point
// existing, independently written verifiers at FARO.
//
// What this package owns is therefore configuration, not algorithms: which
// storage driver, which signing identity, how often a checkpoint is published.
//
// The checkpoint signer is the trust anchor. Whoever holds that key can sign a
// tree head, so it is the single most sensitive secret in FARO. It is read from
// a file or the environment and never generated on the fly, so that a
// misconfigured deployment fails to start rather than quietly starting a log
// nobody can verify against the published key.
package log

import (
	"context"
	"fmt"
	"time"

	"github.com/transparency-dev/tessera"
	"github.com/transparency-dev/tessera/storage/posix"
	"golang.org/x/mod/sumdb/note"
)

// Config describes the log instance to run.
type Config struct {
	// StorageDir is the root of the POSIX-backed log. The thesis specifies the
	// POSIX driver over our own storage rather than a cloud object store, which
	// keeps the log readable with nothing but a filesystem.
	StorageDir string
	// SigningKey is the log's private note key, in the standard
	// "PRIVATE+KEY+name+hash+base64" form.
	SigningKey string
	// CheckpointInterval is how often a new tree head is published. Shorter
	// means a voter waits less to see their ballot included; longer means fewer
	// checkpoints for witnesses to countersign.
	CheckpointInterval time.Duration
	// BatchMaxAge bounds how long an entry waits to be sequenced.
	BatchMaxAge time.Duration
	// BatchMaxSize bounds how many entries are sequenced together.
	BatchMaxSize uint
}

// Log is a running Tessera appender together with its shutdown function.
type Log struct {
	appender *tessera.Appender
	shutdown func(context.Context) error
	// Origin is the checkpoint origin line, derived from the signing key's name.
	// A verifier that does not pin this will accept a checkpoint from any log,
	// so it is surfaced here to be published alongside the public key.
	Origin string
}

// New opens the storage, verifies the signing identity and starts the appender.
func New(ctx context.Context, cfg Config) (*Log, error) {
	if cfg.StorageDir == "" {
		return nil, fmt.Errorf("log: no storage directory configured")
	}
	signer, err := note.NewSigner(cfg.SigningKey)
	if err != nil {
		return nil, fmt.Errorf("log: parsing the checkpoint signing key: %w", err)
	}

	driver, err := posix.New(ctx, posix.Config{Path: cfg.StorageDir})
	if err != nil {
		return nil, fmt.Errorf("log: opening posix storage at %q: %w", cfg.StorageDir, err)
	}

	opts := tessera.NewAppendOptions().
		WithCheckpointSigner(signer).
		WithCheckpointInterval(cfg.CheckpointInterval).
		WithBatching(cfg.BatchMaxSize, cfg.BatchMaxAge)

	appender, shutdown, _, err := tessera.NewAppender(ctx, driver, opts)
	if err != nil {
		return nil, fmt.Errorf("log: starting the appender: %w", err)
	}

	return &Log{appender: appender, shutdown: shutdown, Origin: signer.Name()}, nil
}

// Add appends an entry and blocks until the log assigns it an index.
//
// The index is an assignment, not a proof. A client that needs certainty must
// fetch an inclusion proof against a published checkpoint, which is not
// immediate: entries are batched and checkpoints are published on an interval.
func (l *Log) Add(ctx context.Context, entry []byte) (uint64, error) {
	idx, err := l.appender.Add(ctx, tessera.NewEntry(entry))()
	if err != nil {
		return 0, fmt.Errorf("log: appending entry: %w", err)
	}
	return idx.Index, nil
}

// Close flushes any pending entries and releases the storage.
func (l *Log) Close(ctx context.Context) error {
	if err := l.shutdown(ctx); err != nil {
		return fmt.Errorf("log: shutting down: %w", err)
	}
	return nil
}
