// Package capability models the host resources a Samuel plugin can
// request and the grant flow that gates them at install time.
//
// Capabilities live in three namespaces:
//
//   - filesystem.read  / filesystem.write — path-glob allowlists
//   - exec                                — spawn host processes
//   - network.outbound                    — destination allowlist
//   - samuel.api                          — framework gRPC surface access
//   - assistant.invoke                    — coding-assistant invocation
//
// At install time the loader compares the requested capability set to
// the "safe default" baseline (read-only access scoped to /workspace).
// If everything fits inside the safe default the prompt is skipped and
// the grant is recorded with reason="safe-default". Anything outside
// (write, exec, network, samuel.api, assistant.invoke) is flagged as
// risky and prompted explicitly.
//
// The PromptFn signature is intentionally generic so the CLI can wire
// up a huh-based interactive UI without dragging huh into this package.
// Tests inject a deterministic PromptFn that returns the desired answer.
package capability

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/ar4mirez/samuel/internal/plugin/manifest"
)

// Kind enumerates capability families. The string value is what lands
// in samuel.lock for audit.
type Kind string

const (
	KindFilesystemRead  Kind = "filesystem.read"
	KindFilesystemWrite Kind = "filesystem.write"
	KindExec            Kind = "exec"
	KindNetworkOutbound Kind = "network.outbound"
	KindSamuelAPI       Kind = "samuel.api"
	KindAssistantInvoke Kind = "assistant.invoke"
)

// Requested is a single capability the plugin asks for.
type Requested struct {
	Kind Kind
	// Targets are kind-specific qualifiers: path globs for filesystem,
	// host[:port] patterns for network.outbound, empty for exec /
	// samuel.api / assistant.invoke.
	Targets []string
}

// Risky reports whether this capability falls outside the safe-default
// baseline and must therefore be confirmed by the user.
func (r Requested) Risky() bool {
	switch r.Kind {
	case KindFilesystemWrite, KindExec, KindNetworkOutbound, KindSamuelAPI, KindAssistantInvoke:
		return true
	case KindFilesystemRead:
		// filesystem.read is safe only when scoped to /workspace or
		// subpaths. A read outside that prefix counts as risky.
		for _, t := range r.Targets {
			if !insideWorkspace(t) {
				return true
			}
		}
		return false
	default:
		return true
	}
}

// insideWorkspace reports whether a glob is fully constrained to
// /workspace (or a subpath). Empty target = unrestricted, treated risky.
func insideWorkspace(glob string) bool {
	clean := filepath.Clean(glob)
	if clean == "/" || clean == "" {
		return false
	}
	return clean == "/workspace" || strings.HasPrefix(clean, "/workspace/")
}

// Summary is one human-readable line, used by both the prompt and the
// `samuel info` capability listing.
func (r Requested) Summary() string {
	if len(r.Targets) == 0 {
		return string(r.Kind)
	}
	return fmt.Sprintf("%s: %s", r.Kind, strings.Join(r.Targets, ", "))
}

// FromManifest extracts every capability requested by m, in stable
// order (filesystem.read, filesystem.write, exec, network.outbound,
// samuel.api, assistant.invoke).
func FromManifest(m *manifest.Manifest) []Requested {
	var out []Requested
	if len(m.Capabilities.Filesystem.Read) > 0 {
		out = append(out, Requested{
			Kind:    KindFilesystemRead,
			Targets: append([]string(nil), m.Capabilities.Filesystem.Read...),
		})
	}
	if len(m.Capabilities.Filesystem.Write) > 0 {
		out = append(out, Requested{
			Kind:    KindFilesystemWrite,
			Targets: append([]string(nil), m.Capabilities.Filesystem.Write...),
		})
	}
	if m.Capabilities.Exec {
		out = append(out, Requested{Kind: KindExec})
	}
	if len(m.Capabilities.Network.Outbound) > 0 {
		out = append(out, Requested{
			Kind:    KindNetworkOutbound,
			Targets: append([]string(nil), m.Capabilities.Network.Outbound...),
		})
	}
	if m.Capabilities.Samuel.API {
		out = append(out, Requested{Kind: KindSamuelAPI})
	}
	if m.Capabilities.Assistant.Invoke {
		out = append(out, Requested{Kind: KindAssistantInvoke})
	}
	return out
}

// Grant records one approved capability. Reason explains how it was
// approved: "safe-default" | "user-prompt" | "yes-flag" | "policy".
type Grant struct {
	Kind    Kind     `toml:"kind"`
	Targets []string `toml:"targets,omitempty"`
	Reason  string   `toml:"reason,omitempty"`
}

