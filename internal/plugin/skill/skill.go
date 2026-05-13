// Package skill implements the skill-tier plugin loader: a text-only
// Agent Skill bundle that lands as SKILL.md (+ optional scripts/,
// references/, assets/) under <project>/.samuel/plugins/<name>/.
//
// Skill plugins do not execute code. Install copies the manifest +
// SKILL.md + companion subtrees from the materialized source into the
// project. Detect tests for the on-disk SKILL.md. Check validates the
// frontmatter shape. Uninstall removes the directory.
package skill

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ar4mirez/samuel/internal/errors"
	"github.com/ar4mirez/samuel/internal/plugin"
	"github.com/ar4mirez/samuel/internal/plugin/manifest"
	"github.com/ar4mirez/samuel/internal/plugin/source"
)

// Component is the structured-error namespace.
const Component = "plugin/skill"

// SkillFile is the on-disk filename for the skill body.
const SkillFile = "SKILL.md"

// Plugin is a skill-tier plugin.Plugin implementation.
type Plugin struct {
	Manifest_  manifest.Manifest
	ProjectDir string
	// SourceDir is the materialized plugin tree (output of source.Fetcher).
	// When empty, Detect/Check/Uninstall still work; Install requires it.
	SourceDir string
}

// New constructs a skill-tier Plugin.
func New(m manifest.Manifest, projectDir, sourceDir string) *Plugin {
	return &Plugin{Manifest_: m, ProjectDir: projectDir, SourceDir: sourceDir}
}

// Name returns the manifest's plugin name.
func (p *Plugin) Name() string { return p.Manifest_.Name }

// Manifest returns the v2 framework Manifest snapshot built from the
// skill manifest.
func (p *Plugin) Manifest() plugin.Manifest {
	return plugin.Manifest{
		Name:    p.Manifest_.Name,
		Version: p.Manifest_.Version,
		Kind:    plugin.KindSkill,
		Summary: p.Manifest_.Summary,
		Source:  p.Manifest_.Homepage,
	}
}

// pluginDir returns the on-disk install location:
// <project>/.samuel/plugins/<name>/.
func (p *Plugin) pluginDir() string {
	return filepath.Join(p.ProjectDir, ".samuel", "plugins", p.Name())
}

// Detect reports installed=true when SKILL.md exists.
func (p *Plugin) Detect(_ context.Context) (plugin.DetectResult, error) {
	dir := p.pluginDir()
	skillPath := filepath.Join(dir, SkillFile)
	if _, err := os.Stat(skillPath); err != nil {
		return plugin.DetectResult{Installed: false, Path: dir}, nil
	}
	return plugin.DetectResult{Installed: true, Path: dir, Version: p.Manifest_.Version}, nil
}

// Install copies the source tree into the project's plugins directory.
//
// Atomicity: copy into a sibling tmp dir, then rename onto the target.
func (p *Plugin) Install(_ context.Context, opts plugin.InstallOptions) (plugin.InstallResult, error) {
	res := plugin.InstallResult{Component: p.Name()}
	if p.SourceDir == "" {
		return res, &errors.Error{
			Component:   Component,
			Problem:     "skill plugin has no source dir",
			Recoverable: false,
		}
	}
	skillSrc := filepath.Join(p.SourceDir, SkillFile)
	if _, err := os.Stat(skillSrc); err != nil {
		return res, (&errors.Error{
			Component:   Component,
			Problem:     "skill source missing SKILL.md",
			Path:        skillSrc,
			Recoverable: true,
		}).Wrap(err)
	}
	target := p.pluginDir()
	if opts.DryRun {
		return res, nil
	}
	parent := filepath.Dir(target)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return res, (&errors.Error{
			Component:   Component,
			Problem:     "cannot create parent dir for skill plugin",
			Path:        parent,
			Recoverable: true,
		}).Wrap(err)
	}
	tmp, err := os.MkdirTemp(parent, fmt.Sprintf(".%s.tmp-", p.Name()))
	if err != nil {
		return res, err
	}
	cleanupTmp := func() { _ = os.RemoveAll(tmp) }
	if err := source.CopyTree(p.SourceDir, tmp); err != nil {
		cleanupTmp()
		return res, err
	}
	// Drop manifest into the install dir so `samuel info` reads the
	// plugin's full toml without a registry fetch.
	if err := copyManifestIfPresent(p.SourceDir, tmp); err != nil {
		cleanupTmp()
		return res, err
	}
	var backup string
	if _, statErr := os.Stat(target); statErr == nil {
		backup = target + ".bak"
		if err := os.Rename(target, backup); err != nil {
			cleanupTmp()
			return res, (&errors.Error{
				Component:   Component,
				Problem:     "cannot move existing plugin out of the way",
				Path:        target,
				Recoverable: true,
			}).Wrap(err)
		}
	}
	if err := os.Rename(tmp, target); err != nil {
		if backup != "" {
			_ = os.Rename(backup, target)
		}
		cleanupTmp()
		return res, (&errors.Error{
			Component:   Component,
			Problem:     "cannot rename staged plugin into place",
			Path:        target,
			Recoverable: true,
		}).Wrap(err)
	}
	if backup != "" {
		_ = os.RemoveAll(backup)
	}
	res.Mutations = append(res.Mutations, plugin.Mutation{
		Kind:        plugin.MutationDirCreated,
		Path:        target,
		Description: "installed skill " + p.Name(),
		Reverse: func(context.Context) error {
			return os.RemoveAll(target)
		},
	})
	return res, nil
}

