package plugin

import "context"

// SkillPlugin installs a text-only Agent Skill: a directory containing
// SKILL.md and supporting prompts. Implementation lands in PRD 0003.
type SkillPlugin struct {
	ManifestData Manifest
}

// WasmPlugin installs and runs a sandboxed WASM module via wazero.
// Implementation lands in PRD 0003.
type WasmPlugin struct {
	ManifestData Manifest
}

// OciPlugin installs a heavy tool delivered as an OCI image. Pulled
// via Docker / Podman; sandbox boundaries are container-level.
// Implementation lands in PRD 0003.
type OciPlugin struct {
	ManifestData Manifest
}

// notImplemented is the shared placeholder body. Lifecycle methods
// land in PRD 0003 — the stubs are here so the rest of the framework
// can take *SkillPlugin / *WasmPlugin / *OciPlugin values today.
func notImplemented(name string) (InstallResult, error) {
	return InstallResult{Component: name}, &notImplementedError{kind: name}
}

type notImplementedError struct{ kind string }

func (e *notImplementedError) Error() string {
	return "plugin: " + e.kind + " lifecycle not yet implemented (PRD 0003)"
}

// --- SkillPlugin ---

func (p *SkillPlugin) Name() string       { return p.ManifestData.Name }
func (p *SkillPlugin) Manifest() Manifest { return p.ManifestData }
func (p *SkillPlugin) Detect(context.Context) (DetectResult, error) {
	return DetectResult{}, nil
}
func (p *SkillPlugin) Install(context.Context, InstallOptions) (InstallResult, error) {
	return notImplemented("skill")
}
func (p *SkillPlugin) Check(context.Context) HealthStatus {
	return HealthStatus{Component: p.Name(), OK: false, Message: "lifecycle not yet implemented"}
}
func (p *SkillPlugin) Uninstall(context.Context, UninstallOptions) (UninstallResult, error) {
	return UninstallResult{Component: p.Name(), Skipped: true}, nil
}

// --- WasmPlugin ---

func (p *WasmPlugin) Name() string       { return p.ManifestData.Name }
func (p *WasmPlugin) Manifest() Manifest { return p.ManifestData }
func (p *WasmPlugin) Detect(context.Context) (DetectResult, error) {
	return DetectResult{}, nil
}
func (p *WasmPlugin) Install(context.Context, InstallOptions) (InstallResult, error) {
	return notImplemented("wasm")
}
func (p *WasmPlugin) Check(context.Context) HealthStatus {
	return HealthStatus{Component: p.Name(), OK: false, Message: "lifecycle not yet implemented"}
}
func (p *WasmPlugin) Uninstall(context.Context, UninstallOptions) (UninstallResult, error) {
	return UninstallResult{Component: p.Name(), Skipped: true}, nil
}

// --- OciPlugin ---

func (p *OciPlugin) Name() string       { return p.ManifestData.Name }
func (p *OciPlugin) Manifest() Manifest { return p.ManifestData }
func (p *OciPlugin) Detect(context.Context) (DetectResult, error) {
	return DetectResult{}, nil
}
func (p *OciPlugin) Install(context.Context, InstallOptions) (InstallResult, error) {
	return notImplemented("oci")
}
func (p *OciPlugin) Check(context.Context) HealthStatus {
	return HealthStatus{Component: p.Name(), OK: false, Message: "lifecycle not yet implemented"}
}
func (p *OciPlugin) Uninstall(context.Context, UninstallOptions) (UninstallResult, error) {
	return UninstallResult{Component: p.Name(), Skipped: true}, nil
}

// Compile-time guarantees that each stub satisfies Plugin. The PRD
// 0003 implementations replace bodies, not signatures.
var (
	_ Plugin = (*SkillPlugin)(nil)
	_ Plugin = (*WasmPlugin)(nil)
	_ Plugin = (*OciPlugin)(nil)
)
