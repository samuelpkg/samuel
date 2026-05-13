package config

import (
	stderrors "errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoad_NotFoundReturnsSentinel(t *testing.T) {
	_, err := Load(t.TempDir())
	if !stderrors.Is(err, ErrNotFound) {
		t.Errorf("Load on empty dir should return ErrNotFound; got %v", err)
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := Defaults()
	want.Plugins = []PluginEntry{
		{Name: "go-guide", Version: "1.4.2", Kind: "skill"},
		{Name: "ralph", Version: "2.0.0", Kind: "wasm"},
		{Name: "claude-runner", Version: "1.0.0", Kind: "oci"},
	}
	if err := Save(dir, want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("round-trip mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestSaveLoadLock_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := &Lockfile{
		Version:     SchemaVersion,
		GeneratedAt: "2026-05-12T20:00:00Z",
		TOONSpec:    "3.0",
		Plugins: []LockedPlugin{
			{Name: "go-guide", Version: "1.4.2", Kind: "skill", Digest: "sha256:abc", Signed: true},
		},
		Capabilities: []string{"fs:read", "net:none"},
	}
	if err := SaveLock(dir, want); err != nil {
		t.Fatalf("SaveLock: %v", err)
	}
	got, err := LoadLock(dir)
	if err != nil {
		t.Fatalf("LoadLock: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("lockfile round-trip mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestSave_AtomicRenamesOverExisting(t *testing.T) {
	dir := t.TempDir()
	first := Defaults()
	first.DefaultMethodology = "first"
	if err := Save(dir, first); err != nil {
		t.Fatalf("Save first: %v", err)
	}
	second := Defaults()
	second.DefaultMethodology = "second"
	if err := Save(dir, second); err != nil {
		t.Fatalf("Save second: %v", err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.DefaultMethodology != "second" {
		t.Errorf("DefaultMethodology = %q, want second", got.DefaultMethodology)
	}
	// No leftover temp files in dir.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if filepath.Base(e.Name()) != ProjectFile {
			t.Errorf("unexpected leftover %q in dir", e.Name())
		}
	}
}

func TestLoad_MalformedReturnsStructuredError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ProjectFile), []byte("not = a [valid toml"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	_, err := Load(dir)
	if err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestLoadLock_NotFoundReturnsSentinel(t *testing.T) {
	_, err := LoadLock(t.TempDir())
	if !stderrors.Is(err, ErrNotFound) {
		t.Errorf("LoadLock empty dir should return ErrNotFound; got %v", err)
	}
}

func TestLoadLock_Malformed(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, LockFile), []byte("not [valid toml"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, err := LoadLock(dir); err == nil {
		t.Errorf("expected parse error")
	}
}

func TestSave_FailsOnUnwritableDir(t *testing.T) {
	dir := t.TempDir()
	// Place a file where the config dir should be created — MkdirAll
	// will refuse to convert a regular file into a directory.
	parent := filepath.Join(dir, "blocked")
	if err := os.WriteFile(parent, []byte("x"), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}
	child := filepath.Join(parent, "nested")
	if err := Save(child, Defaults()); err == nil {
		t.Errorf("expected mkdir failure when parent is a file")
	}
}

func TestDefaults_NonEmpty(t *testing.T) {
	c := Defaults()
	if c.Version == "" {
		t.Errorf("Defaults missing version")
	}
	if _, ok := c.Methodology["ralph"]; !ok {
		t.Errorf("Defaults missing ralph methodology")
	}
	if c.Guardrails == nil || c.Guardrails.MaxFunctionLines == 0 {
		t.Errorf("Defaults missing guardrails")
	}
}
