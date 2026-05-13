package toon

import (
	"fmt"
	"strconv"
	"strings"
)

// SpecMajor is the TOON spec major version Samuel v2 pins to.
// Reading a file authored against a different major is a hard error.
const SpecMajor = 3

// SpecVersion is the full pinned spec version string ("3.0").
const SpecVersion = "3.0"

// VersionHeader is the literal first-line header every Samuel-written
// .toon file must carry. The parser checks this on Decode.
const VersionHeader = "# toon v" + SpecVersion

// parseVersionHeader extracts the (major, minor) version from a line
// of the form "# toon vMAJOR[.MINOR]". Returns ok=false for anything
// that does not look like a TOON version header.
func parseVersionHeader(line string) (major, minor int, ok bool) {
	const prefix = "# toon v"
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, prefix) {
		return 0, 0, false
	}
	rest := strings.TrimSpace(trimmed[len(prefix):])
	if rest == "" {
		return 0, 0, false
	}
	parts := strings.SplitN(rest, ".", 2)
	maj, err := strconv.Atoi(parts[0])
	if err != nil || maj < 0 {
		return 0, 0, false
	}
	if len(parts) == 1 {
		return maj, 0, true
	}
	mn, err := strconv.Atoi(parts[1])
	if err != nil || mn < 0 {
		return 0, 0, false
	}
	return maj, mn, true
}

// checkVersionLine validates the first non-empty line of a TOON file
// carries a supported version header. Returns a string describing the
// observed version (for warnings) and an error when the major is
// incompatible.
func checkVersionLine(line string) (observed string, err error) {
	maj, mn, ok := parseVersionHeader(line)
	if !ok {
		return "", fmt.Errorf("missing or malformed TOON version header (want %q on first line)", VersionHeader)
	}
	observed = fmt.Sprintf("%d.%d", maj, mn)
	if maj != SpecMajor {
		return observed, fmt.Errorf("incompatible TOON spec major: file=%d, supported=%d", maj, SpecMajor)
	}
	return observed, nil
}
