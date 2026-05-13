package wasm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"github.com/samuelpkg/samuel/internal/errors"
	"github.com/samuelpkg/samuel/internal/plugin"
	"github.com/samuelpkg/samuel/internal/plugin/capability"
	"github.com/samuelpkg/samuel/internal/plugin/manifest"
	"github.com/samuelpkg/samuel/internal/plugin/source"
)

// ModuleFileName is the canonical filename inside the install directory.
const ModuleFileName = "plugin.wasm"

// Plugin is the wasm-tier plugin.Plugin implementation.
type Plugin struct {
	Manifest_   manifest.Manifest
	ProjectDir  string
	SourceDir   string
	Runtime     *Runtime
	Grants      []capability.Grant
	CacheKeyOut string // populated by Install for samuel.lock recording
}

// New constructs a wasm-tier Plugin. rt may be nil; Detect/Uninstall do
// not need it. Install + Check require a runtime.
func New(m manifest.Manifest, projectDir, sourceDir string, rt *Runtime, grants []capability.Grant) *Plugin {
	return &Plugin{Manifest_: m, ProjectDir: projectDir, SourceDir: sourceDir, Runtime: rt, Grants: grants}
}

// Name returns the manifest plugin name.
func (p *Plugin) Name() string { return p.Manifest_.Name }

// Manifest returns the v2 framework Manifest snapshot.
func (p *Plugin) Manifest() plugin.Manifest {
	return plugin.Manifest{
		Name:    p.Manifest_.Name,
		Version: p.Manifest_.Version,
		Kind:    plugin.KindWasm,
		Summary: p.Manifest_.Summary,
	}
}

func (p *Plugin) pluginDir() string {
	return filepath.Join(p.ProjectDir, ".samuel", "plugins", p.Name())
}

// Detect reports installed=true when plugin.wasm exists.
func (p *Plugin) Detect(_ context.Context) (plugin.DetectResult, error) {
	dst := filepath.Join(p.pluginDir(), ModuleFileName)
	if _, err := os.Stat(dst); err != nil {
		return plugin.DetectResult{Installed: false, Path: dst}, nil
	}
	return plugin.DetectResult{Installed: true, Path: dst, Version: p.Manifest_.Version}, nil
}

// Install copies plugin.wasm out of the source tree, verifies the
// protocol export, and stages the file into .samuel/plugins/<name>/.
//
// Atomicity: tmp file → rename.
func (p *Plugin) Install(ctx context.Context, opts plugin.InstallOptions) (plugin.InstallResult, error) {
	res := plugin.InstallResult{Component: p.Name()}
	if p.SourceDir == "" {
		return res, &errors.Error{
			Component:   Component,
			Problem:     "wasm plugin has no source dir",
			Recoverable: false,
		}
	}
	modSrc := filepath.Join(p.SourceDir, ModuleFileName)
	if p.Manifest_.Wasm != nil && p.Manifest_.Wasm.Module != "" {
		modSrc = filepath.Join(p.SourceDir, p.Manifest_.Wasm.Module)
	}
	body, err := os.ReadFile(modSrc)
	if err != nil {
		return res, (&errors.Error{
			Component:   Component,
			Problem:     "wasm module not found",
			Path:        modSrc,
			Recoverable: true,
		}).Wrap(err)
	}
	if opts.DryRun {
		return res, nil
	}
	if p.Runtime != nil {
		if err := p.verifyProtocol(ctx, body); err != nil {
			return res, err
		}
	}
	target := p.pluginDir()
	if err := os.MkdirAll(target, 0o755); err != nil {
		return res, err
	}
	dst := filepath.Join(target, ModuleFileName)
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, body, 0o600); err != nil {
		return res, (&errors.Error{
			Component:   Component,
			Problem:     "cannot stage wasm module",
			Path:        tmp,
			Recoverable: true,
		}).Wrap(err)
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return res, (&errors.Error{
			Component:   Component,
			Problem:     "cannot rename staged wasm module",
			Path:        dst,
			Recoverable: true,
		}).Wrap(err)
	}
	// Copy manifest along for offline `samuel info`.
	if err := copyManifestIfPresent(p.SourceDir, target); err != nil {
		return res, err
	}
	p.CacheKeyOut = CacheKey(p.Name(), p.Manifest_.Version, body)
	res.Mutations = append(res.Mutations, plugin.Mutation{
		Kind:        plugin.MutationWasmLoaded,
		Path:        dst,
		Description: "installed wasm module " + p.Name(),
		Reverse: func(context.Context) error {
			return os.RemoveAll(target)
		},
	})
	return res, nil
}

func copyManifestIfPresent(src, dst string) error {
	srcM := filepath.Join(src, manifest.FileName)
	if _, err := os.Stat(srcM); err != nil {
		return nil
	}
	return source.CopyTree(srcM, filepath.Join(dst, manifest.FileName))
}

