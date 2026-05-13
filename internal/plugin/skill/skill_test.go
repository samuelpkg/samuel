package skill

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelpkg/samuel/internal/plugin"
	"github.com/samuelpkg/samuel/internal/plugin/manifest"
)

// makeFixtureSource writes a minimal skill plugin fixture into dir and
// returns the path.
func makeFixtureSource(t *testing.T, name, version string) string {
	t.Helper()
	dir := t.TempDir()
	body := `---
name: ` + name + `
description: A test skill plugin used in unit tests.
---

# Hello

Sample skill body.
`
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	tomlBody := []byte("name = \"" + name + "\"\nversion = \"" + version + "\"\nkind = \"skill\"\n")
	if err := os.WriteFile(filepath.Join(dir, manifest.FileName), tomlBody, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestSkill_InstallAndDetect(t *testing.T) {
	src := makeFixtureSource(t, "go-guide", "1.0.0")
	project := t.TempDir()
	m := manifest.Manifest{Name: "go-guide", Version: "1.0.0", Kind: manifest.KindSkill}
	p := New(m, project, src)

	res, err := p.Install(context.Background(), plugin.InstallOptions{})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if len(res.Mutations) != 1 {
		t.Errorf("expected one mutation, got %d", len(res.Mutations))
	}
	det, _ := p.Detect(context.Background())
	if !det.Installed {
		t.Errorf("Detect should report installed")
	}
}

func TestSkill_InstallIsAtomicOnFailure(t *testing.T) {
	src := t.TempDir()
	// no SKILL.md → install must fail
	project := t.TempDir()
	m := manifest.Manifest{Name: "broken", Version: "1.0.0", Kind: manifest.KindSkill}
	p := New(m, project, src)

	if _, err := p.Install(context.Background(), plugin.InstallOptions{}); err == nil {
		t.Errorf("expected install to fail without SKILL.md")
	}
	if _, err := os.Stat(p.pluginDir()); err == nil {
		t.Errorf("plugin dir should not exist after failed install")
	}
}

func TestSkill_Check_InvalidFrontmatter(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("no frontmatter"), 0o644); err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	m := manifest.Manifest{Name: "nofm", Version: "1.0.0", Kind: manifest.KindSkill}
	p := New(m, project, src)
	if _, err := p.Install(context.Background(), plugin.InstallOptions{}); err != nil {
		t.Fatalf("Install: %v", err)
	}
	st := p.Check(context.Background())
	if st.OK {
		t.Errorf("Check should fail when frontmatter is missing")
	}
}

func TestSkill_Uninstall(t *testing.T) {
	src := makeFixtureSource(t, "go-guide", "1.0.0")
	project := t.TempDir()
	p := New(manifest.Manifest{Name: "go-guide", Version: "1.0.0", Kind: manifest.KindSkill}, project, src)
	if _, err := p.Install(context.Background(), plugin.InstallOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Uninstall(context.Background(), plugin.UninstallOptions{}); err != nil {
		t.Fatalf("Uninstall: %v", err)
	}
	if _, err := os.Stat(p.pluginDir()); !os.IsNotExist(err) {
		t.Errorf("plugin dir should be gone")
	}
}

func TestValidateFrontmatter(t *testing.T) {
	good := []byte("---\nname: go-guide\ndescription: hi\n---\nbody\n")
	if err := ValidateFrontmatter(good); err != nil {
		t.Errorf("good frontmatter rejected: %v", err)
	}
	bad := []byte("---\nname: go-guide\n---\nbody\n")
	if err := ValidateFrontmatter(bad); err == nil {
		t.Errorf("missing description should fail")
	}
}
