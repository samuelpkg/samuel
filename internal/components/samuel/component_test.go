package samuel

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/ar4mirez/samuel/internal/plugin"
)

// fixtureFS returns a deterministic two-skill source tree used by tests
// that exercise the install pipeline without touching the real
// builtins.FS().
func fixtureFS() fstest.MapFS {
	return fstest.MapFS{
		"ralph/SKILL.md":        {Data: []byte("---\nname: ralph\n---\nbody\n")},
		"create-skill/SKILL.md": {Data: []byte("---\nname: create-skill\n---\nbody\n")},
	}
}

func newComponent(t *testing.T) (*Component, string) {
	t.Helper()
	home := t.TempDir()
	return New(fixtureFS(), home, "2.0.0-alpha.2"), home
}

func TestDetect_ReportsNotInstalledWhenAbsent(t *testing.T) {
	c, home := newComponent(t)
	d, err := c.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if d.Installed {
		t.Errorf("expected Installed=false on empty home")
	}
	want := filepath.Join(home, globalDir)
	if d.Path != want {
		t.Errorf("Detect.Path = %q, want %q", d.Path, want)
	}
}

func TestDetect_TreatsEmptyDirAsNotInstalled(t *testing.T) {
	c, home := newComponent(t)
	if err := os.MkdirAll(filepath.Join(home, globalDir), 0o700); err != nil {
		t.Fatalf("seed: %v", err)
	}
	d, err := c.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if d.Installed {
		t.Errorf("empty dir must not count as installed (would mask crashed installs)")
	}
}

func TestInstall_WritesEverySkill(t *testing.T) {
	c, home := newComponent(t)
	res, err := c.Install(context.Background(), plugin.InstallOptions{})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if res.AlreadyInstalled {
		t.Errorf("first install should not be no-op")
	}
	for _, name := range []string{"ralph", "create-skill"} {
		p := filepath.Join(home, globalDir, name, "SKILL.md")
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected %s, got %v", p, err)
		}
	}
	if len(res.Mutations) != 1 || res.Mutations[0].Kind != plugin.MutationDirCreated {
		t.Errorf("expected one MutationDirCreated, got %+v", res.Mutations)
	}
}

func TestInstall_IdempotentByContentHash(t *testing.T) {
	c, _ := newComponent(t)
	if _, err := c.Install(context.Background(), plugin.InstallOptions{}); err != nil {
		t.Fatalf("first install: %v", err)
	}
	res, err := c.Install(context.Background(), plugin.InstallOptions{})
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	if !res.AlreadyInstalled {
		t.Errorf("second install should be AlreadyInstalled (content hash matched)")
	}
}

func TestInstall_ForceReinstalls(t *testing.T) {
	c, home := newComponent(t)
	if _, err := c.Install(context.Background(), plugin.InstallOptions{}); err != nil {
		t.Fatalf("first install: %v", err)
	}
	// Mutate the on-disk tree so the second install actually writes.
	target := filepath.Join(home, globalDir, "ralph", "SKILL.md")
	if err := os.WriteFile(target, []byte("tampered"), 0o600); err != nil {
		t.Fatalf("seed tamper: %v", err)
	}
	res, err := c.Install(context.Background(), plugin.InstallOptions{Force: true})
	if err != nil {
		t.Fatalf("force install: %v", err)
	}
	if res.AlreadyInstalled {
		t.Errorf("Force should re-sync even when hash mismatch would skip")
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read after force: %v", err)
	}
	if string(got) == "tampered" {
		t.Errorf("force install should have overwritten tampered content")
	}
}

