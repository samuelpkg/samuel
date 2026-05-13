package ui

import (
	"encoding/json"
	"fmt"
	"os"
)

// JSONSchemaVersion identifies the JSON envelope shape.
//
// v4 changes vs v3 (v1's envelope):
//   - schemaVersion bumped to 4 to mark the v2 framework boundary.
//   - "command" still reflects the invoked command path with the
//     "samuel " prefix stripped (legacy aliases keep their alias name).
//   - "warnings" added as an optional field carrying TOON decoder
//     warnings and other non-fatal events. Consumers that pin schema
//     3 may treat this as forward-compatible noise.
const JSONSchemaVersion = 4

// JSONResponse is the standard envelope for --json output.
type JSONResponse struct {
	SchemaVersion int      `json:"schemaVersion"`
	Command       string   `json:"command"`
	Success       bool     `json:"success"`
	Data          any      `json:"data,omitempty"`
	Error         string   `json:"error,omitempty"`
	Warnings      []string `json:"warnings,omitempty"`
}

// PrintJSON writes a successful JSON response to stdout.
func PrintJSON(command string, data any) {
	resp := JSONResponse{
		SchemaVersion: JSONSchemaVersion,
		Command:       command,
		Success:       true,
		Data:          data,
	}
	writeJSON(stdout, resp)
}

// PrintJSONWithWarnings writes a successful response that carries
// non-fatal warnings (e.g. TOON minor-version drift).
func PrintJSONWithWarnings(command string, data any, warnings []string) {
	resp := JSONResponse{
		SchemaVersion: JSONSchemaVersion,
		Command:       command,
		Success:       true,
		Data:          data,
		Warnings:      warnings,
	}
	writeJSON(stdout, resp)
}

// PrintJSONError writes an error JSON response to stderr. The exit
// status itself is set by main.go based on the returned Go error.
func PrintJSONError(command string, err error) {
	resp := JSONResponse{
		SchemaVersion: JSONSchemaVersion,
		Command:       command,
		Success:       false,
		Error:         err.Error(),
	}
	out, _ := json.Marshal(resp)
	fmt.Fprintln(os.Stderr, string(out))
}

func writeJSON(w any, v any) {
	out, ok := w.(interface{ Write([]byte) (int, error) })
	if !ok {
		out = os.Stdout
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
