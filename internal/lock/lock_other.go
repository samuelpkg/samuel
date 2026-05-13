//go:build !unix

package lock

import (
	"runtime"

	"github.com/samuelpkg/samuel/internal/errors"
)

// acquireFileLock on non-Unix platforms returns an unsupported-platform
// error. v2 targets macOS and Linux only — Windows support is on the
// post-2.0 roadmap. Goreleaser drops the Windows build matrix, but if
// a Windows binary somehow gets built and run, this stub fails fast
// with a clear message rather than silently breaking.
//
// runtime.GOOS is included in the message so the user sees "GOOS=js"
// or "GOOS=windows" rather than a Windows-specific hint regardless of
// which !unix target they hit.
func acquireFileLock(home string) (release func(), err error) {
	_ = home
	return nil, &errors.Error{
		Component:   Component,
		Problem:     "Samuel v2.0 does not support GOOS=" + runtime.GOOS,
		Cause:       "the lock subsystem requires flock(2), unavailable on this platform",
		Fix:         "run Samuel on macOS or Linux (Windows is on the v2.x roadmap)",
		DocsURL:     "https://samuelpkg.github.io/samuel/docs/v2-roadmap#windows",
		Recoverable: false,
	}
}
