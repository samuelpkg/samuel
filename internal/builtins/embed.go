// Package builtins embeds Samuel's built-in skill content into the
// binary so `samuel init` can sync it into ~/.samuel/builtins/ without
// network access.
//
// The four built-ins that ship with v2 are:
//
//   - ralph                — default methodology loop (PRD 0004 body)
//   - create-skill         — scaffold a new Agent Skill
//   - sync                 — per-folder AGENTS.md generator
//   - generate-agents-md   — root AGENTS.md generator (folded into sync)
//
// Each is a directory under content/ with at minimum a SKILL.md manifest
// following the Agent Skills standard (YAML frontmatter + body). Future
// supporting files (prompts/, scripts/) land alongside SKILL.md as the
// bodies graduate from placeholder to executable.
package builtins

import (
	"embed"
	"io/fs"
)

//go:embed all:content
var contentFS embed.FS

// FS returns a read-only filesystem rooted at content/ — top-level
// entries are skill directories (ralph, create-skill, sync,
// generate-agents-md). Callers pass this to SamuelComponent.Source
// during install.
func FS() fs.FS {
	sub, err := fs.Sub(contentFS, "content")
	if err != nil {
		// The embed directive guarantees content/ exists at build time;
		// reaching this branch means the directive was removed. Panic
		// so the regression is caught by `go test` instead of producing
		// a runtime nil filesystem the rest of the stack misuses.
		panic("builtins: embedded content/ directory missing — check the //go:embed directive in embed.go: " + err.Error())
	}
	return sub
}

// SkillNames returns the slugs of the built-ins that ship with v2 in
// deterministic order. Useful for `samuel doctor` rendering and for
// tests that need to assert "the four canonical built-ins are present."
func SkillNames() []string {
	return []string{"ralph", "create-skill", "sync", "generate-agents-md"}
}
