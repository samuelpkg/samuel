// Package config models samuel.toml (the user-edited project config)
// and samuel.lock (the machine-managed resolved-plugin lockfile).
//
// Schemas match the design in .wiki/entities/config-format.md. TOML is
// the source of truth; YAML is intentionally NOT supported in v2 — the
// "TOML primary, YAML secondary" wording in v1 docs was relaxed for
// the v2 launch (one format, one parser, no ambiguity).
//
// Both files use pelletier/go-toml v2 for marshal/unmarshal. Save uses
// the atomic write-tmp-then-rename pattern v1 hardened in its
// orchestrator port — half-written config files are the canonical way
// to lose a user's setup.
package config

// FileName values are the on-disk names callers pass to Load/Save.
const (
	ProjectFile = "samuel.toml"
	LockFile    = "samuel.lock"
)

// SchemaVersion is the version stamped into newly-written samuel.toml
// files. Bumped when the schema gains breaking changes.
const SchemaVersion = "1"

// Config is the parsed samuel.toml.
//
// The structure follows the wiki entity:
//
//	version = "1"
//	default_methodology = "ralph"
//	[[plugins]] name="…" version="…" kind="skill|wasm|oci"
//	[methodology.<name>] enabled=… agent=… max_iterations=… quality_checks=[…]
//	[guardrails] max_function_lines=… max_file_lines=… require_tests=…
//	[[registries]] name="…" url="…" default=true
type Config struct {
	Version            string                 `toml:"version"`
	DefaultMethodology string                 `toml:"default_methodology,omitempty"`
	Plugins            []PluginEntry          `toml:"plugins,omitempty"`
	Methodology        map[string]Methodology `toml:"methodology,omitempty"`
	Guardrails         *Guardrails            `toml:"guardrails,omitempty"`
	Registries         []Registry             `toml:"registries,omitempty"`
}

// PluginEntry is one [[plugins]] block in samuel.toml.
type PluginEntry struct {
	Name    string `toml:"name"`
	Version string `toml:"version,omitempty"`
	Kind    string `toml:"kind"`
	Source  string `toml:"source,omitempty"`
}

// Methodology is one [methodology.<name>] block. The map key on the
// parent Config carries the methodology name (e.g. "ralph"), so Name
// is not stored here.
type Methodology struct {
	Enabled       bool     `toml:"enabled"`
	Agent         string   `toml:"agent,omitempty"`
	MaxIterations int      `toml:"max_iterations,omitempty"`
	QualityChecks []string `toml:"quality_checks,omitempty"`
	Encoding      Encoding `toml:"encoding,omitempty"`
}

// Encoding pins per-file encodings for a methodology's runtime files.
// Defaults: structured=toon, progress=md (see toon-evaluation memo).
type Encoding struct {
	Structured string `toml:"structured,omitempty"`
	Progress   string `toml:"progress,omitempty"`
}

// Guardrails is the [guardrails] block — code-quality limits used by
// the methodology's quality checks and the AGENTS.md template.
type Guardrails struct {
	MaxFunctionLines int  `toml:"max_function_lines,omitempty"`
	MaxFileLines     int  `toml:"max_file_lines,omitempty"`
	RequireTests     bool `toml:"require_tests,omitempty"`
}

// Registry is one [[registries]] block — a remote plugin index.
type Registry struct {
	Name    string `toml:"name,omitempty"`
	URL     string `toml:"url"`
	Default bool   `toml:"default,omitempty"`
}

// Defaults returns the zero-value-with-sensible-defaults Config used
// when a project has no samuel.toml yet (e.g. before `samuel init`).
func Defaults() *Config {
	return &Config{
		Version:            SchemaVersion,
		DefaultMethodology: "ralph",
		Methodology: map[string]Methodology{
			"ralph": {
				Enabled:       true,
				Agent:         "claude",
				MaxIterations: 25,
				Encoding:      Encoding{Structured: "toon", Progress: "md"},
			},
		},
		Guardrails: &Guardrails{
			MaxFunctionLines: 50,
			MaxFileLines:     300,
			RequireTests:     true,
		},
		Registries: []Registry{
			{Name: "official", URL: "github.com/ar4mirez/samuel-registry", Default: true},
		},
	}
}

// Lockfile models samuel.lock.
type Lockfile struct {
	Version      string         `toml:"version"`
	GeneratedAt  string         `toml:"generated_at,omitempty"`
	TOONSpec     string         `toml:"toon_spec,omitempty"`
	Plugins      []LockedPlugin `toml:"plugins,omitempty"`
	Capabilities []string       `toml:"capabilities,omitempty"`
}

// LockedPlugin is the resolved-version + signature record for one
// installed plugin.
type LockedPlugin struct {
	Name         string   `toml:"name"`
	Version      string   `toml:"version"`
	Kind         string   `toml:"kind"`
	Digest       string   `toml:"digest,omitempty"`
	Source       string   `toml:"source,omitempty"`
	Capabilities []string `toml:"capabilities,omitempty"`
	Signed       bool     `toml:"signed,omitempty"`
}
