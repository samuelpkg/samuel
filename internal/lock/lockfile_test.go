package lock

import (
	"context"
	"reflect"
	"testing"

	"github.com/ar4mirez/samuel/internal/config"
	"github.com/ar4mirez/samuel/internal/plugin"
)

func TestReadLockfile_MissingReturnsEmpty(t *testing.T) {
	lf, err := ReadLockfile(t.TempDir())
	if err != nil {
		t.Fatalf("ReadLockfile on missing file: %v", err)
	}
	if lf == nil {
		t.Fatalf("expected non-nil Lockfile for missing file")
	}
	if lf.Version == "" {
		t.Errorf("expected default Version stamped for fresh lockfile")
	}
	if len(lf.Mutations) != 0 {
		t.Errorf("expected empty mutations, got %d", len(lf.Mutations))
	}
}

func TestRecordMutations_AppendsAndPersists(t *testing.T) {
	dir := t.TempDir()

	first := []plugin.Mutation{
		{Kind: plugin.MutationDirCreated, Path: "/tmp/a", Description: "made a"},
		{Kind: plugin.MutationFileWritten, Path: "/tmp/a/x.md", Description: "wrote x"},
	}
	if err := RecordMutations(dir, "samuel-builtins", first); err != nil {
		t.Fatalf("RecordMutations first: %v", err)
	}

	// Second batch from a different producer.
	second := []plugin.Mutation{
		{Kind: plugin.MutationSymlinkCreated, Path: "/proj/.samuel/builtins", Description: "linked"},
	}
	if err := RecordMutations(dir, "samuel-init", second); err != nil {
		t.Fatalf("RecordMutations second: %v", err)
	}

	lf, err := ReadLockfile(dir)
	if err != nil {
		t.Fatalf("ReadLockfile: %v", err)
	}
	if got := len(lf.Mutations); got != 3 {
		t.Fatalf("expected 3 mutation records, got %d (%v)", got, lf.Mutations)
	}
	wantKinds := []string{
		string(plugin.MutationDirCreated),
		string(plugin.MutationFileWritten),
		string(plugin.MutationSymlinkCreated),
	}
	gotKinds := []string{lf.Mutations[0].Kind, lf.Mutations[1].Kind, lf.Mutations[2].Kind}
	if !reflect.DeepEqual(gotKinds, wantKinds) {
		t.Errorf("mutation kind order = %v, want %v", gotKinds, wantKinds)
	}
	if lf.Mutations[2].Plugin != "samuel-init" {
		t.Errorf("third record producer = %q, want samuel-init", lf.Mutations[2].Plugin)
	}
	if lf.Mutations[0].AppliedAt == "" {
		t.Errorf("AppliedAt should be stamped on every record")
	}
}

func TestRecordMutations_EmptyBatchIsNoop(t *testing.T) {
	dir := t.TempDir()
	if err := RecordMutations(dir, "x", nil); err != nil {
		t.Fatalf("empty batch should not error; got %v", err)
	}
	// Nothing should be written for an empty batch (file does not exist).
	if _, err := config.LoadLock(dir); err == nil {
		t.Errorf("empty batch should not create a lockfile")
	}
}

func TestToRecord_SerializeFields(t *testing.T) {
	m := plugin.Mutation{
		Kind:        plugin.MutationCommandRun,
		Path:        "/usr/bin/example",
		Description: "ran example",
		Reverse: func(context.Context) error {
			return nil
		},
	}
	rec := ToRecord("test-producer", m, "2026-05-12T20:00:00Z")
	want := config.MutationRecord{
		Plugin:      "test-producer",
		Kind:        "command_run",
		Path:        "/usr/bin/example",
		Description: "ran example",
		AppliedAt:   "2026-05-12T20:00:00Z",
	}
	if !reflect.DeepEqual(rec, want) {
		t.Errorf("ToRecord =\n got: %#v\nwant: %#v", rec, want)
	}
}

func TestWriteLockfile_StampsTimestamps(t *testing.T) {
	dir := t.TempDir()
	lf := &config.Lockfile{Plugins: []config.LockedPlugin{{Name: "x", Version: "1.0.0", Kind: "skill"}}}
	if err := WriteLockfile(dir, lf); err != nil {
		t.Fatalf("WriteLockfile: %v", err)
	}
	if lf.GeneratedAt == "" {
		t.Errorf("WriteLockfile should stamp GeneratedAt")
	}
	if lf.Version == "" {
		t.Errorf("WriteLockfile should stamp Version")
	}
	round, err := ReadLockfile(dir)
	if err != nil {
		t.Fatalf("ReadLockfile: %v", err)
	}
	if len(round.Plugins) != 1 || round.Plugins[0].Name != "x" {
		t.Errorf("round-trip lost plugin entries: %+v", round.Plugins)
	}
}
