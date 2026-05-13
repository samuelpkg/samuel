package semver

import (
	"fmt"
	"strings"
)

// Range is a parsed version constraint. A range is the AND of one or
// more bounds; "^1.2.3" expands to ">=1.2.3, <2.0.0".
type Range struct {
	// Raw is the original user-supplied string, preserved for error
	// reporting.
	Raw    string
	bounds []bound
	any    bool
}

type bound struct {
	op string // ">=", "<", ">", "<=", "="
	v  Version
}

// ParseRange parses a cargo-style range constraint. Accepted forms:
//
//   - "*"             — matches everything
//   - "1.2.3"         — exact (= 1.2.3)
//   - "=1.2.3"        — exact
//   - "^1.2.3"        — >=1.2.3, <2.0.0 (next major)
//   - "^0.2.3"        — >=0.2.3, <0.3.0 (cargo: 0.x.y locks at minor)
//   - "^0.0.3"        — >=0.0.3, <0.0.4 (cargo: 0.0.x locks at patch)
//   - "~1.2.3"        — >=1.2.3, <1.3.0
//   - ">=1.2.3"       — open upper bound
//   - "<2.0.0"        — open lower bound
//   - ">1.2.3,<2.0.0" — bounded
func ParseRange(s string) (*Range, error) {
	raw := s
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("semver: empty range")
	}
	if s == "*" {
		return &Range{Raw: raw, any: true}, nil
	}
	r := &Range{Raw: raw}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		bs, err := expand(part)
		if err != nil {
			return nil, fmt.Errorf("semver: %s: %w", raw, err)
		}
		r.bounds = append(r.bounds, bs...)
	}
	if len(r.bounds) == 0 {
		return nil, fmt.Errorf("semver: %s yielded no bounds", raw)
	}
	return r, nil
}

func expand(s string) ([]bound, error) {
	switch {
	case strings.HasPrefix(s, "^"):
		v, err := Parse(s[1:])
		if err != nil {
			return nil, err
		}
		// Cargo caret rules: drop trailing zero segments.
		hi := nextCaretUpper(v)
		return []bound{{">=", v}, {"<", hi}}, nil
	case strings.HasPrefix(s, "~"):
		v, err := Parse(s[1:])
		if err != nil {
			return nil, err
		}
		hi := Version{Major: v.Major, Minor: v.Minor + 1, Patch: 0}
		return []bound{{">=", v}, {"<", hi}}, nil
	case strings.HasPrefix(s, ">="):
		v, err := Parse(strings.TrimSpace(s[2:]))
		if err != nil {
			return nil, err
		}
		return []bound{{">=", v}}, nil
	case strings.HasPrefix(s, "<="):
		v, err := Parse(strings.TrimSpace(s[2:]))
		if err != nil {
			return nil, err
		}
		return []bound{{"<=", v}}, nil
	case strings.HasPrefix(s, ">"):
		v, err := Parse(strings.TrimSpace(s[1:]))
		if err != nil {
			return nil, err
		}
		return []bound{{">", v}}, nil
	case strings.HasPrefix(s, "<"):
		v, err := Parse(strings.TrimSpace(s[1:]))
		if err != nil {
			return nil, err
		}
		return []bound{{"<", v}}, nil
	case strings.HasPrefix(s, "="):
		v, err := Parse(strings.TrimSpace(s[1:]))
		if err != nil {
			return nil, err
		}
		return []bound{{"=", v}}, nil
	default:
		// bare X.Y.Z => exact match.
		v, err := Parse(s)
		if err != nil {
			return nil, err
		}
		return []bound{{"=", v}}, nil
	}
}

// nextCaretUpper implements Cargo's caret upper-bound rules.
//
//	^1.2.3  -> <2.0.0
//	^0.2.3  -> <0.3.0
//	^0.0.3  -> <0.0.4
//	^0.0.0  -> <0.0.1
func nextCaretUpper(v Version) Version {
	if v.Major > 0 {
		return Version{Major: v.Major + 1}
	}
	if v.Minor > 0 {
		return Version{Major: 0, Minor: v.Minor + 1}
	}
	return Version{Major: 0, Minor: 0, Patch: v.Patch + 1}
}

// Matches reports whether v satisfies the range. All bounds are AND-ed.
func (r *Range) Matches(v Version) bool {
	if r == nil {
		return false
	}
	if r.any {
		return true
	}
	for _, b := range r.bounds {
		c := v.Compare(b.v)
		switch b.op {
		case ">=":
			if c < 0 {
				return false
			}
		case ">":
			if c <= 0 {
				return false
			}
		case "<=":
			if c > 0 {
				return false
			}
		case "<":
			if c >= 0 {
				return false
			}
		case "=":
			if c != 0 {
				return false
			}
		}
	}
	return true
}

// ResolveOptions controls version selection.
type ResolveOptions struct {
	// AllowPrerelease lets the resolver consider versions with a
	// prerelease tag. Default off (Cargo behaviour).
	AllowPrerelease bool
}

// Resolve picks the highest version from available that satisfies r.
// Returns ErrNoMatch when no candidate fits.
func (r *Range) Resolve(available []Version, opts ResolveOptions) (Version, error) {
	var best Version
	var found bool
	for _, v := range available {
		if v.IsPrerelease() && !opts.AllowPrerelease {
			continue
		}
		if !r.Matches(v) {
			continue
		}
		if !found || v.Compare(best) > 0 {
			best = v
			found = true
		}
	}
	if !found {
		return Version{}, &NoMatchError{Range: r.Raw}
	}
	return best, nil
}

// NoMatchError is returned by Resolve when no candidate version fits.
type NoMatchError struct{ Range string }

func (e *NoMatchError) Error() string {
	return fmt.Sprintf("semver: no version satisfies %q", e.Range)
}
