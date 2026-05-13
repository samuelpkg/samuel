// Package toon implements a Go encoder/decoder for the subset of the
// TOON v3 specification that Samuel v2 actually uses for its runtime
// files in .samuel/run/ (task state, project snapshots, task context).
//
// # Why we ship our own
//
// As of May 2026 the TOON reference implementation is TypeScript; no
// maintained Go library exists. The spec is small (~28 formatting
// rules in v3) so writing our own keeps Samuel free of external
// runtime dependencies and lets us pin the spec version we read and
// write. See .wiki/concepts/toon-evaluation.md for the full rationale.
//
// # Supported subset
//
// The encoder/decoder handles:
//
//   - Primitive scalars: string, bool, int64, float64, nil
//   - Nested objects (indentation-based, two-space indent)
//   - Scalar fields:        key: value
//   - Tabular arrays:       items[N]{c1,c2}:  followed by N CSV rows
//   - Comments:             lines starting with "#" are skipped on read
//   - Version header:       "# toon v3" expected on the first line
//
// The subset is deliberately tight. Heterogeneous arrays, deeply
// recursive types, and the spec's binary/hex literal forms are not
// supported and intentionally rejected.
//
// # Version policy
//
// Every Samuel-written .toon file MUST start with "# toon v3" (see
// VersionHeader). Decode rejects unknown major versions; minor-version
// drift is tolerated and recorded as a Warning.
//
// # Failure mode
//
// Decode is line-oriented and tolerant. A malformed row in a tabular
// array is skipped, a structured Warning is recorded on the Decoder,
// and parsing continues. This matches the v1 prd.json custom-Unmarshal
// pattern that lets the autonomous loop keep running through
// transient agent-emitted malformations.
package toon
