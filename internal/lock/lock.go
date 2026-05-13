// Package lock provides a cross-process advisory file lock used to
// serialise mutating Samuel operations (install / uninstall / sync).
//
// The lock lives at ~/.samuel/lock. v2's path differs from v1's so
// the two versions can coexist on the same host.
//
// Real implementation is flock(2) on Unix (see lock_unix.go). On every
// other platform the package returns a structured "unsupported" error
// (see lock_other.go); this is intentional and matches v1's behaviour.
package lock

const (
	// Component is the namespace used in structured errors produced by
	// this package.
	Component = "lock"

	// DirName is the per-user directory that holds the lock file. It is
	// relative to the resolved home directory.
	DirName = ".samuel"

	// FileName is the lock file inside DirName.
	FileName = "lock"

	// Path is the canonical relative path of the lock file from $HOME.
	// Helpful for tests that need to set up or inspect lock state.
	Path = DirName + "/" + FileName
)

// Acquire takes an exclusive, non-blocking advisory lock for the user
// whose home directory is home. On success it returns a release func
// the caller MUST call (defer release()). On failure it returns a
// *errors.Error explaining the problem.
//
// Acquire is platform-specific: see lock_unix.go and lock_other.go.
func Acquire(home string) (release func(), err error) {
	return acquireFileLock(home)
}
