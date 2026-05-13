package template

import (
	"bytes"
	"strings"
	"testing"
)

// agentsMDLineBudget is the hard limit on AGENTS.md template length.
// RFD 0001 / PRD 0001 mandate ≤150 lines so the embedded template stays
// readable in tool-specific copies. CI runs agents-md-check.yml against
// the same budget; this test is the dev-side mirror.
const agentsMDLineBudget = 150

func TestAgentsMDTemplate_LineBudget(t *testing.T) {
	src := Source()
	lines := strings.Count(src, "\n")
	if !strings.HasSuffix(src, "\n") {
		lines++
	}
	if lines > agentsMDLineBudget {
		t.Fatalf("AGENTS.md.tmpl is %d lines; budget is %d", lines, agentsMDLineBudget)
	}
	t.Logf("AGENTS.md.tmpl line count: %d / %d budget", lines, agentsMDLineBudget)
}

// maxConfigCtx is a saturated context used to verify the rendered
// output, not just the source, stays under the budget. PRD 0001 says
// the rendered max-config form must be ≤150 lines — exercised here.
func maxConfigCtx() map[string]any {
	return map[string]any{
		"Samuel":  map[string]any{"Version": "2.0.0-alpha.1", "Binary": "/usr/local/bin/samuel"},
		"Project": map[string]any{"Name": "demo", "Root": "/home/u/demo", "Branch": "main", "Detected": []string{"go", "typescript", "python"}},
		"Methodology": map[string]any{
			"Name":   "ralph",
			"Source": "built-in",
		},
		"Iteration": map[string]any{"Number": 7, "Max": 25, "Type": "implementation", "LastDiscoveryAt": 3},
		"Config": map[string]any{
			"Agent":         "claude",
			"MaxIterations": 25,
			"QualityChecks": []string{"go build ./...", "go test -race ./...", "go vet ./...", "golangci-lint run"},
		},
		"Guardrails": map[string]any{
			"MaxFunctionLines": 50,
			"MaxFileLines":     300,
			"RequireTests":     true,
			"Extra":            []string{"No magic numbers", "Conventional commits"},
		},
		"Paths": map[string]any{
			"SamuelDir":           ".samuel/",
			"RunDir":              ".samuel/run/",
			"PRDFile":             ".samuel/run/prd.toon",
			"ProgressFile":        ".samuel/run/progress.md",
			"TaskContextFile":     ".samuel/run/task-context.toon",
			"ProgressContextFile": ".samuel/run/progress-context.md",
			"SnapshotFile":        ".samuel/run/project-snapshot.toon",
			"AgentsMD":            "AGENTS.md",
		},
		"State": map[string]any{
			"TotalTasks":      60,
			"PendingTasks":    47,
			"CompletedTasks":  10,
			"InProgressTasks": 2,
			"BlockedTasks":    1,
			"NextTask": map[string]any{
				"ID": "5.3", "Title": "Implement Load helper", "Priority": "high",
				"Description": "Read samuel.toml and return *Config or ErrNotFound.",
			},
		},
		"Mode":  map[string]any{"IsDiscovery": false, "IsPilot": false},
		"Hooks": map[string]any{"Registered": map[string][]string{"quality.check": {"pytest-runner", "lint-strict"}}},
		"Plugins": map[string]any{
			"go-guide":    map[string]any{},
			"ralph":       map[string]any{},
			"create-rfd":  map[string]any{},
			"sync-agents": map[string]any{},
		},
	}
}

func TestAgentsMDTemplate_RendersUnderBudget(t *testing.T) {
	var buf bytes.Buffer
	if err := RenderAgentsMD(&buf, maxConfigCtx()); err != nil {
		t.Fatalf("render: %v", err)
	}
	rendered := buf.String()
	lines := strings.Count(rendered, "\n")
	if !strings.HasSuffix(rendered, "\n") {
		lines++
	}
	if lines > agentsMDLineBudget {
		t.Fatalf("rendered AGENTS.md is %d lines; budget is %d\n---\n%s", lines, agentsMDLineBudget, rendered)
	}
	t.Logf("rendered AGENTS.md line count: %d / %d budget", lines, agentsMDLineBudget)
}

func TestAgentsMDTemplate_RendersAllVisibleSections(t *testing.T) {
	var buf bytes.Buffer
	if err := RenderAgentsMD(&buf, maxConfigCtx()); err != nil {
		t.Fatalf("render: %v", err)
	}
	for _, want := range []string{"4D Methodology", "Boundaries", "Quick reference", "Plugins active", "Guardrails", "Quality checks", "Project context"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("rendered template missing section %q", want)
		}
	}
}
