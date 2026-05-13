package toon

import (
	"strings"
	"testing"
)

func TestMarshal_VersionHeaderFirst(t *testing.T) {
	root := NewObject()
	root.Set("foo", "bar")
	out, err := Marshal(root)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.HasPrefix(string(out), VersionHeader+"\n") {
		t.Errorf("output missing version header: %q", out)
	}
}

func TestMarshal_RejectsNonObjectRoot(t *testing.T) {
	if _, err := Marshal("not an object"); err == nil {
		t.Errorf("expected error for non-object root")
	}
}

func TestMarshal_Scalars(t *testing.T) {
	root := NewObject()
	root.Set("s", "hello")
	root.Set("n", int64(42))
	root.Set("f", 3.14)
	root.Set("b", true)
	root.Set("z", nil)
	out, err := Marshal(root)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []string{
		"s: hello",
		"n: 42",
		"f: 3.14",
		"b: true",
		"z: null",
	}
	got := string(out)
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("output missing line %q in:\n%s", w, got)
		}
	}
}

func TestMarshal_QuotesWhenAmbiguous(t *testing.T) {
	cases := map[string]string{
		"true":          `"true"`,
		"123":           `"123"`,
		"":              `""`,
		"has: colon":    `"has: colon"`,
		" leading":      `" leading"`,
		"trailing ":     `"trailing "`,
		"line\nbreak":   `"line\nbreak"`,
		"quote\"inside": `"quote\"inside"`,
	}
	for raw, want := range cases {
		root := NewObject()
		root.Set("k", raw)
		out, err := Marshal(root)
		if err != nil {
			t.Fatalf("Marshal(%q): %v", raw, err)
		}
		if !strings.Contains(string(out), "k: "+want+"\n") {
			t.Errorf("Marshal(%q) did not quote as expected; want suffix %q in:\n%s", raw, want, out)
		}
	}
}

func TestMarshal_NestedObject(t *testing.T) {
	inner := NewObject()
	inner.Set("id", int64(1))
	inner.Set("title", "Add auth")
	root := NewObject()
	root.Set("task", inner)
	out, err := Marshal(root)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got := string(out)
	if !strings.Contains(got, "task:\n  id: 1\n  title: Add auth\n") {
		t.Errorf("unexpected nested output:\n%s", got)
	}
}

func TestMarshal_EmptyObject(t *testing.T) {
	root := NewObject()
	root.Set("empty", NewObject())
	out, _ := Marshal(root)
	if !strings.Contains(string(out), "empty: {}\n") {
		t.Errorf("empty object should render as {}; got:\n%s", out)
	}
}

func TestMarshal_TabularArray(t *testing.T) {
	tbl := NewTableArray("id", "title", "status")
	tbl.AddRow(map[string]Value{"id": int64(1), "title": "Add auth", "status": "done"})
	tbl.AddRow(map[string]Value{"id": int64(2), "title": "tests, more", "status": "todo"})
	root := NewObject()
	root.Set("tasks", tbl)
	out, err := Marshal(root)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got := string(out)
	if !strings.Contains(got, "tasks[2]{id,title,status}:\n") {
		t.Errorf("missing header:\n%s", got)
	}
	if !strings.Contains(got, `  2,"tests, more",todo`) {
		t.Errorf("expected quoted cell with comma:\n%s", got)
	}
}

func TestMarshal_ScalarArray(t *testing.T) {
	root := NewObject()
	root.Set("xs", []any{int64(1), int64(2), int64(3)})
	out, _ := Marshal(root)
	if !strings.Contains(string(out), "xs[3]: 1,2,3\n") {
		t.Errorf("scalar array output unexpected:\n%s", out)
	}
}

func TestMarshal_ScalarArrayRejectsComposite(t *testing.T) {
	root := NewObject()
	root.Set("bad", []any{NewObject()})
	if _, err := Marshal(root); err == nil {
		t.Errorf("expected error for composite-in-scalar-array")
	}
}

func TestTableArray_ValidateDuplicates(t *testing.T) {
	tbl := NewTableArray("a", "a")
	if err := tbl.validate(); err == nil {
		t.Errorf("expected duplicate-column error")
	}
}