func copyManifestIfPresent(src, dst string) error {
	srcManifest := filepath.Join(src, manifest.FileName)
	if _, err := os.Stat(srcManifest); err != nil {
		return nil // not all sources carry the manifest at root
	}
	return source.CopyTree(srcManifest, filepath.Join(dst, manifest.FileName))
}

// Check validates SKILL.md frontmatter shape.
func (p *Plugin) Check(_ context.Context) plugin.HealthStatus {
	dir := p.pluginDir()
	skillPath := filepath.Join(dir, SkillFile)
	body, err := os.ReadFile(skillPath)
	if err != nil {
		return plugin.HealthStatus{
			Component: p.Name(),
			OK:        false,
			Message:   "SKILL.md not found at " + skillPath,
			FixHint:   "samuel install " + p.Name(),
		}
	}
	if err := ValidateFrontmatter(body); err != nil {
		return plugin.HealthStatus{
			Component: p.Name(),
			OK:        false,
			Message:   err.Error(),
			FixHint:   "fix SKILL.md frontmatter (name + description required)",
		}
	}
	return plugin.HealthStatus{Component: p.Name(), OK: true, Message: "skill " + p.Name() + " healthy"}
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
		return res, (&errors.Error{
			Component:   Component,
			Problem:     "cannot remove skill plugin directory",
			Path:        dir,
			Recoverable: true,
		}).Wrap(err)
	}
	res.Mutations = append(res.Mutations, plugin.Mutation{
		Kind:        plugin.MutationDirCreated,
		Path:        dir,
		Description: "removed skill " + p.Name(),
	})
	return res, nil
}

// ValidateFrontmatter requires a name+description pair in YAML
// frontmatter (--- block at the top of SKILL.md). The Anthropic Agent
// Skills spec is intentionally light; we only enforce the two
// load-bearing fields.
func ValidateFrontmatter(body []byte) error {
	s := string(body)
	if !strings.HasPrefix(s, "---") {
		return fmt.Errorf("SKILL.md missing YAML frontmatter")
	}
	end := strings.Index(s[3:], "---")
	if end < 0 {
		return fmt.Errorf("SKILL.md frontmatter is not closed")
	}
	front := s[3 : 3+end]
	hasName := false
	hasDesc := false
	for _, line := range strings.Split(front, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "name:"):
			if strings.TrimSpace(strings.TrimPrefix(line, "name:")) != "" {
				hasName = true
			}
		case strings.HasPrefix(line, "description:"):
			if strings.TrimSpace(strings.TrimPrefix(line, "description:")) != "" {
				hasDesc = true
			}
		}
	}
	if !hasName || !hasDesc {
		return fmt.Errorf("SKILL.md frontmatter missing required fields (name, description)")
	}
	return nil
}

// Compile-time guarantee.
var _ plugin.Plugin = (*Plugin)(nil)
