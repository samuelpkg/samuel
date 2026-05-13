// Package wasm implements the WASM-tier plugin loader using
// tetratelabs/wazero. The runtime is pure-Go and embedded in the samuel
// binary — no host wasm-runtime install required.
//
// Host functions exposed under the `samuel` namespace:
//
//   - samuel.fs.read(ptr, len) -> bytes  — capability-gated by filesystem.read
//   - samuel.fs.write(ptr, len, body)    — capability-gated by filesystem.write
//   - samuel.exec(cmd) -> exit_code      — capability-gated by exec
//   - samuel.net.outbound(host, body)    — capability-gated by network.outbound
//   - samuel.log(level, msg)             — always allowed
//   - samuel.config.get(key) -> value    — always allowed (read-only)
//   - samuel.callback(name, payload)     — always allowed
//
// Each module must export `samuel_protocol_version() -> u32`. The
// runtime rejects modules whose protocol falls outside the framework's
// supported range (RFD 0001 resolution #2). Modules also expose
// `health() -> i32` (0=ok, non-zero=fail) used by Check.
//
// Compiled modules are cached at
// ~/.samuel/cache/wasm-compiled/<plugin>@<version>-<hash>.bin
// per wazero's CompilationCache contract.
package wasm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"github.com/ar4mirez/samuel/internal/errors"
	"github.com/ar4mirez/samuel/internal/plugin/capability"
)

// Component is the structured-error namespace.
const Component = "plugin/wasm"

// SupportedProtocolMin / Max define the protocol-version window the
// framework accepts. Modules report their protocol via
// `samuel_protocol_version() -> u32` (see RFD 0001).
const (
	SupportedProtocolMin = 1
	SupportedProtocolMax = 1
)

// Runtime is the per-process wazero runtime + host-function registry.
// One Runtime is shared across all WASM plugins to enable module reuse
// and a single compilation cache.
type Runtime struct {
	cacheDir string
	wzRT     wazero.Runtime
	cache    wazero.CompilationCache
	hostBuilt bool
}

// NewRuntime constructs a Runtime backed by an on-disk wazero
// compilation cache rooted at cacheDir. cacheDir may be empty (cache
// disabled, slower startup, useful for tests).
func NewRuntime(ctx context.Context, cacheDir string) (*Runtime, error) {
	cfg := wazero.NewRuntimeConfig()
	var cache wazero.CompilationCache
	if cacheDir != "" {
		if err := os.MkdirAll(cacheDir, 0o700); err != nil {
			return nil, (&errors.Error{
				Component:   Component,
				Problem:     "cannot create wasm cache dir",
				Path:        cacheDir,
				Recoverable: true,
			}).Wrap(err)
		}
		var err error
		cache, err = wazero.NewCompilationCacheWithDir(cacheDir)
		if err != nil {
			return nil, (&errors.Error{
				Component:   Component,
				Problem:     "cannot init wasm compilation cache",
				Path:        cacheDir,
				Recoverable: true,
			}).Wrap(err)
		}
		cfg = cfg.WithCompilationCache(cache)
	}
	rt := wazero.NewRuntimeWithConfig(ctx, cfg)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, rt); err != nil {
		_ = rt.Close(ctx)
		return nil, (&errors.Error{
			Component:   Component,
			Problem:     "cannot register WASI snapshot preview 1",
			Recoverable: false,
		}).Wrap(err)
	}
	return &Runtime{cacheDir: cacheDir, wzRT: rt, cache: cache}, nil
}

// Close releases runtime + cache resources.
func (r *Runtime) Close(ctx context.Context) error {
	if r == nil {
		return nil
	}
	var err error
	if r.wzRT != nil {
		err = r.wzRT.Close(ctx)
	}
	if r.cache != nil {
		_ = r.cache.Close(ctx)
	}
	return err
}

// HostState holds the per-instance authorization context for the
// `samuel.*` host functions. The instance log buffer captures everything
// the plugin emits via samuel.log so tests can assert.
type HostState struct {
	Plugin     string
	Grants     []capability.Grant
	LogBuf     []string
	OutboundOK func(host string) bool
	ExecOK     func(cmd string) bool
}

