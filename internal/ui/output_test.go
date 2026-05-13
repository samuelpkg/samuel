package ui

import (
	"bytes"
	"encoding/json"
	stderrors "errors"
	"strings"
	"testing"

	"github.com/samuelpkg/samuel/internal/errors"
)

func TestSuccess_GoesToStdout(t *testing.T) {
	out, errw := &bytes.Buffer{}, &bytes.Buffer{}
	SetWriters(out, errw)
	t.Cleanup(func() { SetWriters(nil, nil) })

	Success("hello %s", "world")

	if !strings.Contains(out.String(), "hello world") {
		t.Errorf("stdout = %q", out.String())
	}
	if errw.Len() != 0 {
		t.Errorf("Success should not write to stderr; got %q", errw.String())
	}
}

func TestError_GoesToStderr(t *testing.T) {
	out, errw := &bytes.Buffer{}, &bytes.Buffer{}
	SetWriters(out, errw)
	t.Cleanup(func() { SetWriters(nil, nil) })

	Error("boom")

	if !strings.Contains(errw.String(), "boom") {
		t.Errorf("stderr = %q", errw.String())
	}
	if out.Len() != 0 {
		t.Errorf("Error should not write to stdout; got %q", out.String())
	}
}

func TestPrintJSON_EnvelopeShape(t *testing.T) {
	out, errw := &bytes.Buffer{}, &bytes.Buffer{}
	SetWriters(out, errw)
	t.Cleanup(func() { SetWriters(nil, nil) })

	PrintJSON("version", map[string]any{"v": "0.0.1"})

	var resp JSONResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v\n%s", err, out.String())
	}
	if resp.SchemaVersion != 4 {
		t.Errorf("schemaVersion = %d, want 4", resp.SchemaVersion)
	}
	if !resp.Success {
		t.Errorf("Success = false, want true")
	}
	if resp.Command != "version" {
		t.Errorf("Command = %q", resp.Command)
	}
}

func TestRenderError_MultiLineForStructured(t *testing.T) {
	out, errw := &bytes.Buffer{}, &bytes.Buffer{}
	SetWriters(out, errw)
	t.Cleanup(func() { SetWriters(nil, nil) })

	RenderError(&errors.Error{
		Component:   "lock",
		Problem:     "another samuel process is running",
		Cause:       "flock busy",
		Fix:         "wait for the other samuel process to finish",
		DocsURL:     "https://example/docs/errors/SAM-LOCK-001",
		Recoverable: true,
		Path:        "/tmp/.samuel/lock",
	})
	got := errw.String()
	for _, want := range []string{"another samuel process is running", "flock busy", "wait for", "Docs:", "Path:"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestRenderError_FallbackForBareError(t *testing.T) {
	out, errw := &bytes.Buffer{}, &bytes.Buffer{}
	SetWriters(out, errw)
	t.Cleanup(func() { SetWriters(nil, nil) })

	RenderError(stderrors.New("plain"))
	if !strings.Contains(errw.String(), "plain") {
		t.Errorf("stderr = %q", errw.String())
	}
}
