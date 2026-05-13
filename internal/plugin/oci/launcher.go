package oci

import (
	"context"
	"fmt"
	"os/exec"
	"os/user"
	"strings"

	"github.com/ar4mirez/samuel/internal/plugin/capability"
)

// LaunchOptions parameterizes a `<runtime> run` invocation. The
// orchestration that builds these from a plugin Manifest lives in the
// install path; the launcher itself is policy-free.
type LaunchOptions struct {
	Image  string
	Layout MountLayout
	// Capability grants the plugin received at install time. Drives
	// workspace writability + network policy.
	Grants []capability.Grant
	// EnvAllowlist names environment variables the framework may
	// forward into the container. Anything else is stripped.
	EnvAllowlist []string
	// HostEnv is the snapshot of the host environment the launcher
	// filters against EnvAllowlist. Tests pin this; production passes
	// os.Environ().
	HostEnv []string
	// Command overrides the image entrypoint when set.
	Command []string
}

// BuildRunArgs assembles the argument list for `<runtime> run`. The
// returned slice excludes the runtime binary itself.
//
// Layout:
//   --rm --user UID:GID
//   -v <Workspace>:/workspace[:ro]
//   -v <Skills>:/skills:ro
//   -v <SamuelRun>:/.samuel/run[:ro]
//   -v <PluginConfig>:/plugin/config:ro
//   -v <BridgeSocket>:/samuel-bridge
//   --network <policy>
//   -e KEY=VALUE  (one per allowed env var)
//   <image> [command...]
func BuildRunArgs(opts LaunchOptions) []string {
	args := []string{"run", "--rm"}
	if uid, gid := currentUIDGID(); uid != "" {
		args = append(args, "--user", uid+":"+gid)
	}
	workspaceMode := ":ro"
	if hasWriteCapability(opts.Grants) {
		workspaceMode = ""
	}
	samuelRunMode := workspaceMode
	if opts.Layout.Workspace != "" {
		args = append(args, "-v", opts.Layout.Workspace+":/workspace"+workspaceMode)
	}
	if opts.Layout.Skills != "" {
		args = append(args, "-v", opts.Layout.Skills+":/skills:ro")
	}
	if opts.Layout.SamuelRun != "" {
		args = append(args, "-v", opts.Layout.SamuelRun+":/.samuel/run"+samuelRunMode)
	}
	if opts.Layout.PluginConfig != "" {
		args = append(args, "-v", opts.Layout.PluginConfig+":/plugin/config:ro")
	}
	if opts.Layout.BridgeSocket != "" {
		args = append(args, "-v", opts.Layout.BridgeSocket+":/samuel-bridge")
	}
	args = append(args, "--network", networkPolicy(opts.Grants))

	for _, kv := range filterEnv(opts.HostEnv, opts.EnvAllowlist) {
		args = append(args, "-e", kv)
	}

	args = append(args, opts.Image)
	args = append(args, opts.Command...)
	return args
}

// Launch shells out to the runtime CLI with the prepared run args. The
// process is started in the background; the bridge handles all
// further communication. Callers can context.WithCancel to terminate.
func Launch(ctx context.Context, runtime DetectedRuntime, opts LaunchOptions) (*exec.Cmd, error) {
	args := BuildRunArgs(opts)
	cmd := exec.CommandContext(ctx, runtime.Path, args...)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("oci: launch %s: %w", opts.Image, err)
	}
	return cmd, nil
}

// hasWriteCapability returns true when the grants contain any
// filesystem.write entry pointing at /workspace.
func hasWriteCapability(grants []capability.Grant) bool {
	for _, g := range grants {
		if g.Kind != capability.KindFilesystemWrite {
			continue
		}
		for _, t := range g.Targets {
			if t == "/workspace" || strings.HasPrefix(t, "/workspace/") {
				return true
			}
		}
	}
	return false
}

// networkPolicy maps the network.outbound capability to a runtime
// network flag value. The PRD calls for deny-by-default; only when the
// plugin explicitly requests outbound do we pass through `bridge`.
//
// The runtime-level allowlist (per-destination filter) is enforced by
// `--add-host` rules + iptables in production; v2.0 ships the binary
// gate (none / bridge) and the bridge.MatchHost host-function check
// handles per-call filtering.
func networkPolicy(grants []capability.Grant) string {
	for _, g := range grants {
		if g.Kind == capability.KindNetworkOutbound && len(g.Targets) > 0 {
			return "bridge"
		}
	}
	return "none"
}

// filterEnv returns key=value pairs from env where the key is in allow.
func filterEnv(env, allow []string) []string {
	if len(allow) == 0 {
		return nil
	}
	allowSet := make(map[string]struct{}, len(allow))
	for _, a := range allow {
		allowSet[a] = struct{}{}
	}
	var out []string
	for _, kv := range env {
		i := strings.IndexByte(kv, '=')
		if i < 0 {
			continue
		}
		if _, ok := allowSet[kv[:i]]; ok {
			out = append(out, kv)
		}
	}
	return out
}

func currentUIDGID() (string, string) {
	u, err := user.Current()
	if err != nil {
		return "", ""
	}
	return u.Uid, u.Gid
}