// verifyProtocol compiles the module and reads `samuel_protocol_version`.
// Rejects modules outside the supported window.
func (p *Plugin) verifyProtocol(ctx context.Context, body []byte) error {
	rt := p.Runtime
	cm, err := rt.wzRT.CompileModule(ctx, body)
	if err != nil {
		return (&errors.Error{
			Component:   Component,
			Problem:     "wasm module did not compile",
			Path:        p.Name(),
			Recoverable: true,
		}).Wrap(err)
	}
	defer cm.Close(ctx)
	if err := rt.RegisterHost(ctx); err != nil {
		return err
	}
	// Instantiate as a throwaway module to read the protocol export.
	cfg := wazero.NewModuleConfig().WithName("__samuel_probe_" + p.Name())
	mod, err := rt.wzRT.InstantiateModule(ctx, cm, cfg)
	if err != nil {
		// Modules without a _start entry are fine — InstantiateModule
		// will succeed for those; only return on hard instantiation
		// errors.
		return (&errors.Error{
			Component:   Component,
			Problem:     "wasm module instantiation failed",
			Recoverable: true,
		}).Wrap(err)
	}
	defer mod.Close(ctx)

	pv := mod.ExportedFunction("samuel_protocol_version")
	if pv == nil {
		// Treat the missing export as protocol=1 to keep the bar low
		// for the v2.0 plugin authoring experience. We log via the
		// stderr layer in production.
		return nil
	}
	out, err := pv.Call(ctx)
	if err != nil || len(out) == 0 {
		return fmt.Errorf("wasm: samuel_protocol_version call failed: %v", err)
	}
	pvNum := uint32(out[0])
	if pvNum < SupportedProtocolMin || pvNum > SupportedProtocolMax {
		return &errors.Error{
			Component:   Component,
			Problem:     fmt.Sprintf("wasm plugin protocol %d outside framework window [%d, %d]", pvNum, SupportedProtocolMin, SupportedProtocolMax),
			Fix:         "rebuild the plugin against a compatible Samuel SDK",
			Recoverable: true,
		}
	}
	return nil
}

// Check instantiates the module and calls health().
func (p *Plugin) Check(ctx context.Context) plugin.HealthStatus {
	dst := filepath.Join(p.pluginDir(), ModuleFileName)
	body, err := os.ReadFile(dst)
	if err != nil {
		return plugin.HealthStatus{Component: p.Name(), OK: false, Message: "wasm module missing: " + dst}
	}
	if p.Runtime == nil {
		return plugin.HealthStatus{Component: p.Name(), OK: true, Message: "wasm module present (no runtime probe)"}
	}
	cm, err := p.Runtime.wzRT.CompileModule(ctx, body)
	if err != nil {
		return plugin.HealthStatus{Component: p.Name(), OK: false, Message: "compile failed: " + err.Error()}
	}
	defer cm.Close(ctx)
	if err := p.Runtime.RegisterHost(ctx); err != nil {
		return plugin.HealthStatus{Component: p.Name(), OK: false, Message: err.Error()}
	}
	state := &HostState{Plugin: p.Name(), Grants: p.Grants}
	hostCtx := WithHostState(ctx, state)
	cfg := wazero.NewModuleConfig().WithName("__samuel_health_" + p.Name())
	mod, err := p.Runtime.wzRT.InstantiateModule(hostCtx, cm, cfg)
	if err != nil {
		return plugin.HealthStatus{Component: p.Name(), OK: false, Message: "instantiate failed: " + err.Error()}
	}
	defer mod.Close(hostCtx)
	health := mod.ExportedFunction("health")
	if health == nil {
		return plugin.HealthStatus{Component: p.Name(), OK: true, Message: "wasm module instantiated (no health export)"}
	}
	out, err := health.Call(hostCtx)
	if err != nil || len(out) == 0 {
		return plugin.HealthStatus{Component: p.Name(), OK: false, Message: "health call failed"}
	}
	if out[0] != 0 {
		return plugin.HealthStatus{Component: p.Name(), OK: false, Message: fmt.Sprintf("health returned %d", out[0])}
	}
	return plugin.HealthStatus{Component: p.Name(), OK: true, Message: "wasm health ok"}
}

// Uninstall removes the plugin directory.
func (p *Plugin) Uninstall(_ context.Context, opts plugin.UninstallOptions) (plugin.UninstallResult, error) {
	res := plugin.UninstallResult{Component: p.Name()}
	dir := p.pluginDir()
	if _, err := os.Stat(dir); err != nil {
		res.Skipped = true
		return res, nil
	}
	if opts.DryRun {
		return res, nil
	}
	if err := os.RemoveAll(dir); err != nil {
		return res, err
	}
	res.Mutations = append(res.Mutations, plugin.Mutation{
		Kind:        plugin.MutationWasmLoaded,
		Path:        dir,
		Description: "removed wasm plugin " + p.Name(),
	})
	return res, nil
}

// Run instantiates and invokes a named export with no args. Used for
// hook execution and tests. Returns the raw u64 result list.
func (p *Plugin) Run(ctx context.Context, export string) ([]uint64, error) {
	if p.Runtime == nil {
		return nil, fmt.Errorf("wasm: plugin %q has no runtime", p.Name())
	}
	dst := filepath.Join(p.pluginDir(), ModuleFileName)
	body, err := os.ReadFile(dst)
	if err != nil {
		return nil, err
	}
	cm, err := p.Runtime.wzRT.CompileModule(ctx, body)
	if err != nil {
		return nil, err
	}
	defer cm.Close(ctx)
	if err := p.Runtime.RegisterHost(ctx); err != nil {
		return nil, err
	}
	state := &HostState{Plugin: p.Name(), Grants: p.Grants}
	hostCtx := WithHostState(ctx, state)
	cfg := wazero.NewModuleConfig().WithName("__samuel_run_" + p.Name())
	mod, err := p.Runtime.wzRT.InstantiateModule(hostCtx, cm, cfg)
	if err != nil {
		return nil, err
	}
	defer mod.Close(hostCtx)
	fn := mod.ExportedFunction(export)
	if fn == nil {
		return nil, fmt.Errorf("wasm: plugin %q has no export %q", p.Name(), export)
	}
	return fn.Call(hostCtx)
}

// Compile-time guarantee.
var _ plugin.Plugin = (*Plugin)(nil)

// Silence unused import for wasi when no WASI module references it
// directly elsewhere in this file.
var _ = wasi_snapshot_preview1.Instantiate
