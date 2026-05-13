package toon

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

const indentUnit = "  "

// Marshal encodes v as a TOON document. The root MUST be an *Object;
// other types are rejected to keep the on-disk shape predictable for
// Samuel's runtime files (every .samuel/run/*.toon file is a struct).
//
// Output always begins with the pinned version header (VersionHeader).
func Marshal(v Value) ([]byte, error) {
	root, ok := v.(*Object)
	if !ok {
		return nil, fmt.Errorf("toon: Marshal expects *Object at root, got %T", v)
	}
	var buf bytes.Buffer
	buf.WriteString(VersionHeader)
	buf.WriteByte('\n')
	if err := writeObject(&buf, root, 0); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MarshalIndent is provided for API parity; TOON uses fixed two-space
// indentation, so the function ignores extra options and simply
// delegates to Marshal.
func MarshalIndent(v Value, _, _ string) ([]byte, error) { return Marshal(v) }

func writeObject(w io.Writer, obj *Object, depth int) error {
	for _, k := range obj.keys {
		if err := writeKeyValue(w, k, obj.values[k], depth); err != nil {
			return err
		}
	}
	return nil
}

func writeKeyValue(w io.Writer, key string, v Value, depth int) error {
	pad := strings.Repeat(indentUnit, depth)
	switch val := v.(type) {
	case *Object:
		// Empty object renders as "key: {}" to round-trip cleanly.
		if val.Len() == 0 {
			fmt.Fprintf(w, "%s%s: {}\n", pad, key)
			return nil
		}
		fmt.Fprintf(w, "%s%s:\n", pad, key)
		return writeObject(w, val, depth+1)
	case *TableArray:
		return writeTableArray(w, key, val, depth)
	case []any:
		return writeScalarArray(w, key, val, depth)
	default:
		fmt.Fprintf(w, "%s%s: %s\n", pad, key, encodeScalar(v))
		return nil
	}
}

func writeTableArray(w io.Writer, key string, t *TableArray, depth int) error {
	if err := t.validate(); err != nil {
		return err
	}
	pad := strings.Repeat(indentUnit, depth)
	cols := strings.Join(t.Columns, ",")
	// Skip nil rows on encode; they signal a previous decode marked the
	// entry as malformed. The header N counts only emitted rows so the
	// file always parses back consistently.
	emitted := 0
	for _, r := range t.Rows {
		if r != nil {
			emitted++
		}
	}
	fmt.Fprintf(w, "%s%s[%d]{%s}:\n", pad, key, emitted, cols)
	rowPad := strings.Repeat(indentUnit, depth+1)
	for _, r := range t.Rows {
		if r == nil {
			continue
		}
		cells := make([]string, 0, len(t.Columns))
		for _, c := range t.Columns {
			cells = append(cells, encodeCell(r[c]))
		}
		fmt.Fprintf(w, "%s%s\n", rowPad, strings.Join(cells, ","))
	}
	return nil
}

func writeScalarArray(w io.Writer, key string, arr []any, depth int) error {
	pad := strings.Repeat(indentUnit, depth)
	if len(arr) == 0 {
		fmt.Fprintf(w, "%s%s: []\n", pad, key)
		return nil
	}
	cells := make([]string, 0, len(arr))
	for _, v := range arr {
		switch v.(type) {
		case *Object, *TableArray, []any:
			return fmt.Errorf("toon: scalar array %q may not contain composite values", key)
		}
		cells = append(cells, encodeCell(v))
	}
	fmt.Fprintf(w, "%s%s[%d]: %s\n", pad, key, len(arr), strings.Join(cells, ","))
	return nil
}

// encodeScalar renders a scalar value for the "key: value" form. The
// rules:
//
//   - nil      -> "null"
//   - bool     -> "true" / "false"
//   - int64    -> decimal
//   - float64  -> %g (compact; Go's default)
//   - string   -> quoted only when needed (whitespace, comma, colon,
//     bracket, or starts with a TOON literal that could
//     be misread).
func encodeScalar(v Value) string {
	switch val := v.(type) {
	case nil:
		return "null"
	case bool:
		if val {
			return "true"
		}
		return "false"
	case int:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'g', -1, 64)
	case string:
		return encodeString(val, false)
	default:
		// Conservative fallback: %v stringified, then quoted.
		return encodeString(fmt.Sprintf("%v", val), false)
	}
}

// encodeCell renders a value inside a tabular row. Cells are stricter:
// commas always force quoting, and embedded newlines are rejected
// (rows are one line each).
func encodeCell(v Value) string {
	switch val := v.(type) {
	case nil:
		return "null"
	case bool:
		if val {
			return "true"
		}
		return "false"
	case int:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'g', -1, 64)
	case string:
		return encodeString(val, true)
	default:
		return encodeString(fmt.Sprintf("%v", val), true)
	}
}

// encodeString quotes a string only when necessary. inTable=true
// applies the stricter cell rules.
func encodeString(s string, inTable bool) string {
	if s == "" {
		return `""`
	}
	needsQuote := false
	for _, r := range s {
		switch r {
		case '"', '\\', '\n', '\r', '\t':
			needsQuote = true
		case ',':
			if inTable {
				needsQuote = true
			}
		case ':':
			needsQuote = true
		case ' ':
			// Leading/trailing spaces require quoting; we check below.
		}
		if needsQuote {
			break
		}
	}
	if !needsQuote {
		if s[0] == ' ' || s[len(s)-1] == ' ' {
			needsQuote = true
		}
	}
	if !needsQuote && wouldShadowLiteral(s) {
		needsQuote = true
	}
	if !needsQuote {
		return s
	}
	// JSON-style escape — the v3 spec defers to JSON string escapes.
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// wouldShadowLiteral returns true if s, written bare, would be parsed
// back as a different type (bool, null, number).
func wouldShadowLiteral(s string) bool {
	switch s {
	case "true", "false", "null":
		return true
	}
	if _, err := strconv.ParseInt(s, 10, 64); err == nil {
		return true
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}
	return false
}

// MapToObject converts a plain map[string]any into a TOON *Object with
// keys sorted lexicographically. Useful for callers that source data
// from JSON or YAML where key order is undefined.
func MapToObject(m map[string]any) *Object {
	obj := NewObject()
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		obj.Set(k, m[k])
	}
	return obj
}
