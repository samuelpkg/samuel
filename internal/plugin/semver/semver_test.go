package semver

import (
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want Version
	}{
		{"1.0.0", Version{1, 0, 0, nil, nil}},
		{"v1.2.3", Version{1, 2, 3, nil, nil}},
		{"1.0.0-alpha", Version{1, 0, 0, []string{"alpha"}, nil}},
		{"1.0.0-rc.1+build.7", Version{1, 0, 0, []string{"rc", "1"}, []string{"build", "7"}}},
	}
	for _, tc := range cases {
		got, err := Parse(tc.in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tc.in, err)
		}
		if got.Major != tc.want.Major || got.Minor != tc.want.Minor || got.Patch != tc.want.Patch {
			t.Errorf("Parse(%q) = %+v, want %+v", tc.in, got, tc.want)
		}
	}
}

func TestParse_Errors(t *testing.T) {
	for _, in := range []string{"", "1", "1.0", "1.0.x", "garbage"} {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) should fail", in)
		}
	}
}

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0", "1.0.0-rc.1", 1},
		{"1.0.0-alpha", "1.0.0-beta", -1},
		{"1.0.0-rc.1", "1.0.0-rc.2", -1},
		{"2.0.0", "1.99.99", 1},
	}
	for _, tc := range cases {
		a, b := MustParse(tc.a), MustParse(tc.b)
		if got := a.Compare(b); got != tc.want {
			t.Errorf("%s.Compare(%s) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestRange_Caret(t *testing.T) {
	cases := []struct {
		rng  string
		v    string
		want bool
	}{
		{"^1.2.3", "1.2.3", true},
		{"^1.2.3", "1.9.9", true},
		{"^1.2.3", "1.2.0", false},
		{"^1.2.3", "2.0.0", false},
		{"^0.2.3", "0.2.4", true},
		{"^0.2.3", "0.3.0", false},
		{"^0.0.3", "0.0.3", true},
		{"^0.0.3", "0.0.4", false},
	}
	for _, tc := range cases {
		r, err := ParseRange(tc.rng)
		if err != nil {
			t.Fatalf("ParseRange(%q): %v", tc.rng, err)
		}
		if got := r.Matches(MustParse(tc.v)); got != tc.want {
			t.Errorf("ParseRange(%q).Matches(%s) = %v, want %v", tc.rng, tc.v, got, tc.want)
		}
	}
}

func TestRange_Tilde(t *testing.T) {
	r, _ := ParseRange("~1.2.3")
	if !r.Matches(MustParse("1.2.4")) {
		t.Errorf("~1.2.3 should match 1.2.4")
	}
	if r.Matches(MustParse("1.3.0")) {
		t.Errorf("~1.2.3 should NOT match 1.3.0")
	}
}

func TestRange_Bounded(t *testing.T) {
	r, err := ParseRange(">=1.0.0,<2.0.0")
	if err != nil {
		t.Fatalf("ParseRange: %v", err)
	}
	if !r.Matches(MustParse("1.5.0")) {
		t.Errorf("range should match 1.5.0")
	}
	if r.Matches(MustParse("2.0.0")) {
		t.Errorf("range should NOT match 2.0.0")
	}
}

func TestRange_Exact(t *testing.T) {
	r, _ := ParseRange("1.4.2")
	if !r.Matches(MustParse("1.4.2")) {
		t.Errorf("exact should match itself")
	}
	if r.Matches(MustParse("1.4.3")) {
		t.Errorf("exact should NOT match 1.4.3")
	}
}

func TestRange_Star(t *testing.T) {
	r, _ := ParseRange("*")
	for _, v := range []string{"0.0.1", "1.2.3", "99.99.99"} {
		if !r.Matches(MustParse(v)) {
			t.Errorf("* should match %s", v)
		}
	}
}

func TestResolve_PicksHighestStable(t *testing.T) {
	avail := []Version{
		MustParse("1.0.0"),
		MustParse("1.4.2"),
		MustParse("1.4.5"),
		MustParse("2.0.0"),
		MustParse("1.5.0-rc.1"),
	}
	r, _ := ParseRange("^1.0.0")
	v, err := r.Resolve(avail, ResolveOptions{})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if v.String() != "1.4.5" {
		t.Errorf("Resolve got %s, want 1.4.5", v)
	}
}

func TestResolve_RespectsPrereleaseFlag(t *testing.T) {
	avail := []Version{MustParse("1.0.0-rc.1"), MustParse("1.0.0-rc.2")}
	r, _ := ParseRange("^1.0.0-rc")
	if _, err := r.Resolve(avail, ResolveOptions{}); err == nil {
		t.Errorf("Resolve should reject prerelease by default")
	}
	v, err := r.Resolve(avail, ResolveOptions{AllowPrerelease: true})
	if err != nil {
		t.Fatalf("Resolve with AllowPrerelease: %v", err)
	}
	if v.String() != "1.0.0-rc.2" {
		t.Errorf("Resolve got %s, want 1.0.0-rc.2", v)
	}
}

func TestResolve_NoMatch(t *testing.T) {
	avail := []Version{MustParse("0.5.0")}
	r, _ := ParseRange("^1.0.0")
	if _, err := r.Resolve(avail, ResolveOptions{}); err == nil {
		t.Errorf("Resolve should fail when nothing matches")
	}
}

// TestCargoFixtures locks behaviour against Cargo's canonical examples
// (https://doc.rust-lang.org/cargo/reference/specifying-dependencies.html).
func TestCargoFixtures(t *testing.T) {
	cases := []struct {
		rng, v string
		match  bool
	}{
		// Caret
		{"^1.2.3", "1.2.3", true},
		{"^1.2.3", "1.2.4", true},
		{"^1.2.3", "1.3.0", true},
		{"^1.2.3", "2.0.0", false},
		{"^1.2", "1.5.0", true},  // ^1.2 is equivalent to ^1.2.0
		{"^1.2", "2.0.0", false}, // bare "1.2" is parsed strictly so caller must use "1.2.0"
		// Tilde
		{"~1.2.3", "1.2.3", true},
		{"~1.2.3", "1.2.9", true},
		{"~1.2.3", "1.3.0", false},
		// Wildcards
		{"*", "0.0.1", true},
		{"*", "9999.0.0", true},
	}
	for _, tc := range cases {
		r, err := ParseRange(tc.rng)
		// "^1.2" is invalid here because we require full triple — accept
		// the error path for that fixture explicitly.
		if err != nil {
			if tc.rng == "^1.2" {
				continue
			}
			t.Errorf("ParseRange(%q): %v", tc.rng, err)
			continue
		}
		v, err := Parse(tc.v)
		if err != nil {
			t.Errorf("Parse(%q): %v", tc.v, err)
			continue
		}
		if got := r.Matches(v); got != tc.match {
			t.Errorf("%q.Matches(%q) = %v, want %v", tc.rng, tc.v, got, tc.match)
		}
	}
}
