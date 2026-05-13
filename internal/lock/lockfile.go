package lock

import (
	stderrors "errors"
	"time"

	"github.com/ar4mirez/samuel/internal/config"
	"github.com/ar4mirez/samuel/internal/errors"
	"github.com/ar4mirez/samuel/internal/plugin"
)

// LockfileName is the on-disk name of the samuel.lock file.
const LockfileName = config.LockFile

// ReadLockfile returns the parsed samuel.lock at projectDir. A missing
// file is returned as an empty Lockfile (not an error) so callers can
// safely append to a fresh log.
func ReadLockfile(projectDir string) (*config.Lockfile, error) {
	lf, err := config.LoadLock(projectDir)
	if err == nil {
		return lf, nil
	}
	if stderrors.Is(err, config.ErrNotFound) {
		return &config.Lockfile{Version: config.SchemaVersion}, nil
	}
	return nil, err
}

// WriteLockfile saves lf to samuel.lock at projectDir atomically.
// GeneratedAt is stamped if empty.
func WriteLockfile(projectDir string, lf *config.Lockfile) error {
	if lf == nil {
		return &errors.Error{
			Component:   Component,
			Problem:     "cannot write nil lockfile",
			Recoverable: false,
		}
	}
	if lf.Version == "" {
		lf.Version = config.SchemaVersion
	}
	if lf.GeneratedAt == "" {
		lf.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	}
	return config.SaveLock(projectDir, lf)
}

// RecordMutations appends plugin mutations to the project's samuel.lock
// in chronological order. producer is the originating plugin's Name().
// The lockfile is written atomically.
//
// The Reverse closures on plugin.Mutation are NOT serialized — only the
// kind, path, and description survive across processes. Uninstall reads
// the records back and dispatches the appropriate undo per kind.
func RecordMutations(projectDir, producer string, muts []plugin.Mutation) error {
	if len(muts) == 0 {
		return nil
	}
	lf, err := ReadLockfile(projectDir)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	for _, m := range muts {
		lf.Mutations = append(lf.Mutations, ToRecord(producer, m, now))
	}
	return WriteLockfile(projectDir, lf)
}

// ToRecord converts a plugin.Mutation into the serializable form. The
// timestamp is applied externally so a batch of mutations can share a
// single appliedAt (the orchestrator passes time.Now().UTC()).
func ToRecord(producer string, m plugin.Mutation, appliedAt string) config.MutationRecord {
	return config.MutationRecord{
		Plugin:      producer,
		Kind:        string(m.Kind),
		Path:        m.Path,
		Description: m.Description,
		AppliedAt:   appliedAt,
	}
}
