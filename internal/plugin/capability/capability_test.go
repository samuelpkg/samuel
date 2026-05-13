package capability

import (
	"testing"

	"github.com/samuelpkg/samuel/internal/plugin/manifest"
)

func TestFromManifest(t *testing.T) {
	m := &manifest.Manifest{}
	m.Capabilities.Filesystem.Read = []string{"/workspace"}
	m.Capabilities.Filesystem.Write = []string{"/workspace/**/*.md"}
	m.Capabilities.Exec = true
	m.Capabilities.Network.Outbound = []string{"api.openai.com"}
	m.Capabilities.Assistant.Invoke = true
	reqs := FromManifest(m)
	if len(reqs) != 5 {
		t.Fatalf("expected 5 caps, got %d: %+v", len(reqs), reqs)
	}
}

func TestRisky_SafeDefault(t *testing.T) {
	r := Requested{Kind: KindFilesystemRead, Targets: []string{"/workspace"}}
	if r.Risky() {
		t.Errorf("read-only workspace must be safe default")
	}
	r2 := Requested{Kind: KindFilesystemRead, Targets: []string{"/etc"}}
	if !r2.Risky() {
		t.Errorf("read outside /workspace must be risky")
	}
}

func TestDecide_SafeDefault_NoPrompt(t *testing.T) {
	reqs := []Requested{{Kind: KindFilesystemRead, Targets: []string{"/workspace"}}}
	called := false
	prompt := PromptFn(func(string, []Requested) PromptDecision {
		called = true
		return PromptDecision{Granted: false}
	})
	grants, ok, err := Decide("test", reqs, prompt, DecideOptions{})
	if err != nil || !ok {
		t.Fatalf("Decide should succeed: ok=%v err=%v", ok, err)
	}
	if called {
		t.Errorf("prompt must not fire for safe-default requests")
	}
	if len(grants) != 1 || grants[0].Reason != "safe-default" {
		t.Errorf("unexpected grant: %+v", grants)
	}
}

func TestDecide_Risky_PromptsForUser(t *testing.T) {
	reqs := []Requested{
		{Kind: KindFilesystemRead, Targets: []string{"/workspace"}},
		{Kind: KindExec},
	}
	calls := 0
	prompt := PromptFn(func(_ string, got []Requested) PromptDecision {
		calls++
		if len(got) != 2 {
			t.Errorf("prompt got %d caps, want 2", len(got))
		}
		return PromptDecision{Granted: true}
	})
	grants, ok, err := Decide("test", reqs, prompt, DecideOptions{})
	if err != nil || !ok {
		t.Fatalf("Decide: ok=%v err=%v", ok, err)
	}
	if calls != 1 {
		t.Errorf("prompt should fire exactly once, got %d", calls)
	}
	if len(grants) != 2 {
		t.Errorf("expected 2 grants, got %d", len(grants))
	}
}

func TestDecide_YesFlagBypasses(t *testing.T) {
	reqs := []Requested{{Kind: KindExec}}
	grants, ok, err := Decide("p", reqs, AutoDeny, DecideOptions{YesAll: true})
	if !ok || err != nil {
		t.Fatalf("YesAll should approve risky: ok=%v err=%v", ok, err)
	}
	if grants[0].Reason != "yes-flag" {
		t.Errorf("reason = %s, want yes-flag", grants[0].Reason)
	}
}

func TestDecide_NonInteractiveFailsClosed(t *testing.T) {
	reqs := []Requested{{Kind: KindExec}}
	_, ok, err := Decide("p", reqs, nil, DecideOptions{NonInteractive: true})
	if ok || err == nil {
		t.Errorf("non-interactive must fail-closed on risky caps")
	}
}

func TestMatch_PathGlobs(t *testing.T) {
	grants := []Grant{{Kind: KindFilesystemWrite, Targets: []string{"/workspace/**/*.md"}}}
	if !Match(grants, KindFilesystemWrite, "/workspace/docs/readme.md") {
		t.Errorf("glob match failed")
	}
	if Match(grants, KindFilesystemWrite, "/workspace/secret.tar") {
		t.Errorf("glob should not match non-.md")
	}
	if Match(grants, KindFilesystemWrite, "/etc/passwd") {
		t.Errorf("glob should not match outside /workspace")
	}
}

func TestMatch_PathPrefix(t *testing.T) {
	grants := []Grant{{Kind: KindFilesystemRead, Targets: []string{"/workspace"}}}
	if !Match(grants, KindFilesystemRead, "/workspace") {
		t.Errorf("exact prefix should match")
	}
	if !Match(grants, KindFilesystemRead, "/workspace/sub/file") {
		t.Errorf("subpath should match")
	}
	if Match(grants, KindFilesystemRead, "/etc/passwd") {
		t.Errorf("outside paths must not match")
	}
}

func TestMatchHost(t *testing.T) {
	grants := []Grant{{Kind: KindNetworkOutbound, Targets: []string{"api.openai.com", "*.example.com"}}}
	cases := map[string]bool{
		"api.openai.com:443":      true,
		"foo.example.com:443":     true,
		"example.com":             true,
		"bar.com":                 false,
	}
	for hp, want := range cases {
		if got := MatchHost(grants, hp); got != want {
			t.Errorf("MatchHost(%q) = %v, want %v", hp, got, want)
		}
	}
}
