package sync

// HookPoints names the lifecycle stages a methodology body can
// intercept. The framework calls each at a documented boundary in
// SyncFolderContext; methodology plugins register handlers in PRD 0004.
//
// This file defines the contract — the stubs return immediately. The
// orchestrator and CLI ship with `defaultHooks()` so existing tests
// don't need a methodology installed.
const (
	HookBefore        = "sync.before"
	HookAnalyzeFolder = "sync.analyze-folder"
	HookWriteAgentsMD = "sync.write-agents-md"
	HookAfter         = "sync.after"
)

// HookSet bundles the four hook callbacks. A nil receiver or nil field
// is safe; callers should always go through the methods, which
// internally check for nil.
type HookSet struct {
	BeforeFn        func(opts Options)
	AnalyzeFolderFn func(a *FolderAnalysis)
	// WriteAgentsMDFn receives a pointer to the rendered body so a
	// methodology can mutate it (e.g. append a guardrail block) before
	// the writer commits it to disk.
	WriteAgentsMDFn func(a *FolderAnalysis, content *string)
	AfterFn         func(r *Result)
}

// Before runs the registered sync.before handler, if any.
func (h *HookSet) Before(opts Options) {
	if h == nil || h.BeforeFn == nil {
		return
	}
	h.BeforeFn(opts)
}

// AnalyzeFolder runs the registered sync.analyze-folder handler, if any.
func (h *HookSet) AnalyzeFolder(a *FolderAnalysis) {
	if h == nil || h.AnalyzeFolderFn == nil {
		return
	}
	h.AnalyzeFolderFn(a)
}

// WriteAgentsMD runs the registered sync.write-agents-md handler. The
// content pointer is mutable so the methodology can rewrite the body
// before the writer commits it.
func (h *HookSet) WriteAgentsMD(a *FolderAnalysis, content *string) {
	if h == nil || h.WriteAgentsMDFn == nil {
		return
	}
	h.WriteAgentsMDFn(a, content)
}

// After runs the registered sync.after handler, if any.
func (h *HookSet) After(r *Result) {
	if h == nil || h.AfterFn == nil {
		return
	}
	h.AfterFn(r)
}

// defaultHooks returns a no-op HookSet. Replaced by a methodology-
// supplied HookSet in PRD 0004; until then SyncFolderContext uses this
// to keep the call sites uniform.
func defaultHooks() *HookSet { return &HookSet{} }
