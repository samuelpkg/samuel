package orchestrator

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/samuelpkg/samuel/internal/errors"
	"github.com/samuelpkg/samuel/internal/lock"
	"github.com/samuelpkg/samuel/internal/plugin"
)

// fakePlugin is the configurable plugin.Plugin implementation used by
// orchestrator tests. Counters are atomic so tests that touch
// concurrent paths (Doctor parallelism) stay race-clean.
type fakePlugin struct {
	name string

	detectFn    func(context.Context) (plugin.DetectResult, error)
	installFn   func(context.Context, plugin.InstallOptions) (plugin.InstallResult, error)
	checkFn     func(context.Context) plugin.HealthStatus
	uninstallFn func(context.Context, plugin.UninstallOptions) (plugin.UninstallResult, error)

	installCalls   atomic.Int64
	checkCalls     atomic.Int64
	uninstallCalls atomic.Int64
}

func (f *fakePlugin) Name() string            { return f.name }
func (f *fakePlugin) Manifest() plugin.Manifest { return plugin.Manifest{Name: f.name, Kind: plugin.KindBuiltin} }
func (f *fakePlugin) Detect(ctx context.Context) (plugin.DetectResult, error) {
	if f.detectFn != nil {
		return f.detectFn(ctx)
	}
	return plugin.DetectResult{Installed: false}, nil
}
func (f *fakePlugin) Install(ctx context.Context, opts plugin.InstallOptions) (plugin.InstallResult, error) {
	f.installCalls.Add(1)
	if f.installFn != nil {
		return f.installFn(ctx, opts)
	}
	return plugin.InstallResult{Component: f.name}, nil
}
func (f *fakePlugin) Check(ctx context.Context) plugin.HealthStatus {
	f.checkCalls.Add(1)
	if f.checkFn != nil {
		return f.checkFn(ctx)
	}
	return plugin.HealthStatus{Component: f.name, OK: true, Message: "healthy"}
}
func (f *fakePlugin) Uninstall(ctx context.Context, opts plugin.UninstallOptions) (plugin.UninstallResult, error) {
	f.uninstallCalls.Add(1)
	if f.uninstallFn != nil {
		return f.uninstallFn(ctx, opts)
	}
	return plugin.UninstallResult{Component: f.name}, nil
}

func newOrchWithTempHome(t *testing.T, plugins ...plugin.Plugin) *Orchestrator {
	t.Helper()
	dir := t.TempDir()
	return New(plugins...).WithHomeDir(dir)
}

