package config

import (
	stderrors "errors"
	"strings"

	"github.com/ar4mirez/samuel/internal/errors"
)

// validKinds is the closed enum every PluginEntry.Kind must belong to.
// New kinds land alongside the loader changes that introduce them.
var validKinds = map[string]bool{
	"builtin": true,
	"skill":   true,
	"wasm":    true,
	"oci":     true,
}

// builtinMethodologies is the set of methodology names that are
// always-known by the binary itself (no plugin install required).
// As of PRD 0002 the only one is "ralph". Other names must resolve
// to an entry in cfg.Methodology[<name>].
var builtinMethodologies = map[string]bool{
	"ralph": true,
}

// Validate enforces the samuel.toml schema rules called out in PRD
// 0002 §9. Returns a structured *errors.Error per failure; the caller
// joins them or surfaces the first one.
//
// Rules:
//
//	required: Version must be set
//	default_methodology resolves to ralph (builtin) or a key in
//	  Methodology
//	every [[plugins]] entry has a valid kind enum value
func (c *Config) Validate() error {
	if c == nil {
		return &errors.Error{
			Component:   "config",
			Problem:     "samuel.toml is nil",
			Recoverable: false,
		}
	}
	var errs []error
	if strings.TrimSpace(c.Version) == "" {
		errs = append(errs, &errors.Error{
			Component:   "config",
			Problem:     "samuel.toml missing required `version`",
			Fix:         "set version = \"" + SchemaVersion + "\" at the top of samuel.toml",
			DocsURL:     "https://ar4mirez.github.io/samuel/docs/errors/SAM-CFG-010",
			Recoverable: true,
		})
	}
	if c.DefaultMethodology != "" {
		if !builtinMethodologies[c.DefaultMethodology] {
			if _, ok := c.Methodology[c.DefaultMethodology]; !ok {
				errs = append(errs, &errors.Error{
					Component:   "config",
					Problem:     "default_methodology=\"" + c.DefaultMethodology + "\" is not installed",
					Fix:         "add a [methodology." + c.DefaultMethodology + "] block, install a plugin that provides it, or change to a builtin (e.g. ralph)",
					DocsURL:     "https://ar4mirez.github.io/samuel/docs/errors/SAM-CFG-011",
					Recoverable: true,
				})
			}
		}
	}
	for i, p := range c.Plugins {
		if p.Name == "" {
			errs = append(errs, &errors.Error{
				Component:   "config",
				Problem:     "plugin entry has empty `name`",
				Fix:         "set name=\"…\" on every [[plugins]] block",
				Cause:       indexedCause(i),
				Recoverable: true,
			})
		}
		if !validKinds[p.Kind] {
			errs = append(errs, &errors.Error{
				Component:   "config",
				Problem:     "plugin \"" + p.Name + "\" has invalid kind \"" + p.Kind + "\"",
				Fix:         "use one of: builtin, skill, wasm, oci",
				DocsURL:     "https://ar4mirez.github.io/samuel/docs/errors/SAM-CFG-012",
				Cause:       indexedCause(i),
				Recoverable: true,
			})
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return stderrors.Join(errs...)
}

func indexedCause(i int) string {
	return "[[plugins]] entry #" + itoa(i)
}

// itoa is the std-lib-free integer formatter — keeps validate.go from
// pulling fmt for a single int rendering.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}
