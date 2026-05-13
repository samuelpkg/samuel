// Package template embeds and renders the canonical AGENTS.md
// template that Samuel writes into every initialised project.
//
// The .tmpl file is the single source of truth: the CI line-check
// reads it, the AGENTS.md renderer reads it, and the docs site can
// link to it. Keeping both in one package lets us go:embed it
// without copying.
//
// The full PromptContext shape (Samuel, Project, Methodology, …) is
// defined in PRDs 0002 + 0004; this package ships the renderer and a
// minimal context type that the alpha smoke test can build against.
package template

import (
	_ "embed"
	"io"
	"strings"
	gotemplate "text/template"
)

//go:embed AGENTS.md.tmpl
var agentsMDSource string

// Source returns the raw template text. Useful for diffs and tests
// that need to count lines.
func Source() string { return agentsMDSource }

// Funcs are the helpers exposed inside templates. Keep this list
// auditable — every additional helper expands the surface that
// translator plugins must understand.
var Funcs = gotemplate.FuncMap{
	"join": func(items []string, sep string) string {
		return strings.Join(items, sep)
	},
	"hasPlugin": func(plugins map[string]any, name string) bool {
		if plugins == nil {
			return false
		}
		_, ok := plugins[name]
		return ok
	},
}

// AgentsMD returns a parsed *Template ready to Execute against a
// PromptContext-shaped value.
func AgentsMD() (*gotemplate.Template, error) {
	return gotemplate.New("AGENTS.md").Funcs(Funcs).Parse(agentsMDSource)
}

// RenderAgentsMD renders AGENTS.md against ctx and writes to w.
func RenderAgentsMD(w io.Writer, ctx any) error {
	t, err := AgentsMD()
	if err != nil {
		return err
	}
	return t.Execute(w, ctx)
}