func TestNew_ZeroPlugins(t *testing.T) {
	o := newOrchWithTempHome(t)
	results, err := o.Install(context.Background(), plugin.InstallOptions{})
	if err != nil {
		t.Fatalf("Install with no plugins: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
	if got := o.Doctor(context.Background()); len(got) != 0 {
		t.Errorf("Doctor with no plugins: %d statuses", len(got))
	}
}

func TestInstall_HappyPath(t *testing.T) {
	a := &fakePlugin{name: "a"}
	b := &fakePlugin{name: "b"}
	c := &fakePlugin{name: "c"}
	o := newOrchWithTempHome(t, a, b, c)
	results, err := o.Install(context.Background(), plugin.InstallOptions{})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, p := range []*fakePlugin{a, b, c} {
		if got := p.installCalls.Load(); got != 1 {
			t.Errorf("plugin %s install calls = %d, want 1", p.name, got)
		}
	}
}

func TestInstall_RollbackInLIFOOrder(t *testing.T) {
	var order []string
	var mu sync.Mutex
	rec := func(name string) func(context.Context) error {
		return func(context.Context) error {
			mu.Lock()
			defer mu.Unlock()
			order = append(order, name)
			return nil
		}
	}
	a := &fakePlugin{
		name: "a",
		installFn: func(context.Context, plugin.InstallOptions) (plugin.InstallResult, error) {
			return plugin.InstallResult{Mutations: []plugin.Mutation{
				{Path: "a-1", Reverse: rec("a-1")},
				{Path: "a-2", Reverse: rec("a-2")},
			}}, nil
		},
	}
	b := &fakePlugin{
		name: "b",
		installFn: func(context.Context, plugin.InstallOptions) (plugin.InstallResult, error) {
			return plugin.InstallResult{Mutations: []plugin.Mutation{
				{Path: "b-1", Reverse: rec("b-1")},
			}}, nil
		},
	}
	c := &fakePlugin{
		name: "c",
		installFn: func(context.Context, plugin.InstallOptions) (plugin.InstallResult, error) {
			return plugin.InstallResult{}, stderrors.New("c failed")
		},
	}
	o := newOrchWithTempHome(t, a, b, c)
	if _, err := o.Install(context.Background(), plugin.InstallOptions{}); err == nil {
		t.Fatal("expected install error")
	}
	mu.Lock()
	defer mu.Unlock()
	want := []string{"b-1", "a-2", "a-1"}
	if fmt.Sprint(order) != fmt.Sprint(want) {
		t.Errorf("rollback order = %v, want %v", order, want)
	}
}

func TestInstall_2ndOf3Failure_FirstMutationsReversed(t *testing.T) {
	// PRD 0002 acceptance criterion: simulated failure in 2nd of 3
	// components → 1st component's mutations reversed, 3rd never runs.
	var reversed []string
	var ran []string
	a := &fakePlugin{
		name: "a",
		installFn: func(context.Context, plugin.InstallOptions) (plugin.InstallResult, error) {
			ran = append(ran, "a")
			return plugin.InstallResult{Mutations: []plugin.Mutation{
				{Path: "a-1", Reverse: func(context.Context) error { reversed = append(reversed, "a-1"); return nil }},
				{Path: "a-2", Reverse: func(context.Context) error { reversed = append(reversed, "a-2"); return nil }},
			}}, nil
		},
	}
	b := &fakePlugin{
		name: "b",
		installFn: func(context.Context, plugin.InstallOptions) (plugin.InstallResult, error) {
			ran = append(ran, "b")
			return plugin.InstallResult{}, stderrors.New("b failed")
		},
	}
	c := &fakePlugin{
		name: "c",
		installFn: func(context.Context, plugin.InstallOptions) (plugin.InstallResult, error) {
			ran = append(ran, "c")
			return plugin.InstallResult{}, nil
		},
	}
	o := newOrchWithTempHome(t, a, b, c)
	_, err := o.Install(context.Background(), plugin.InstallOptions{})
	if err == nil {
		t.Fatal("expected install error from b")
	}
	if fmt.Sprint(ran) != fmt.Sprint([]string{"a", "b"}) {
		t.Errorf("3rd component must not run after 2nd fails; ran = %v", ran)
	}
	if fmt.Sprint(reversed) != fmt.Sprint([]string{"a-2", "a-1"}) {
		t.Errorf("1st component's mutations must be reversed in LIFO order; got %v", reversed)
	}
}

func TestInstall_FailingPluginPartialMutationsRolledBack(t *testing.T) {
	var reversed []string
	a := &fakePlugin{
		name: "a",
		installFn: func(context.Context, plugin.InstallOptions) (plugin.InstallResult, error) {
			return plugin.InstallResult{Mutations: []plugin.Mutation{
				{Path: "a-1", Reverse: func(context.Context) error { reversed = append(reversed, "a-1"); return nil }},
			}}, nil
		},
	}
	b := &fakePlugin{
		name: "b",
		installFn: func(context.Context, plugin.InstallOptions) (plugin.InstallResult, error) {
			return plugin.InstallResult{Mutations: []plugin.Mutation{
				{Path: "b-partial", Reverse: func(context.Context) error { reversed = append(reversed, "b-partial"); return nil }},
			}}, stderrors.New("b failed mid-install")
		},
	}
	o := newOrchWithTempHome(t, a, b)
	if _, err := o.Install(context.Background(), plugin.InstallOptions{}); err == nil {
		t.Fatal("expected error")
	}
	want := []string{"b-partial", "a-1"}
	if fmt.Sprint(reversed) != fmt.Sprint(want) {
		t.Errorf("rollback = %v, want %v", reversed, want)
	}
}

func TestInstall_NilReverseSkipped(t *testing.T) {
	a := &fakePlugin{
		name: "a",
		installFn: func(context.Context, plugin.InstallOptions) (plugin.InstallResult, error) {
			return plugin.InstallResult{Mutations: []plugin.Mutation{{Path: "no-reverse"}}}, nil
		},
	}
	b := &fakePlugin{
		name: "b",
		installFn: func(context.Context, plugin.InstallOptions) (plugin.InstallResult, error) {
			return plugin.InstallResult{}, stderrors.New("fail")
		},
	}
	o := newOrchWithTempHome(t, a, b)
	if _, err := o.Install(context.Background(), plugin.InstallOptions{}); err == nil {
		t.Fatal("expected error (nil Reverse must not panic)")
	}
}

func TestInstall_RollbackUsesFreshContext(t *testing.T) {
	a := &fakePlugin{
		name: "a",
		installFn: func(context.Context, plugin.InstallOptions) (plugin.InstallResult, error) {
			return plugin.InstallResult{Mutations: []plugin.Mutation{
				{Path: "a-1", Reverse: func(rbCtx context.Context) error {
					if rbCtx.Err() != nil {
						return fmt.Errorf("rollback ctx canceled: %w", rbCtx.Err())
					}
					return nil
				}},
			}}, nil
		},
	}
	b := &fakePlugin{
		name: "b",
		installFn: func(context.Context, plugin.InstallOptions) (plugin.InstallResult, error) {
			return plugin.InstallResult{}, stderrors.New("trigger rollback")
		},
	}
	o := newOrchWithTempHome(t, a, b)
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := o.Install(canceled, plugin.InstallOptions{})
	if err == nil {
		t.Fatal("expected install error")
	}
	if strings.Contains(err.Error(), "rollback ctx canceled") {
		t.Errorf("rollback inherited canceled ctx: %v", err)
	}
}

func TestInstall_DryRunDoesNotCreateLock(t *testing.T) {
	dir := t.TempDir()
	a := &fakePlugin{name: "a"}
	o := New(a).WithHomeDir(dir)
	if _, err := o.Install(context.Background(), plugin.InstallOptions{DryRun: true}); err != nil {
		t.Fatalf("DryRun Install: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, lock.Path)); !os.IsNotExist(err) {
		t.Errorf("DryRun must not create the lock file; stat = %v", err)
	}
	if got := a.installCalls.Load(); got != 1 {
		t.Errorf("DryRun should still call plugin; got %d installs", got)
	}
}

func TestInstall_RollbackFailureWrappedNonRecoverable(t *testing.T) {
	a := &fakePlugin{
		name: "a",
		installFn: func(context.Context, plugin.InstallOptions) (plugin.InstallResult, error) {
			return plugin.InstallResult{Mutations: []plugin.Mutation{
				{Path: "a-1", Reverse: func(context.Context) error { return stderrors.New("cannot reverse") }},
			}}, nil
		},
	}
	b := &fakePlugin{
		name: "b",
		installFn: func(context.Context, plugin.InstallOptions) (plugin.InstallResult, error) {
			return plugin.InstallResult{}, &errors.Error{Component: "b", Problem: "transient", Recoverable: true}
		},
	}
	o := newOrchWithTempHome(t, a, b)
	_, err := o.Install(context.Background(), plugin.InstallOptions{})
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.IsRecoverable(err) {
		t.Errorf("rollback failure must yield non-recoverable error; got recoverable")
	}
	var oe *errors.Error
	if !stderrors.As(err, &oe) {
		t.Fatalf("expected *Error at top of chain, got %T: %v", err, err)
	}
	if !strings.Contains(oe.Problem, "rollback") {
		t.Errorf("top-level error should mention rollback; got %q", oe.Problem)
	}
	if oe.DocsURL == "" {
		t.Errorf("rollback-failure error should carry DocsURL pointing at SAM-ROLLBACK-001")
	}
}

func TestUninstall_ReverseOrderAndJoinErrors(t *testing.T) {
	var order []string
	mk := func(name string, fail bool) *fakePlugin {
		return &fakePlugin{
			name: name,
			uninstallFn: func(context.Context, plugin.UninstallOptions) (plugin.UninstallResult, error) {
				order = append(order, name)
				if fail {
					return plugin.UninstallResult{Component: name}, stderrors.New(name + "-uninstall-failed")
				}
				return plugin.UninstallResult{Component: name}, nil
			},
		}
	}
	a := mk("a", false)
	b := mk("b", true)
	c := mk("c", true)
	o := newOrchWithTempHome(t, a, b, c)
	_, err := o.Uninstall(context.Background(), plugin.UninstallOptions{All: true})
	if err == nil {
		t.Fatal("expected joined error")
	}
	for _, w := range []string{"b-uninstall-failed", "c-uninstall-failed"} {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("joined error missing %q: %v", w, err)
		}
	}
	want := []string{"c", "b", "a"}
	if fmt.Sprint(order) != fmt.Sprint(want) {
		t.Errorf("uninstall order = %v, want %v", order, want)
	}
}

func TestDoctor_DoesNotAcquireLock(t *testing.T) {
	// Two parallel Doctor calls must complete promptly — if Doctor took
	// the advisory lock they would serialize and the second would block
	// on flock until the first released.
	c := &fakePlugin{name: "c"}
	o := newOrchWithTempHome(t, c)
	done := make(chan struct{}, 2)
	for i := 0; i < 2; i++ {
		go func() {
			o.Doctor(context.Background())
			done <- struct{}{}
		}()
	}
	timeout := time.After(2 * time.Second)
	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-timeout:
			t.Fatalf("Doctor calls did not complete; lock may have been acquired")
		}
	}
}

func TestDoctor_PopulatesComponentNameWhenMissing(t *testing.T) {
	c := &fakePlugin{
		name: "samuel-builtins",
		checkFn: func(context.Context) plugin.HealthStatus {
			return plugin.HealthStatus{OK: true, Message: "ok"}
		},
	}
	o := New(c)
	statuses := o.Doctor(context.Background())
	if statuses[0].Component != "samuel-builtins" {
		t.Errorf("Doctor should fill Component from Name(); got %q", statuses[0].Component)
	}
}

func TestLock_SerializesInstallsButNotDoctor(t *testing.T) {
	// Sanity check: two successive installs both succeed (flock auto-
	// releases on close); a Doctor call between them does not block.
	dir := t.TempDir()
	c := &fakePlugin{name: "c"}
	o := New(c).WithHomeDir(dir)
	if _, err := o.Install(context.Background(), plugin.InstallOptions{}); err != nil {
		t.Fatalf("first install: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, lock.Path)); err != nil {
		t.Errorf("lock file should persist after release; stat = %v", err)
	}
	_ = o.Doctor(context.Background())
	if _, err := o.Install(context.Background(), plugin.InstallOptions{}); err != nil {
		t.Errorf("second install: %v", err)
	}
}
