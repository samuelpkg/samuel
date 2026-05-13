package builtins

import (
	"io/fs"
	"sort"
	"strings"
	"testing"
)

func TestFS_ContainsCanonicalBuiltins(t *testing.T) {
	root := FS()
	got := topLevelDirs(t, root)
	want := SkillNames()
	sort.Strings(got)
	wantSorted := append([]string(nil), want...)
	sort.Strings(wantSorted)
	if !equalSlice(got, wantSorted) {
		t.Errorf("top-level builtins = %v, want %v", got, wantSorted)
	}
}

func TestFS_EverySkillHasManifest(t *testing.T) {
	root := FS()
	for _, name := range SkillNames() {
		manifest := name + "/SKILL.md"
		b, err := fs.ReadFile(root, manifest)
		if err != nil {
			t.Errorf("missing %s: %v", manifest, err)
			continue
		}
		if !strings.HasPrefix(string(b), "---\n") {
			t.Errorf("%s should start with YAML frontmatter `---\\n`", manifest)
		}
		if !strings.Contains(string(b), "name: "+name) {
			t.Errorf("%s frontmatter should declare name: %q", manifest, name)
		}
	}
}

func TestSkillNames_IsDeterministic(t *testing.T) {
	first := SkillNames()
	second := SkillNames()
	if !equalSlice(first, second) {
		t.Errorf("SkillNames must be deterministic; got %v vs %v", first, second)
	}
}

func topLevelDirs(t *testing.T, root fs.FS) []string {
	t.Helper()
	entries, err := fs.ReadDir(root, ".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out
}

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
