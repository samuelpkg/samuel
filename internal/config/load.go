package config

import (
	stderrors "errors"
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/samuelpkg/samuel/internal/errors"
)

// component is the namespace used in structured errors produced by
// this package.
const component = "config"

// Load reads samuel.toml from dir and returns the parsed Config. When
// the file does not exist, Load returns ErrNotFound — callers can test
// this with errors.Is to fall back to Defaults() during `samuel init`.
func Load(dir string) (*Config, error) {
	path := filepath.Join(dir, ProjectFile)
	b, err := os.ReadFile(path)
	if err != nil {
		if stderrors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, (&errors.Error{
			Component:   component,
			Problem:     "cannot read samuel.toml",
			Fix:         "check the file exists and is readable",
			Recoverable: true,
			Path:        path,
		}).Wrap(err)
	}
	cfg := &Config{}
	if err := toml.Unmarshal(b, cfg); err != nil {
		return nil, (&errors.Error{
			Component:   component,
			Problem:     "cannot parse samuel.toml",
			Cause:       err.Error(),
			Fix:         "fix the TOML syntax (run `samuel doctor` for hints)",
			DocsURL:     "https://samuelpkg.github.io/samuel/docs/errors/SAM-CFG-001",
			Recoverable: true,
			Path:        path,
		}).Wrap(err)
	}
	if cfg.Version == "" {
		cfg.Version = SchemaVersion
	}
	if vErr := cfg.Validate(); vErr != nil {
		return cfg, vErr
	}
	return cfg, nil
}

// LoadLock reads samuel.lock from dir. Like Load, ErrNotFound on a
// missing file is fine — the lockfile is created lazily.
func LoadLock(dir string) (*Lockfile, error) {
	path := filepath.Join(dir, LockFile)
	b, err := os.ReadFile(path)
	if err != nil {
		if stderrors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, (&errors.Error{
			Component:   component,
			Problem:     "cannot read samuel.lock",
			Fix:         "check the file exists and is readable",
			Recoverable: true,
			Path:        path,
		}).Wrap(err)
	}
	lf := &Lockfile{}
	if err := toml.Unmarshal(b, lf); err != nil {
		return nil, (&errors.Error{
			Component:   component,
			Problem:     "cannot parse samuel.lock",
			Cause:       err.Error(),
			Fix:         "regenerate with `samuel install --lock-only` (lockfile is machine-managed)",
			DocsURL:     "https://samuelpkg.github.io/samuel/docs/errors/SAM-CFG-002",
			Recoverable: true,
			Path:        path,
		}).Wrap(err)
	}
	if lf.Version == "" {
		lf.Version = SchemaVersion
	}
	return lf, nil
}

// ErrNotFound is returned by Load/LoadLock when the file does not
// exist. Use stderrors.Is(err, ErrNotFound) to test for it.
var ErrNotFound = stderrors.New("config: file not found")

// String form so stderrors.Is sees a recognizable sentinel even when
// wrapped.
var _ = fmt.Sprintf
