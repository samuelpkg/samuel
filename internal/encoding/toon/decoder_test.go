package toon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnmarshal_RejectsMissingHeader(t *testing.T) {
	if _, _, err := Unmarshal([]byte("foo: bar\n")); err == nil {
		t.Errorf("expected error for missing version header")
	}
}

func TestUnmarshal_RejectsIncompatibleMajor(t *testing.T) {
	src := "# toon v99.0\nfoo: bar\n"
	if _, _, err := Unmarshal([]byte(src)); err == nil {
		t.Errorf("expected error for incompatible major version")
	}
}

func TestUnmarshal_AcceptsMinorDriftWithWarning(t *testing.T) {
	src := "# toon v3.5\nfoo: bar\n"
	root, warns, err := Unmarshal([]byte(src))
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if v, _ := root.Get("foo"); v != "bar" {
		t.Errorf("foo = %v, want bar", v)
	}
	if len(warns) == 0 {
		t.Errorf("expected minor-drift warning")
	}
}

func TestUnmarshal_Scalars(t *testing.T) {
	src := "# toon v3.0\n" +
		"s: hello\n" +
		"n: 42\n" +
		"f: 3.14\n" +
		"b: true\n" +
		"z: null\n"
	root, _, err := Unmarshal([]byte(src))
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if v, _ := root.Get("s"); v != "hello" {
		t.Errorf("s = %v", v)
	}
	if v, _ := root.Get("n"); v != int64(42) {
		t.Errorf("n = %v (%T)", v, v)
	}
	if v, _ := root.Get("f"); v != 3.14 {
		t.Errorf("f = %v", v)
	}
	if v, _ := root.Get("b"); v != true {
		t.Errorf("b = %v", v)
	}
	if v, ok := root.Get("z"); !ok || v != nil {
		t.Errorf("z = %v, ok=%v", v, ok)
	}
}

func TestUnmarshal_NestedObject(t *testing.T) {
	src := "# toon v3.0\n" +
		"task:\n" +
		"  id: 1\n" +
		"  title: Add auth\n"
	root, _, err := Unmarshal([]byte(src))
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	inner, ok := root.Get("task")
	if !ok {
		t.Fatalf("missing task key")
	}
	innerObj, ok := inner.(*Object)
	if !ok {
		t.Fatalf("task is %T, want *Object", inner)
	}
	if v, _ := innerObj.Get("id"); v != int64(1) {
		t.Errorf("id = %v", v)
	}
}

func TestUnmarshal_TabularArray(t *testing.T) {
	src := "# toon v3.0\n" +
		"tasks[2]{id,title,status}:\n" +
		"  1,Add auth,done\n" +
		"  2,\"tests, more\",todo\n"
	root, _, err := Unmarshal([]byte(src))
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	tbl, ok := root.Get("tasks")
	if !ok {
		t.Fatalf("missing tasks")
	}
	ta, ok := tbl.(*TableArray)
	if !ok {
		t.Fatalf("tasks is %T", tbl)
	}
	if len(ta.Rows) != 2 {
		t.Fatalf("rows=%d", len(ta.Rows))
	}
	if ta.Rows[1]["title"] != "tests, more" {
		t.Errorf("row 1 title = %q", ta.Rows[1]["title"])
	}
}

func TestUnmarshal_MalformedRowRecovery(t *testing.T) {
	// Middle row has only 2 cells; decoder must skip with warning and
	// continue to the third row.
	src := "# toon v3.0\n" +
		"tasks[3]{id,title,status}:\n" +
		"  1,Add auth,done\n" +
		"  2,oops\n" +
		"  3,done one,todo\n"
	root, warns, err := Unmarshal([]byte(src))
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	tbl := mustTable(t, root, "tasks")
	if len(tbl.Rows) != 3 {
		t.Fatalf("rows=%d, want 3", len(tbl.Rows))
	}
	if tbl.Rows[1] != nil {
		t.Errorf("malformed row should be nil; got %v", tbl.Rows[1])
	}
	if tbl.Rows[2]["title"] != "done one" {
		t.Errorf("third row not parsed correctly: %v", tbl.Rows[2])
	}
	if len(warns) == 0 {
		t.Errorf("expected a warning for malformed row")
	}
}

func TestUnmarshal_ScalarArrayInline(t *testing.T) {
	src := "# toon v3.0\nxs[3]: 1,2,3\n"
	root, _, err := Unmarshal([]byte(src))
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	arr, _ := root.Get("xs")
	got := arr.([]any)
	if len(got) != 3 || got[0] != int64(1) || got[2] != int64(3) {
		t.Errorf("xs = %v", got)
	}
}

