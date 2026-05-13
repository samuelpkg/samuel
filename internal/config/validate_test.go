package config

import (
	stderrors "errors"
	"strings"
	"testing"

	"github.com/ar4mirez/samuel/internal/errors"
)

func TestValidate_DefaultsPass(t *testing.T) {
	if err := Defaults().Validate(); err != nil {
		t.Errorf("Defaults() must pass Validate(); got %v", err)
	}
}

func TestValidate_MissingVersionFails(t *testing.T) {
	c := Defaults()
	c.Version = ""
	err := c.Validate()
	if err == nil {
		t.Fatal("expected validation error for missing version")
	}
	if !strings.Contains(err.Error(), "version") {
		t.Errorf("error should mention `version`; got %v", err)
	}
}

func TestValidate_UnknownMethodologyFails(t *testing.T) {
	c := Defaults()
	c.DefaultMethodology = "unknown-method"
	c.Methodology = nil
	err := c.Validate()
	if err == nil {
		t.Fatal("expected validation error for unknown methodology")
	}
	if !strings.Contains(err.Error(), "default_methodology") {
		t.Errorf("error should mention default_methodology; got %v", err)
	}
}

func TestValidate_BuiltinRalphAlwaysOK(t *testing.T) {
	c := &Config{Version: SchemaVersion, DefaultMethodology: "ralph"}
	if err := c.Validate(); err != nil {
		t.Errorf("ralph is a builtin and should validate without a Methodology block; got %v", err)
	}
}

func TestValidate_RejectsBadPluginKind(t *testing.T) {
	c := Defaults()
	c.Plugins = []PluginEntry{{Name: "bad", Kind: "shellscript"}}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected validation error for invalid kind")
	}
	if !strings.Contains(err.Error(), "invalid kind") {
		t.Errorf("error should call out invalid kind; got %v", err)
	}
}

func TestValidate_AcceptsAllKnownKinds(t *testing.T) {
	c := Defaults()
	c.Plugins = []PluginEntry{
		{Name: "a", Kind: "builtin"},
		{Name: "b", Kind: "skill"},
		{Name: "c", Kind: "wasm"},
		{Name: "d", Kind: "oci"},
	}
	if err := c.Validate(); err != nil {
		t.Errorf("all known kinds should validate; got %v", err)
	}
}

func TestValidate_EmptyPluginNameFails(t *testing.T) {
	c := Defaults()
	c.Plugins = []PluginEntry{{Name: "", Kind: "skill"}}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty name")
	}
}

func TestValidate_StructuredErrorTypePropagated(t *testing.T) {
	c := Defaults()
	c.Version = ""
	err := c.Validate()
	// The joined error wraps *errors.Error instances.
	var oe *errors.Error
	if !stderrors.As(err, &oe) {
		t.Errorf("Validate should return *errors.Error (possibly joined); got %T: %v", err, err)
	}
}