// Authorize is the universal capability gate the host functions call
// before doing any privileged work. Returns nil iff the action is
// allowed under the instance's grant set.
func (s *HostState) Authorize(kind capability.Kind, target string) error {
	switch kind {
	case capability.KindFilesystemRead, capability.KindFilesystemWrite:
		if !capability.Match(s.Grants, kind, target) {
			return fmt.Errorf("wasm: plugin %q lacks %s capability for %q", s.Plugin, kind, target)
		}
	case capability.KindNetworkOutbound:
		if !capability.MatchHost(s.Grants, target) {
			return fmt.Errorf("wasm: plugin %q lacks network.outbound capability for %q", s.Plugin, target)
		}
	case capability.KindExec, capability.KindSamuelAPI, capability.KindAssistantInvoke:
		if !capability.Match(s.Grants, kind, "") {
			return fmt.Errorf("wasm: plugin %q lacks %s capability", s.Plugin, kind)
		}
	}
	return nil
}

// RegisterHost installs the `samuel.*` host functions on the runtime.
// Call once per Runtime; idempotent. Per-instance authorization is
// performed by reading HostState off the wazero context (see Module.Run).
func (r *Runtime) RegisterHost(ctx context.Context) error {
	if r.hostBuilt {
		return nil
	}
	b := r.wzRT.NewHostModuleBuilder("samuel")
	b.NewFunctionBuilder().
		WithFunc(hostLog).
		Export("log")
	b.NewFunctionBuilder().
		WithFunc(hostFsRead).
		Export("fs_read")
	b.NewFunctionBuilder().
		WithFunc(hostFsWrite).
		Export("fs_write")
	b.NewFunctionBuilder().
		WithFunc(hostExec).
		Export("exec")
	b.NewFunctionBuilder().
		WithFunc(hostNetOutbound).
		Export("net_outbound")
	b.NewFunctionBuilder().
		WithFunc(hostConfigGet).
		Export("config_get")
	b.NewFunctionBuilder().
		WithFunc(hostCallback).
		Export("callback")
	if _, err := b.Instantiate(ctx); err != nil {
		return (&errors.Error{
			Component:   Component,
			Problem:     "cannot install samuel.* host module",
			Recoverable: false,
		}).Wrap(err)
	}
	r.hostBuilt = true
	return nil
}

// Host function bodies — the contract is intentionally simple: each
// function reads its arguments off linear memory, applies the
// capability check via HostState (carried in context), and either
// performs the privileged work or returns an error code. Real
// implementations land alongside the first WASM plugin.

// hostStateKey is the context.Context key for the per-call HostState.
type hostStateKey struct{}

// WithHostState attaches a HostState pointer to ctx for host functions
// to read. The pointer is mutable so host functions can append to LogBuf.
func WithHostState(ctx context.Context, s *HostState) context.Context {
	return context.WithValue(ctx, hostStateKey{}, s)
}

func hostStateFrom(ctx context.Context) *HostState {
	v, _ := ctx.Value(hostStateKey{}).(*HostState)
	return v
}

func hostLog(ctx context.Context, m api.Module, levelPtr, levelLen, msgPtr, msgLen uint32) {
	s := hostStateFrom(ctx)
	if s == nil {
		return
	}
	level := readString(m, levelPtr, levelLen)
	msg := readString(m, msgPtr, msgLen)
	s.LogBuf = append(s.LogBuf, fmt.Sprintf("[%s] %s", level, msg))
}

