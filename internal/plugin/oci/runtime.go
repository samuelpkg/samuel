// Package oci implements the OCI-tier plugin loader: container images
// pulled by Podman (rootless) or Docker and launched on demand with a
// fixed mount layout and a Unix-socket gRPC bridge for the framework
// hooks (see internal/plugin/oci/server.go).
//
// Detection order: Podman → Docker → SAMUEL_RUNTIME override. The
// detected runtime is cached per-process in DetectedRuntime; callers
// pass the result around explicitly so tests can stub it.
package oci

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/ar4mirez/samuel/internal/errors"
)

// Component is the structured-error namespace.
const Component = "plugin/oci"

// RuntimeKind enumerates the supported container engines.
type RuntimeKind string

const (
	RuntimePodman RuntimeKind = "podman"
	RuntimeDocker RuntimeKind = "docker"
)

// DetectedRuntime carries the resolved engine + how it was chosen.
type DetectedRuntime struct {
	Kind   RuntimeKind
	Path   string // absolute path to the CLI binary
	Reason string // "env-override" | "podman-rootless" | "docker"
}

// DetectRuntime resolves the container engine following the order
// documented in the PRD. The SAMUEL_RUNTIME env var wins outright when
// set and resolves to a valid CLI on PATH.
func DetectRuntime() (DetectedRuntime, error) {
	if override := strings.TrimSpace(os.Getenv("SAMUEL_RUNTIME")); override != "" {
		path, err := exec.LookPath(override)
		if err != nil {
			return DetectedRuntime{}, &errors.Error{
				Component:   Component,
				Problem:     "SAMUEL_RUNTIME refers to unknown binary",
				Cause:       override,
				Fix:         "unset SAMUEL_RUNTIME or set it to podman/docker",
				Recoverable: true,
			}
		}
		return DetectedRuntime{Kind: RuntimeKind(filepathBase(override)), Path: path, Reason: "env-override"}, nil
	}
	if path, err := exec.LookPath("podman"); err == nil {
		return DetectedRuntime{Kind: RuntimePodman, Path: path, Reason: "podman-rootless"}, nil
	}
	if path, err := exec.LookPath("docker"); err == nil {
		return DetectedRuntime{Kind: RuntimeDocker, Path: path, Reason: "docker"}, nil
	}
	return DetectedRuntime{}, &errors.Error{
		Component:   Component,
		Problem:     "no container runtime found",
		Fix:         "install podman or docker, or set SAMUEL_RUNTIME=<binary>",
		DocsURL:     "https://ar4mirez.github.io/samuel/docs/errors/SAM-OCI-001",
		Recoverable: true,
	}
}

// imageNameRE is the image-reference regex ported from v1
// internal/core/docker.go lines 60-75.
//
// Pattern: <registry>/<owner>/<name>[:<tag>][@<digest>]
//
//	registry: hostname[.tld][:port]
//	owner:    [a-z0-9]+(?:[._-][a-z0-9]+)*
//	name:     [a-z0-9]+(?:[._-][a-z0-9]+)*
//	tag:      [A-Za-z0-9_][A-Za-z0-9._-]{0,127}
//	digest:   sha256:[a-f0-9]{64}
var imageNameRE = regexp.MustCompile(`^(?P<registry>(?:[A-Za-z0-9][A-Za-z0-9.-]*(?::[0-9]+)?))/(?P<owner>[a-z0-9]+(?:[._-][a-z0-9]+)*)/(?P<name>[a-z0-9]+(?:[._-][a-z0-9]+)*)(?::(?P<tag>[A-Za-z0-9_][A-Za-z0-9._-]{0,127}))?(?:@(?P<digest>sha256:[a-f0-9]{64}))?$`)

// ValidateImageName runs the regex check + a couple of structural sanity
// rules (no double colons, no trailing slash). Returns the parsed parts.
type ImageRef struct {
	Registry string
	Owner    string
	Name     string
	Tag      string
	Digest   string
}