func TestUnmarshal_RejectsTabsInIndent(t *testing.T) {
	src := "# toon v3.0\ntask:\n\tid: 1\n"
	if _, _, err := Unmarshal([]byte(src)); err == nil {
		t.Errorf("expected error for tab indentation")
	}
}

func TestUnmarshal_DuplicateKeyWarning(t *testing.T) {
	src := "# toon v3.0\nfoo: a\nfoo: b\n"
	root, warns, err := Unmarshal([]byte(src))
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if v, _ := root.Get("foo"); v != "b" {
		t.Errorf("foo = %v, want b (last-write-wins)", v)
	}
	if len(warns) == 0 {
		t.Errorf("expected duplicate-key warning")
	}
}

func TestRoundTrip_Golden(t *testing.T) {
	files := []string{"prd-small.toon", "project-snapshot.toon", "task-context.toon", "prd-60-tasks.toon"}
	for _, name := range files {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("testdata", name)
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			root, warns, err := Unmarshal(src)
			if err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if len(warns) != 0 {
				for _, w := range warns {
					t.Logf("warning: line %d: %s", w.Line, w.Message)
				}
				t.Fatalf("unexpected warnings from %s", name)
			}
			out, err := Marshal(root)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			// Re-decode the re-encoded form. Round trip is round-trip
			// stable in structure even if cosmetic whitespace differs.
			root2, _, err := Unmarshal(out)
			if err != nil {
				t.Fatalf("Unmarshal re-encoded: %v\n%s", err, out)
			}
			if !equalObjects(t, root, root2) {
				t.Errorf("round-trip diff for %s", name)
			}
		})
	}
}

func TestPRD60Tasks_FixtureCount(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("testdata", "prd-60-tasks.toon"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	root, _, err := Unmarshal(src)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	tbl := mustTable(t, root, "tasks")
	if len(tbl.Rows) < 60 {
		t.Fatalf("expected ≥60 tasks, got %d", len(tbl.Rows))
	}
}

func mustTable(t *testing.T, root *Object, key string) *TableArray {
	t.Helper()
	v, ok := root.Get(key)
	if !ok {
		t.Fatalf("missing key %q", key)
	}
	tbl, ok := v.(*TableArray)
	if !ok {
		t.Fatalf("%q is %T, want *TableArray", key, v)
	}
	return tbl
}

func equalObjects(t *testing.T, a, b *Object) bool {
	t.Helper()
	if a.Len() != b.Len() {
		return false
	}
	for _, k := range a.Keys() {
		av, _ := a.Get(k)
		bv, _ := b.Get(k)
		if !equalValues(t, av, bv) {
			t.Logf("diff at %q: %v vs %v", k, av, bv)
			return false
		}
	}
	return true
}

func equalValues(t *testing.T, a, b Value) bool {
	switch av := a.(type) {
	case *Object:
		bv, ok := b.(*Object)
		if !ok {
			return false
		}
		return equalObjects(t, av, bv)
	case *TableArray:
		bv, ok := b.(*TableArray)
		if !ok {
			return false
		}
		if strings.Join(av.Columns, ",") != strings.Join(bv.Columns, ",") {
			return false
		}
		if len(av.Rows) != len(bv.Rows) {
			return false
		}
		for i := range av.Rows {
			if (av.Rows[i] == nil) != (bv.Rows[i] == nil) {
				return false
			}
			if av.Rows[i] == nil {
				continue
			}
			for _, c := range av.Columns {
				if !equalValues(t, av.Rows[i][c], bv.Rows[i][c]) {
					return false
				}
			}
		}
		return true
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !equalValues(t, av[i], bv[i]) {
				return false
			}
		}
		return true
	default:
		return numericEqual(a, b) || a == b
	}
}

// numericEqual collapses int64/float64/int variants so a round-tripped
// "1.0" (decoded as float64 then re-encoded as "1") still compares
// equal to its original parsed form.
func numericEqual(a, b Value) bool {
	af, aOk := toFloat(a)
	bf, bOk := toFloat(b)
	if !aOk || !bOk {
		return false
	}
	return af == bf
}

func toFloat(v Value) (float64, bool) {
	switch x := v.(type) {
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case float64:
		return x, true
	default:
		return 0, false
	}
}
