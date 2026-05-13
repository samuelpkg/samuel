# TinyGo + WASM

WASM-tier plugins are TinyGo modules executed inside Samuel by [wazero](https://wazero.io). The runtime is pure Go — no host toolchain, no shared libraries, no platform-specific binaries to ship.

## Toolchain

You need:

- **TinyGo** ≥ 0.31 — `brew install tinygo` or [download](https://tinygo.org/getting-started/install/).
- **Go** ≥ 1.22 (TinyGo's host).
- A working `samuel` install for local validate.

## Project layout

```text
samuel-my-translator/
├── samuel-plugin.toml
├── go.mod
├── plugin.go            # main package, host-exported functions
├── plugin_test.go
└── .github/workflows/
    └── release.yml      # uses samuelpkg/samuel-plugin-release@v1
```

## Minimum module

```go
package main

import (
    "unsafe"

    samuel "github.com/samuelpkg/samuel-wasm-sdk"
)

//export samuel_protocol_version
func samuel_protocol_version() int32 { return 1 }

//export health
func health() int32 {
    return 1 // 1 = healthy, 0 = unhealthy
}

//export sync_after
func sync_after(payloadPtr, payloadLen uint32) (resultPtr uint64) {
    payload := samuel.ReadBytes(payloadPtr, payloadLen)
    var in samuel.SyncAfterIn
    samuel.UnmarshalJSON(payload, &in)

    for _, f := range in.FilesWritten {
        if !samuel.HasSuffix(f, "AGENTS.md") {
            continue
        }
        body, _ := samuel.FsRead(f)
        target := samuel.ReplaceSuffix(f, "AGENTS.md", "CLAUDE.md")
        samuel.FsWrite(target, body)
    }
    return samuel.WriteJSON(samuel.OK)
}

func main() {} // required by TinyGo, never executed
```

## Build

```bash
tinygo build -o plugin.wasm -target=wasi ./...
```

The `wasi` target is required — Samuel runs modules under wazero with the WASI snapshot 1 surface plus the `samuel.*` host module.

## Host ABI

Samuel exposes a host module named `samuel` with these calls:

| Function | Purpose | Capability |
| --- | --- | --- |
| `samuel.log(level, ptr, len)` | structured log line | none |
| `samuel.fs_read(path_ptr, path_len) -> (out_ptr, out_len, err)` | read a file | `filesystem.read` |
| `samuel.fs_write(path_ptr, path_len, body_ptr, body_len) -> err` | write a file | `filesystem.write` |
| `samuel.exec(argv_ptr, argv_len, env_ptr, env_len) -> (exit, out_ptr, out_len)` | shell out | `exec` |
| `samuel.net_outbound(req_ptr, req_len) -> (resp_ptr, resp_len)` | HTTP(S) | `network.outbound` |
| `samuel.config_get(key_ptr, key_len) -> (val_ptr, val_len)` | read `samuel.toml` slot | `samuel.api` |
| `samuel.callback(event_ptr, event_len, payload_ptr, payload_len) -> err` | fire a sub-event | varies |

The SDK at [`samuelpkg/samuel-wasm-sdk`](https://github.com/samuelpkg/samuel-wasm-sdk) wraps these in Go-idiomatic helpers so plugins don't deal with raw pointers.

## Restrictions

TinyGo's WASI target imposes:

- No goroutines (the runtime is single-threaded).
- No `reflect` beyond what TinyGo's subset allows.
- No `cgo`.
- Limited stdlib (notably `net/http` is replaced by `samuel.net_outbound`).
- `time.Now()` and friends work; `time.Sleep` blocks the runtime — avoid.

Stick to pure logic, file I/O via host calls, and JSON over the SDK helpers. Anything that wants threads or sockets is OCI-tier territory.

## Local validate

```bash
samuel plugin validate samuel-plugin.toml
samuel install file://./   # installs from local checkout for testing
samuel doctor
```

## Compile cache

The first install of a WASM plugin compiles the module via wazero and stores the result at `~/.samuel/cache/wasm-compiled/<sha256>/`. Subsequent installs and runs hit the cache. The cache is per-process for the running framework version — bumping `samuel` invalidates everything.