// ParseImageName validates ref and returns the structured form. Empty
// tag falls back to "latest" (matching docker pull semantics).
func ParseImageName(ref string) (ImageRef, error) {
	matches := imageNameRE.FindStringSubmatch(ref)
	if matches == nil {
		return ImageRef{}, &errors.Error{
			Component:   Component,
			Problem:     "invalid OCI image reference",
			Path:        ref,
			Fix:         "use registry/owner/name[:tag][@digest]",
			Recoverable: true,
		}
	}
	out := ImageRef{}
	for i, name := range imageNameRE.SubexpNames() {
		switch name {
		case "registry":
			out.Registry = matches[i]
		case "owner":
			out.Owner = matches[i]
		case "name":
			out.Name = matches[i]
		case "tag":
			out.Tag = matches[i]
		case "digest":
			out.Digest = matches[i]
		}
	}
	if out.Tag == "" {
		out.Tag = "latest"
	}
	return out, nil
}

// String renders the canonical "registry/owner/name:tag" form (digest
// is appended when present).
func (r ImageRef) String() string {
	s := fmt.Sprintf("%s/%s/%s:%s", r.Registry, r.Owner, r.Name, r.Tag)
	if r.Digest != "" {
		s += "@" + r.Digest
	}
	return s
}

// Engine is the per-Runtime CLI invoker. Tests inject a FakeEngine.
type Engine interface {
	// Pull pulls the image and returns the content digest.
	Pull(ctx context.Context, image string) (string, error)
	// Inspect returns the digest if the image is locally available.
	Inspect(ctx context.Context, image string) (string, error)
	// Remove deletes a local image. Returns nil if absent.
	Remove(ctx context.Context, image string) error
}

// CLI wraps the resolved runtime binary.
type CLI struct {
	rt      DetectedRuntime
	timeout time.Duration
}

// NewCLI constructs an Engine backed by the user's container runtime.
func NewCLI(rt DetectedRuntime) *CLI { return &CLI{rt: rt, timeout: 5 * time.Minute} }

// WithTimeout overrides the default 5-minute timeout (tests pin a short
// timeout; production keeps the slow default for large pulls).
func (c *CLI) WithTimeout(d time.Duration) *CLI { c.timeout = d; return c }

// Pull invokes `<runtime> pull <image>` and parses the resulting digest.
func (c *CLI) Pull(ctx context.Context, image string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.rt.Path, "pull", image)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", (&errors.Error{
			Component:   Component,
			Problem:     "image pull failed",
			Cause:       fmt.Sprintf("%s: %s", err, strings.TrimSpace(string(out))),
			Path:        image,
			Recoverable: true,
		}).Wrap(err)
	}
	return c.Inspect(ctx, image)
}

// Inspect returns the local image's content digest (sha256:...).
func (c *CLI) Inspect(ctx context.Context, image string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.rt.Path, "image", "inspect", image, "--format", "{{.Id}}")
	out, err := cmd.Output()
	if err != nil {
		return "", (&errors.Error{
			Component:   Component,
			Problem:     "image inspect failed",
			Cause:       err.Error(),
			Path:        image,
			Recoverable: true,
		}).Wrap(err)
	}
	digest := strings.TrimSpace(string(out))
	if digest == "" {
		return "", &errors.Error{
			Component:   Component,
			Problem:     "image inspect returned empty digest",
			Path:        image,
			Recoverable: true,
		}
	}
	return digest, nil
}

// Remove deletes a local image. A "not found" condition is mapped to nil.
func (c *CLI) Remove(ctx context.Context, image string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.rt.Path, "image", "rm", image)
	if out, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(out), "No such image") {
			return nil
		}
		return (&errors.Error{
			Component:   Component,
			Problem:     "image rm failed",
			Cause:       fmt.Sprintf("%s: %s", err, strings.TrimSpace(string(out))),
			Path:        image,
			Recoverable: true,
		}).Wrap(err)
	}
	return nil
}

// filepathBase is a tiny helper that avoids importing path/filepath
// only for one Base call (the runtime path is already absolute and the
// last segment is the binary name).
func filepathBase(p string) string {
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}
