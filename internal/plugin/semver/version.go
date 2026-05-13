// Package semver implements the subset of SemVer 2.0 + Cargo range
// syntax that Samuel v2's plugin loader needs:
//
//   - Versions: MAJOR.MINOR.PATCH with optional `-pre` prerelease suffix
//     and optional `+build` metadata. Build metadata is preserved on
//     parse but ignored for ordering, per SemVer §10.
//   - Ranges: ^X.Y.Z, ~X.Y.Z, >=X.Y.Z, <X.Y.Z, =X.Y.Z, X.Y.Z (exact),
//     "*" (any), and comma-separated bounded ranges: ">=1.0.0,<2.0.0".
//   - Resolution: pick the highest version from a list that satisfies a
//     range. Prereleases are excluded by default; the caller can opt in
//     via ResolveOptions.AllowPrerelease.
//
// This is intentionally a slim, hand-rolled implementation rather than a
// dependency on golang.org/x/mod/semver — that package targets Go module
// versions ("v1.2.3"), Samuel uses bare SemVer ("1.2.3"), and the cargo
// range vocabulary is outside its scope. Behaviour is matched against
// Cargo's reference fixtures (TestCargoFixtures).
package semver

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Version is a parsed SemVer 2.0 value.
type Version struct {
	Major, Minor, Patch int
	// Prerelease is the dot-separated identifier list after "-" (e.g.
	// "rc.1" → ["rc","1"]). Empty for a stable release.
	Prerelease []string
	// Build is the dot-separated identifier list after "+". Preserved on
	// parse, ignored for ordering.
	Build []string
}

// Parse parses a SemVer string. Leading "v" is tolerated.
func Parse(s string) (Version, error) {
	orig := s
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	var v Version

	// Split build metadata first.
	if idx := strings.Index(s, "+"); idx >= 0 {
		v.Build = strings.Split(s[idx+1:], ".")
		s = s[:idx]
	}
	// Then prerelease.
	if idx := strings.Index(s, "-"); idx >= 0 {
		v.Prerelease = strings.Split(s[idx+1:], ".")
		s = s[:idx]
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("semver: %q is not MAJOR.MINOR.PATCH", orig)
	}
	var err error
	if v.Major, err = strconv.Atoi(parts[0]); err != nil || v.Major < 0 {
		return Version{}, fmt.Errorf("semver: bad MAJOR in %q", orig)
	}
	if v.Minor, err = strconv.Atoi(parts[1]); err != nil || v.Minor < 0 {
		return Version{}, fmt.Errorf("semver: bad MINOR in %q", orig)
	}
	if v.Patch, err = strconv.Atoi(parts[2]); err != nil || v.Patch < 0 {
		return Version{}, fmt.Errorf("semver: bad PATCH in %q", orig)
	}
	return v, nil
}

// MustParse panics on error; use only for compile-time-known constants.
func MustParse(s string) Version {
	v, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return v
}

// String renders the canonical form (omits build metadata; SemVer §10
// allows that for equality/ordering purposes).
func (v Version) String() string {
	out := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if len(v.Prerelease) > 0 {
		out += "-" + strings.Join(v.Prerelease, ".")
	}
	return out
}

// IsPrerelease reports whether v carries a prerelease tag.
func (v Version) IsPrerelease() bool { return len(v.Prerelease) > 0 }

// Compare returns -1, 0, +1 per the SemVer §11 precedence rules.
// Build metadata is ignored.
func (a Version) Compare(b Version) int {
	if c := cmpInt(a.Major, b.Major); c != 0 {
		return c
	}
	if c := cmpInt(a.Minor, b.Minor); c != 0 {
		return c
	}
	if c := cmpInt(a.Patch, b.Patch); c != 0 {
		return c
	}
	// Prerelease handling: absence > presence.
	switch {
	case len(a.Prerelease) == 0 && len(b.Prerelease) == 0:
		return 0
	case len(a.Prerelease) == 0:
		return 1
	case len(b.Prerelease) == 0:
		return -1
	}
	for i := 0; i < len(a.Prerelease) && i < len(b.Prerelease); i++ {
		if c := cmpIdent(a.Prerelease[i], b.Prerelease[i]); c != 0 {
			return c
		}
	}
	return cmpInt(len(a.Prerelease), len(b.Prerelease))
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// cmpIdent compares two prerelease identifiers. Numeric identifiers
// compare numerically; non-numeric compare lexically; numeric < non-
// numeric (SemVer §11.4.3).
func cmpIdent(a, b string) int {
	an, aerr := strconv.Atoi(a)
	bn, berr := strconv.Atoi(b)
	switch {
	case aerr == nil && berr == nil:
		return cmpInt(an, bn)
	case aerr == nil:
		return -1
	case berr == nil:
		return 1
	default:
		return strings.Compare(a, b)
	}
}

// Sort orders vs ascending. The slice is sorted in place.
func Sort(vs []Version) {
	sort.Slice(vs, func(i, j int) bool { return vs[i].Compare(vs[j]) < 0 })
}