// AnyRisky reports whether the request set contains anything outside
// the safe-default baseline. Drives whether the loader prompts.
func AnyRisky(reqs []Requested) bool {
	for _, r := range reqs {
		if r.Risky() {
			return true
		}
	}
	return false
}

// PromptDecision is what a PromptFn returns. The CLI sets Granted=true
// when the user approves and false to refuse the install.
type PromptDecision struct {
	Granted bool
	Reason  string
}

// PromptFn is the abstract grant prompt. Tests substitute a
// deterministic implementation; the CLI uses huh.
type PromptFn func(plugin string, reqs []Requested) PromptDecision

// AutoYes is a PromptFn for `--yes` invocations: always grants, reason
// "yes-flag".
var AutoYes PromptFn = func(string, []Requested) PromptDecision {
	return PromptDecision{Granted: true, Reason: "yes-flag"}
}

// AutoDeny is a PromptFn for tests that exercise the deny path.
var AutoDeny PromptFn = func(string, []Requested) PromptDecision {
	return PromptDecision{Granted: false, Reason: "denied"}
}

// Decide runs the safe-default check, calls prompt when needed, and
// returns the grant list to record in samuel.lock. The boolean is true
// when the install should proceed.
//
//   - opts.YesAll skips the prompt and auto-approves everything.
//   - opts.NonInteractive without YesAll fails-closed on risky requests.
func Decide(plugin string, reqs []Requested, prompt PromptFn, opts DecideOptions) ([]Grant, bool, error) {
	if !AnyRisky(reqs) {
		return grantsForSafeDefault(reqs), true, nil
	}
	if opts.YesAll {
		return grantsWithReason(reqs, "yes-flag"), true, nil
	}
	if opts.NonInteractive {
		return nil, false, fmt.Errorf("capability: plugin %q requests risky capabilities and --yes was not set", plugin)
	}
	if prompt == nil {
		return nil, false, fmt.Errorf("capability: plugin %q requests risky capabilities and no prompt is available", plugin)
	}
	dec := prompt(plugin, reqs)
	if !dec.Granted {
		return nil, false, nil
	}
	reason := dec.Reason
	if reason == "" {
		reason = "user-prompt"
	}
	return grantsWithReason(reqs, reason), true, nil
}

// DecideOptions controls the prompt path.
type DecideOptions struct {
	YesAll         bool
	NonInteractive bool
}

func grantsForSafeDefault(reqs []Requested) []Grant {
	return grantsWithReason(reqs, "safe-default")
}

func grantsWithReason(reqs []Requested, reason string) []Grant {
	out := make([]Grant, 0, len(reqs))
	for _, r := range reqs {
		out = append(out, Grant{
			Kind:    r.Kind,
			Targets: append([]string(nil), r.Targets...),
			Reason:  reason,
		})
	}
	sort.Slice(out, func(i, j int) bool { return string(out[i].Kind) < string(out[j].Kind) })
	return out
}

// Match reports whether path satisfies any of grants' targets for
// the given kind. Path globs are evaluated with doublestar — the
// canonical glob library Cargo and gh both use for `.gitignore`-style
// matching.
func Match(grants []Grant, kind Kind, path string) bool {
	for _, g := range grants {
		if g.Kind != kind {
			continue
		}
		if len(g.Targets) == 0 {
			// Kind-only grants (exec, samuel.api, assistant.invoke).
			return true
		}
		for _, glob := range g.Targets {
			ok, _ := doublestar.PathMatch(glob, path)
			if ok {
				return true
			}
			// Also accept a literal prefix match for "/workspace"-style
			// targets without globs.
			if !strings.ContainsAny(glob, "*?[") {
				if path == glob || strings.HasPrefix(path, strings.TrimRight(glob, "/")+"/") {
					return true
				}
			}
		}
	}
	return false
}

// MatchHost reports whether host:port satisfies any network.outbound
// grant. "*" matches everything; "*.example.com" matches subdomains;
// "example.com" matches exact + any port. The match is host-only — the
// allowlist intentionally does not constrain ports yet.
func MatchHost(grants []Grant, hostPort string) bool {
	host := hostPort
	if i := strings.LastIndex(host, ":"); i >= 0 {
		host = host[:i]
	}
	for _, g := range grants {
		if g.Kind != KindNetworkOutbound {
			continue
		}
		for _, pat := range g.Targets {
			if pat == "*" {
				return true
			}
			if strings.HasPrefix(pat, "*.") {
				suf := pat[1:]
				if strings.HasSuffix(host, suf) || host == suf[1:] {
					return true
				}
				continue
			}
			if pat == host {
				return true
			}
		}
	}
	return false
}