func TestInstall_DryRunMakesNoChanges(t *testing.T) {
	c, home := newComponent(t)
	if _, err := c.Install(context.Background(), plugin.InstallOptions{DryRun: true}); err != nil {
		t.Fatalf("DryRun: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, globalDir)); !os.IsNotExist(err) {
		t.Errorf("DryRun must not create the target tree; stat=%v", err)
	}
}

func TestInstall_AtomicSwapPreservesPreviousOnFailure(t *testing.T) {
	c, home := newComponent(t)
	// Seed an "old" install.
	if _, err := c.Install(context.Background(), plugin.InstallOptions{}); err != nil {
		t.Fatalf("seed install: %v", err)
	}
	// Swap source to a poisoned FS whose entry is a non-local path so
	// syncFS triggers a path-traversal error mid-stage. The on-disk
	// tree must be left intact (the failed install's tmp dir is the
	// only thing the orchestrator should be cleaning up).
	bad := fstest.MapFS{
		"ralph/SKILL.md":  {Data: []byte("---\nname: ralph\n---\nbody\n")},
		"../escape.md":    {Data: []byte("escape attempt")},
	}
	c.Source = bad
	if _, err := c.Install(context.Background(), plugin.InstallOptions{Force: true}); err == nil {
		t.Fatal("expected install to reject path traversal")
	}
	// Live tree still intact.
	live := filepath.Join(home, globalDir, "ralph", "SKILL.md")
	if _, err := os.Stat(live); err != nil {
		t.Errorf("live tree should survive failed install; stat = %v", err)
	}
}

func TestInstall_RejectsNonLocalPath(t *testing.T) {
	bad := fstest.MapFS{"../etc/passwd": {Data: []byte("oops")}}
	c := New(bad, t.TempDir(), "v0")
	if _, err := c.Install(context.Background(), plugin.InstallOptions{}); err == nil {
		t.Error("expected path-traversal rejection")
	}
}

func TestCheck_ReportsHealthAfterInstall(t *testing.T) {
	c, _ := newComponent(t)
	if _, err := c.Install(context.Background(), plugin.InstallOptions{}); err != nil {
		t.Fatalf("install: %v", err)
	}
	hs := c.Check(context.Background())
	if !hs.OK {
		t.Errorf("post-install Check should report OK; got %+v", hs)
	}
	if hs.Component != Name {
		t.Errorf("Check.Component = %q, want %q", hs.Component, Name)
	}
}

func TestCheck_ReportsMissingBeforeInstall(t *testing.T) {
	c, _ := newComponent(t)
	hs := c.Check(context.Background())
	if hs.OK {
		t.Errorf("Check before install should report not-OK")
	}
	if hs.FixHint == "" {
		t.Errorf("Check should suggest a fix when unhealthy")
	}
}

func TestUninstall_RemovesGlobalTree(t *testing.T) {
	c, home := newComponent(t)
	if _, err := c.Install(context.Background(), plugin.InstallOptions{}); err != nil {
		t.Fatalf("install: %v", err)
	}
	if _, err := c.Uninstall(context.Background(), plugin.UninstallOptions{All: true}); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, globalDir)); !os.IsNotExist(err) {
		t.Errorf("global tree should be removed after All uninstall")
	}
}

func TestUninstall_SkippedWithoutGlobalOrAll(t *testing.T) {
	c, _ := newComponent(t)
	res, err := c.Uninstall(context.Background(), plugin.UninstallOptions{Project: true})
	if err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if !res.Skipped {
		t.Errorf("Project-only uninstall should be Skipped (component owns no project state)")
	}
}

func TestManifest_DeclaresBuiltinKind(t *testing.T) {
	c, _ := newComponent(t)
	m := c.Manifest()
	if m.Kind != plugin.KindBuiltin {
		t.Errorf("Manifest.Kind = %q, want %q", m.Kind, plugin.KindBuiltin)
	}
	if m.Name != Name {
		t.Errorf("Manifest.Name = %q, want %q", m.Name, Name)
	}
	if m.Version == "" {
		t.Errorf("Manifest.Version should match the constructor version")
	}
}

func TestNoClaudeWritesAnywhere(t *testing.T) {
	// Invariant: this component must never write to .claude/* (RFD 0009).
	c, home := newComponent(t)
	if _, err := c.Install(context.Background(), plugin.InstallOptions{}); err != nil {
		t.Fatalf("install: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude")); !os.IsNotExist(err) {
		t.Errorf("SamuelComponent must not create ~/.claude/; stat = %v", err)
	}
}
