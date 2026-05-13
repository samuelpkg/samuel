//go:build unix

package lock

import (
	stderrors "errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/samuelpkg/samuel/internal/errors"
)

// holdFlock takes an exclusive flock on the lock file at the given home
// dir from the test process. Returns a release function the caller MUST
// call. Used to simulate a live cross-process lock holder.
func holdFlock(t *testing.T, home string) func() {
	t.Helper()
	lockFile := filepath.Join(home, Path)
	if err := os.MkdirAll(filepath.Dir(lockFile), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	f, err := os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open lock: %v", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		t.Fatalf("flock: %v", err)
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}
}

func TestReadHolderHint(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		name     string
		body     []byte
		write    bool
		expected string
	}{
		{"missing", nil, false, "unknown"},
		{"empty", []byte(""), true, "unknown"},
		{"whitespace", []byte("   \n\t"), true, "unknown"},
		{"non-numeric", []byte("not-a-pid"), true, "unknown"},
		{"zero", []byte("0"), true, "unknown"},
		{"negative", []byte("-1"), true, "unknown"},
		{"valid", []byte("12345"), true, "pid=12345"},
		{"valid-trailing-newline", []byte("12345\n"), true, "pid=12345"},
		// 64KB blob — must not be fully read (bounded reader caps at 32 bytes)
		// and must not parse as a PID.
		{"large-blob", append([]byte("12345"), make([]byte, 65536)...), true, "unknown"},
		// ANSI escape sequence — must not flow into output.
		{"ansi-escape", []byte("\x1b[31mevil\x1b[0m"), true, "unknown"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, tc.name+".lock")
			if tc.write {
				if err := os.WriteFile(path, tc.body, 0o600); err != nil {
					t.Fatalf("write: %v", err)
				}
			}
			got := readHolderHint(path)
			if got != tc.expected {
				t.Errorf("readHolderHint(%q) = %q, want %q", tc.name, got, tc.expected)
			}
		})
	}
}

func TestAcquire_LiveLockReturnsBusy(t *testing.T) {
	dir := t.TempDir()
	releaseHeld := holdFlock(t, dir)
	defer releaseHeld()

	release, err := Acquire(dir)
	if err == nil {
		release()
		t.Fatalf("expected lock-busy error while flock is held, got nil")
	}
	var oe *errors.Error
	if !stderrors.As(err, &oe) {
		t.Fatalf("expected *errors.Error, got %T: %v", err, err)
	}
	if !oe.Recoverable {
		t.Errorf("lock-busy error should be Recoverable")
	}
	if oe.DocsURL == "" {
		t.Errorf("lock-busy error should have DocsURL")
	}
}

func TestAcquire_ReleasedThenAcquired(t *testing.T) {
	// Hold flock, release it, then verify Acquire succeeds. Proves the
	// kernel-managed release path works end-to-end.
	dir := t.TempDir()
	holdRelease := holdFlock(t, dir)
	holdRelease()

	release, err := Acquire(dir)
	if err != nil {
		t.Fatalf("Acquire should succeed after test lock released; got %v", err)
	}
	release()
}

func TestAcquire_MkdirFailure_StructuredError(t *testing.T) {
	// MkdirAll fails when a non-directory exists at the parent path.
	// Place a regular file at <home>/.samuel so MkdirAll cannot create
	// the lock directory; acquireFileLock must return a structured
	// *Error with Recoverable=true.
	dir := t.TempDir()
	clashPath := filepath.Join(dir, DirName)
	if err := os.WriteFile(clashPath, []byte("not a dir"), 0o600); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	_, err := Acquire(dir)
	if err == nil {
		t.Fatalf("expected mkdir failure error, got nil")
	}
	var oe *errors.Error
	if !stderrors.As(err, &oe) {
		t.Fatalf("expected *errors.Error, got %T: %v", err, err)
	}
	if !strings.Contains(oe.Problem, "lock directory") {
		t.Errorf("expected mkdir error Problem to mention lock directory, got %q", oe.Problem)
	}
	if !oe.Recoverable {
		t.Errorf("mkdir error should be Recoverable")
	}
}

func TestAcquire_ConcurrentSerializes(t *testing.T) {
	// Two goroutines calling Acquire on the same home dir must
	// serialise via flock. Second caller fails fast with the structured
	// lock-busy error while the first holds the lock.
	dir := t.TempDir()

	firstHeld := make(chan struct{})
	gate := make(chan struct{})

	var wg sync.WaitGroup
	var (
		err1, err2 error
		mu         sync.Mutex
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		release, e := Acquire(dir)
		mu.Lock()
		err1 = e
		mu.Unlock()
		if e == nil {
			close(firstHeld)
			<-gate
			release()
		} else {
			close(firstHeld)
		}
	}()

	select {
	case <-firstHeld:
	case <-time.After(2 * time.Second):
		close(gate)
		wg.Wait()
		t.Fatalf("first Acquire never returned")
	}

	release2, e2 := Acquire(dir)
	err2 = e2
	if release2 != nil {
		release2()
	}
	close(gate)
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if err1 != nil {
		t.Errorf("first Acquire should succeed; got %v", err1)
	}
	if err2 == nil {
		t.Fatalf("second Acquire must fail with lock-busy while first holds the lock")
	}
	var oe *errors.Error
	if !stderrors.As(err2, &oe) {
		t.Fatalf("expected *errors.Error from second Acquire, got %T: %v", err2, err2)
	}
	if !strings.Contains(oe.Problem, "samuel process") {
		t.Errorf("expected lock-busy Problem, got %q", oe.Problem)
	}
}
