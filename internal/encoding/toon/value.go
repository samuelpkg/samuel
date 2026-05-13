package toon

import (
	"fmt"
	"sort"
)

// Value is the in-memory representation of a TOON document. It is the
// pivot type that encoder and decoder both produce/consume so callers
// can use either side independently.
//
// A Value is one of:
//   - nil                                  (TOON null)
//   - bool
//   - int64
//   - float64
//   - string
//   - *Object       (ordered map)
//   - *TableArray   (tabular array of homogeneous records)
//   - []any         (scalar array — primitive elements only)
//
// The encoder accepts any of these as the root. The decoder always
// produces an *Object as the root.
type Value = any

// Object is an ordered map[string]Value. TOON files preserve key
// order on write (cosmetic), and the decoder preserves source order
// so round-trips are deterministic.
type Object struct {
	keys   []string
	values map[string]Value
}

// NewObject returns an empty ordered Object.
func NewObject() *Object {
	return &Object{values: map[string]Value{}}
}

// Set inserts or updates key. Insertion order is preserved; updating
// an existing key keeps its original position.
func (o *Object) Set(key string, v Value) {
	if _, ok := o.values[key]; !ok {
		o.keys = append(o.keys, key)
	}
	o.values[key] = v
}

// Get returns the value at key and whether it was present.
func (o *Object) Get(key string) (Value, bool) {
	v, ok := o.values[key]
	return v, ok
}

// Keys returns the keys in insertion order. The slice is a copy; the
// caller may mutate it freely.
func (o *Object) Keys() []string {
	out := make([]string, len(o.keys))
	copy(out, o.keys)
	return out
}

// Len reports the number of entries.
func (o *Object) Len() int { return len(o.keys) }

// SortKeys re-orders keys lexicographically. Useful for golden-file
// stability when the source data is a Go map (random order).
func (o *Object) SortKeys() {
	sort.Strings(o.keys)
}

// TableArray is a tabular array of homogeneous records — TOON's flagship
// shape. Each record is a map keyed by the declared columns. The
// encoder writes:
//
//	field[N]{c1,c2}:
//	  v1,v2
//	  v3,v4
//
// Rows may be nil to record a malformed/skipped entry; encoders skip
// nil rows, decoders insert nil for rows that failed to parse.
type TableArray struct {
	Columns []string
	Rows    []map[string]Value
}

// NewTableArray builds an empty table with the given columns.
func NewTableArray(columns ...string) *TableArray {
	cols := make([]string, len(columns))
	copy(cols, columns)
	return &TableArray{Columns: cols}
}

// AddRow appends a row. Keys present in the row that are not declared
// columns are silently dropped on encode (decoder errors instead).
func (t *TableArray) AddRow(row map[string]Value) {
	cp := make(map[string]Value, len(t.Columns))
	for _, c := range t.Columns {
		if v, ok := row[c]; ok {
			cp[c] = v
		} else {
			cp[c] = nil
		}
	}
	t.Rows = append(t.Rows, cp)
}

// validate checks the table is internally consistent. Used by encoder
// to short-circuit obvious mistakes (duplicate columns, etc).
func (t *TableArray) validate() error {
	seen := map[string]struct{}{}
	for _, c := range t.Columns {
		if c == "" {
			return fmt.Errorf("table column name cannot be empty")
		}
		if _, dup := seen[c]; dup {
			return fmt.Errorf("duplicate table column %q", c)
		}
		seen[c] = struct{}{}
	}
	return nil
}
