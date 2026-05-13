package commands

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/samuelpkg/samuel/internal/ui"
)

func executeRoot(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	ResetFlagsForTest()
	out, errw := &bytes.Buffer{}, &bytes.Buffer{}
	ui.SetWriters(out, errw)
	t.Cleanup(func() { ui.SetWriters(nil, nil) })

	rootCmd.SetOut(out)
	rootCmd.SetErr(errw)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return out.String(), errw.String(), err
}

func TestVersion_HumanReadable(t *testing.T) {
	out, _, err := executeRoot(t, "version")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{"Samuel CLI", "Version:", "Commit:", "Built:"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q in:\n%s", want, out)
		}
	}
}

func TestVersion_JSONEnvelope(t *testing.T) {
	out, _, err := executeRoot(t, "version", "--json")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	var resp ui.JSONResponse
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	if resp.SchemaVersion != ui.JSONSchemaVersion {
		t.Errorf("schemaVersion = %d, want %d", resp.SchemaVersion, ui.JSONSchemaVersion)
	}
	if resp.Command != "version" {
		t.Errorf("command = %q", resp.Command)
	}
	if !resp.Success {
		t.Errorf("success = false")
	}
	data, _ := resp.Data.(map[string]any)
	if data["version"] == "" || data["version"] == nil {
		t.Errorf("data.version missing: %v", resp.Data)
	}
}

func TestJSONMode_NilSafe(t *testing.T) {
	if JSONMode(nil) {
		t.Errorf("JSONMode(nil) = true, want false")
	}
}
