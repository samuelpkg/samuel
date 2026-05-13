package config

import (
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/ar4mirez/samuel/internal/errors"
)

// Save writes cfg to samuel.toml in dir atomically. It creates dir if
// necessary. The write is staged through a same-directory temp file
// and rename(2)'d into place so a crash mid-write cannot leave a
// half-written config — a class of failure the v1 orchestrator port
// fought hard to eliminate.
func Save(dir string, cfg *Config) error {
	return saveAtomic(dir, ProjectFile, cfg, "samuel.toml")
}

// SaveLock writes lf atomically (see Save).
func SaveLock(dir string, lf *Lockfile) error {
	return saveAtomic(dir, LockFile, lf, "samuel.lock")
}

func saveAtomic(dir, name string, v any, label string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return (&errors.Error{
			Component:   component,
			Problem:     "cannot create config directory",
			Fix:         "check write permissions on " + dir,
			Recoverable: true,
			Path:        dir,
		}).Wrap(err)
	}
	target := filepath.Join(dir, name)
	tmp, err := os.CreateTemp(dir, "."+name+".tmp.*")
	if err != nil {
		return (&errors.Error{
			Component:   component,
			Problem:     "cannot create temp file for " + label,
			Fix:         "check write permissions on " + dir,
			Recoverable: true,
			Path:        dir,
		}).Wrap(err)
	}
	tmpName := tmp.Name()
	// Best-effort cleanup if we bail before rename succeeds.
	cleanup := func() { _ = os.Remove(tmpName) }
	defer func() {
		if cleanup != nil {
			cleanup()
		}
	}()

	enc := toml.NewEncoder(tmp)
	enc.SetIndentTables(true)
	if err := enc.Encode(v); err != nil {
		_ = tmp.Close()
		return (&errors.Error{
			Component:   component,
			Problem:     "cannot encode " + label,
			Cause:       err.Error(),
			Recoverable: false,
			Path:        target,
		}).Wrap(err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return (&errors.Error{
			Component:   component,
			Problem:     "cannot fsync " + label,
			Recoverable: true,
			Path:        target,
		}).Wrap(err)
	}
	if err := tmp.Close(); err != nil {
		return (&errors.Error{
			Component:   component,
			Problem:     "cannot close temp " + label,
			Recoverable: true,
			Path:        tmpName,
		}).Wrap(err)
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return (&errors.Error{
			Component:   component,
			Problem:     "cannot chmod " + label,
			Recoverable: true,
			Path:        tmpName,
		}).Wrap(err)
	}
	if err := os.Rename(tmpName, target); err != nil {
		return (&errors.Error{
			Component:   component,
			Problem:     "cannot rename " + label + " into place",
			Recoverable: true,
			Path:        target,
		}).Wrap(err)
	}
	cleanup = nil
	return nil
}