func hostFsRead(ctx context.Context, m api.Module, pathPtr, pathLen uint32) uint64 {
	s := hostStateFrom(ctx)
	if s == nil {
		return 0
	}
	path := readString(m, pathPtr, pathLen)
	if err := s.Authorize(capability.KindFilesystemRead, path); err != nil {
		s.LogBuf = append(s.LogBuf, "denied: "+err.Error())
		return 0
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	return uint64(writeBytes(m, body))
}

func hostFsWrite(ctx context.Context, m api.Module, pathPtr, pathLen, bodyPtr, bodyLen uint32) uint32 {
	s := hostStateFrom(ctx)
	if s == nil {
		return 1
	}
	path := readString(m, pathPtr, pathLen)
	if err := s.Authorize(capability.KindFilesystemWrite, path); err != nil {
		s.LogBuf = append(s.LogBuf, "denied: "+err.Error())
		return 2
	}
	body, ok := m.Memory().Read(bodyPtr, bodyLen)
	if !ok {
		return 1
	}
	if err := os.WriteFile(path, body, 0o600); err != nil {
		return 1
	}
	return 0
}

func hostExec(ctx context.Context, m api.Module, cmdPtr, cmdLen uint32) uint32 {
	s := hostStateFrom(ctx)
	if s == nil {
		return 1
	}
	cmd := readString(m, cmdPtr, cmdLen)
	if err := s.Authorize(capability.KindExec, ""); err != nil {
		s.LogBuf = append(s.LogBuf, "denied: "+err.Error())
		return 2
	}
	if s.ExecOK != nil && !s.ExecOK(cmd) {
		return 3
	}
	// The framework does NOT spawn host processes from inside WASM in
	// v2.0 — exec is reserved for future plugin hooks. The capability
	// gate is in place so tests can assert the deny path; a future
	// release adds the actual exec backend.
	s.LogBuf = append(s.LogBuf, "exec: "+cmd)
	return 0
}

func hostNetOutbound(ctx context.Context, m api.Module, hostPtr, hostLen, bodyPtr, bodyLen uint32) uint32 {
	s := hostStateFrom(ctx)
	if s == nil {
		return 1
	}
	host := readString(m, hostPtr, hostLen)
	if err := s.Authorize(capability.KindNetworkOutbound, host); err != nil {
		s.LogBuf = append(s.LogBuf, "denied: "+err.Error())
		return 2
	}
	if s.OutboundOK != nil && !s.OutboundOK(host) {
		return 3
	}
	_ = bodyPtr
	_ = bodyLen
	return 0
}

func hostConfigGet(ctx context.Context, m api.Module, keyPtr, keyLen uint32) uint64 {
	_ = keyPtr
	_ = keyLen
	return 0
}

func hostCallback(ctx context.Context, m api.Module, namePtr, nameLen, payloadPtr, payloadLen uint32) uint32 {
	_ = namePtr
	_ = nameLen
	_ = payloadPtr
	_ = payloadLen
	return 0
}

func readString(m api.Module, ptr, length uint32) string {
	b, ok := m.Memory().Read(ptr, length)
	if !ok {
		return ""
	}
	return string(b)
}

// writeBytes attempts to allocate via the module's `samuel_alloc` export
// (if present) and write the buffer there, returning the pointer. When
// no allocator is present we return 0 — the host function contract is
// "best effort" for v2.0.
func writeBytes(m api.Module, b []byte) uint32 {
	allocFn := m.ExportedFunction("samuel_alloc")
	if allocFn == nil {
		return 0
	}
	res, err := allocFn.Call(context.Background(), uint64(len(b)))
	if err != nil || len(res) == 0 {
		return 0
	}
	ptr := uint32(res[0])
	if !m.Memory().Write(ptr, b) {
		return 0
	}
	return ptr
}

// CompileAndCache compiles a wasm module by path; cache is shared with
// the runtime. Returns the wazero compiled module ready for instantiation.
func (r *Runtime) CompileAndCache(ctx context.Context, modulePath string) (wazero.CompiledModule, error) {
	body, err := os.ReadFile(modulePath)
	if err != nil {
		return nil, (&errors.Error{
			Component:   Component,
			Problem:     "cannot read wasm module",
			Path:        modulePath,
			Recoverable: true,
		}).Wrap(err)
	}
	cm, err := r.wzRT.CompileModule(ctx, body)
	if err != nil {
		return nil, (&errors.Error{
			Component:   Component,
			Problem:     "cannot compile wasm module",
			Path:        modulePath,
			Recoverable: true,
		}).Wrap(err)
	}
	return cm, nil
}

// CacheKey builds the cache filename pattern for a plugin module. Used
// by Install to materialize a deterministic cache path even though
// wazero's CompilationCache uses its own internal layout.
func CacheKey(name, version string, body []byte) string {
	sum := sha256.Sum256(body)
	return fmt.Sprintf("%s@%s-%s", name, version, hex.EncodeToString(sum[:])[:12])
}

// CachePath returns the on-disk path under cacheRoot for a module's
// pre-compiled artifact (used by Install for symbolic per-plugin
// directory layout — the wazero cache itself is content-addressed).
func CachePath(cacheRoot, key string) string {
	return filepath.Join(cacheRoot, key+".bin")
}
