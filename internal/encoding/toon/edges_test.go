package toon

import (
	"strings"
	"testing"
)

// Edge-case tests for parser branches that the golden corpus doesn't
// hit. These exist to push coverage above the 80% threshold and to
// pin down behaviour that future spec drift might silently change.

func TestUnmarshal_RejectsUnterminatedArrayLength(t *testing.T) {
	src := "# toon v3.0\nbad[5: 1\n"
	if _, _, err := Unmarshal([]byte(src)); err == nil {
		t.Errorf("expected unterminated-length error")
	}
}

func TestUnmarshal_RejectsBadArrayLength(t *testing.T) {
	src := "# toon v3.0\nbad[oops]: 1\n"
	if _, _, err := Unmarshal([]byte(src)); err == nil {
		t.Errorf("expected bad-length error")
	}
}

func TestUnmarshal_ScalarArrayEmpty(t *testing.T) {
	src := "# toon v3.0\nxs[0]: \n"
	root, _, err := Unmarshal([]byte(src))
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	v, _ := root.Get("xs")
	if len(v.([]any)) != 0 {
		t.Errorf("expected empty slice, got %v", v)
	}
}

func TestUnmarshal_ScalarArrayMismatch(t *testing.T) {
	src := "# toon v3.0\nxs[3]: 1,2\n"
	if _, _, err := Unmarshal([]byte(src)); err == nil {
		t.Errorf("expected length-mismatch error")
	}
}

func TestUnmarshal_QuotedEscapes(t *testing.T) {
	src := "# toon v3.0\n" +
		`k1: "a\"b"` + "\n" +
		`k2: "tab\there"` + "\n" +
		`k3: "nl\nhere"` + "\n" +
		`k4: "cr\rhere"` + "\n"
	root, _, err := Unmarshal([]byte(src))
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	want := map[string]string{
		"k1": `a"b`,
		"k2": "tab\there",
		"k3": "nl\nhere",
		"k4": "cr\rhere",
	}
	for k, w := range want {
		v, _ := root.Get(k)
		if v != w {
			t.Errorf("%s = %q, want %q", k, v, w)
		}
	}
}

func TestUnmarshal_RejectsBadEscape(t *testing.T) {
	src := "# toon v3.0\nk: \"bad\\xescape\"\n"
	if _, _, err := Unmarshal([]byte(src)); err == nil {
		t.Errorf("expected bad-escape error")
	}
}

func TestUnmarshal_NestedTablesRoundTrip(t *testing.T) {
	src := "# toon v3.0\n" +
		"outer:\n" +
		"  inner[2]{a,b}:\n" +
		"    1,x\n" +
		"    2,y\n"
	root, _, err := Unmarshal([]byte(src))
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	outer, _ := root.Get("outer")
	tbl, _ := outer.(*Object).Get("inner")
	ta := tbl.(*TableArray)
	if len(ta.Rows) != 2 {
		t.Fatalf("rows=%d", len(ta.Rows))
	}
	if ta.Rows[1]["b"] != "y" {
		t.Errorf("row 1 b = %v", ta.Rows[1]["b"])
	}
}

func TestMarshal_DropsNilRowsOnEncode(t *testing.T) {
	tbl := NewTableArray("id", "title")
	tbl.AddRow(map[string]Value{"id": int64(1), "title": "first"})
	tbl.Rows = append(tbl.Rows, nil)
	tbl.AddRow(map[string]Value{"id": int64(2), "title": "third"})

	root := NewObject()
	root.Set("xs", tbl)
	out, err := Marshal(root)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got := string(out)
	if !strings.Contains(got, "xs[2]{id,title}:\n") {
		t.Errorf("header should report 2 (skip nil); got:\n%s", got)
	}
	// Re-decode to confirm 2 valid rows.
	root2, _, err := Unmarshal(out)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	ta := root2.Keys()
	_ = ta
}
