package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedTree builds a fixture project under t.TempDir(). Layout:
//
//	root/
//	  cmd/main.go
//	  internal/foo/foo.go
//	  internal/foo/foo_test.go
//	  docs/README.md
//	  .git/HEAD              (skipped — hidden dir)
//	  vendor/skip.go         (skipped — embedded skip set)
//	  user-customized/AGENTS.md  (no autogen marker — must be skipped)
func seedTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	must := func(p, content string) {
		full := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", full, err)
		}
	}
	must("cmd/main.go", "package main\nfunc main(){}\n")
	must("internal/foo/foo.go", "package foo\n")
	must("internal/foo/foo_test.go", "package foo\nimport \"testing\"\nfunc TestX(t *testing.T){}\n")
	must("docs/README.md", "# Docs\n")
	must(".git/HEAD", "ref: refs/heads/main\n")
	must("vendor/skip.go", "package vendor\n")
	must("user-customized/AGENTS.md", "# Custom (no marker)\nUser-authored body.\n")
	must("go.mod", "module example\n")
	return root
}

func TestSync_CreatesAgentsMDInWalkedDirs(t *testing.T) {
	root := seedTree(t)
	res, err := SyncFolderContext(Options{RootDir: root, MaxDepth: -1})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if len(res.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	// Expect AGENTS.md in cmd/, internal/, internal/foo/, docs/ (created).
	for _, p := range []string{
		"cmd/AGENTS.md",
		"internal/AGENTS.md",
		"internal/foo/AGENTS.md",
		"docs/AGENTS.md",
	} {
		full := filepath.Join(root, p)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("expected %s, got %v", full, err)
		}
	}
	// Hidden + skip-set dirs must not get AGENTS.md.
	for _, p := range []string{".git/AGENTS.md", "vendor/AGENTS.md"} {
		if _, err := os.Stat(filepath.Join(root, p)); !os.IsNotExist(err) {
			t.Errorf("%s should not be created (skip-set / hidden)", p)
		}
	}
}

func TestSync_NeverWritesCLAUDEmd(t *testing.T) {
	// PRD 0002 invariant: AGENTS.md only — no CLAUDE.md anywhere.
	root := seedTree(t)
	if _, err := SyncFolderContext(Options{RootDir: root, MaxDepth: -1}); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	err := filepath.WalkDir(root, func(path string, _ os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if filepath.Base(path) == "CLAUDE.md" {
			t.Errorf("CLAUDE.md must never be written by v2; found at %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}

func TestSync_SkipsUserCustomizedWithoutMarker(t *testing.T) {
	root := seedTree(t)
	res, err := SyncFolderContext(Options{RootDir: root, MaxDepth: -1})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	// user-customized/AGENTS.md should appear in Skipped, not Updated.
	target := filepath.Join(root, "user-customized/AGENTS.md")
	for _, u := range res.Updated {
		if u == target {
			t.Errorf("user-customized AGENTS.md must NOT be updated without --force")
		}
	}
	body, _ := os.ReadFile(target)
	if !strings.Contains(string(body), "User-authored body") {
		t.Errorf("user body should be preserved; got %q", string(body))
	}
}

func TestSync_ForceOverwritesUserCustomized(t *testing.T) {
	root := seedTree(t)
	if _, err := SyncFolderContext(Options{RootDir: root, MaxDepth: -1, Force: true}); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(root, "user-customized/AGENTS.md"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(body), AutoGenMarker) {
		t.Errorf("--force should overwrite user-customized files; body = %q", string(body))
	}
}

func TestSync_DryRunMakesNoWrites(t *testing.T) {
	root := seedTree(t)
	res, err := SyncFolderContext(Options{RootDir: root, MaxDepth: -1, DryRun: true})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if len(res.Created) == 0 {
		t.Errorf("DryRun should still report Created entries (planned)")
	}
	if _, err := os.Stat(filepath.Join(root, "cmd/AGENTS.md")); !os.IsNotExist(err) {
		t.Errorf("DryRun must not create files; cmd/AGENTS.md exists")
	}
}

func TestSync_AutogenMarkerIsUpdated(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cmd"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "cmd/main.go"), []byte("package main"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	old := "# cmd\n" + AutoGenMarker + ". old body -->\n"
	if err := os.WriteFile(filepath.Join(root, "cmd/AGENTS.md"), []byte(old), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	res, err := SyncFolderContext(Options{RootDir: root, MaxDepth: -1})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if len(res.Updated) == 0 {
		t.Errorf("expected an Updated entry for the autogen-marked file")
	}
}

func TestSync_OverridesExtendDefaults(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "myDomainStuff"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "myDomainStuff/x.go"), []byte("package x"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	ov := Overrides{Folders: map[string]string{"myDomainStuff": "Custom domain logic."}}
	if _, err := SyncFolderContext(Options{RootDir: root, MaxDepth: -1, Overrides: ov}); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	body, _ := os.ReadFile(filepath.Join(root, "myDomainStuff/AGENTS.md"))
	if !strings.Contains(string(body), "Custom domain logic.") {
		t.Errorf("override purpose not applied; body = %q", string(body))
	}
}

func TestSync_HooksAreInvokedInOrder(t *testing.T) {
	// Wire a HookSet manually to verify the sync stages call each hook
	// at the right boundary. Until methodology bodies arrive (PRD 0004)
	// SyncFolderContext uses defaultHooks() — but the contract is
	// tested directly here against the HookSet methods.
	root := seedTree(t)
	calls := []string{}
	hs := &HookSet{
		BeforeFn:        func(Options) { calls = append(calls, "before") },
		AnalyzeFolderFn: func(*FolderAnalysis) { calls = append(calls, "analyze") },
		WriteAgentsMDFn: func(*FolderAnalysis, *string) { calls = append(calls, "write") },
		AfterFn:         func(*Result) { calls = append(calls, "after") },
	}
	hs.Before(Options{RootDir: root})
	hs.AnalyzeFolder(&FolderAnalysis{})
	body := ""
	hs.WriteAgentsMD(&FolderAnalysis{}, &body)
	hs.After(&Result{})
	want := []string{"before", "analyze", "write", "after"}
	for i, w := range want {
		if calls[i] != w {
			t.Errorf("hook[%d] = %q, want %q", i, calls[i], w)
		}
	}
}

func TestAnalyze_DetectsLanguagesAndTests(t *testing.T) {
	root := seedTree(t)
	a, err := AnalyzeFolder(filepath.Join(root, "internal/foo"), Overrides{})
	if err != nil {
		t.Fatalf("AnalyzeFolder: %v", err)
	}
	if a.Languages["Go"] == 0 {
		t.Errorf("expected Go in language map, got %#v", a.Languages)
	}
	if !a.HasTests {
		t.Errorf("expected HasTests=true (foo_test.go present)")
	}
}
