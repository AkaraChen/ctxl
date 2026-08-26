package skillsgen

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AkaraChen/ctxl/core/schema"
	"github.com/AkaraChen/ctxl/core/schema/loader"
	"github.com/AkaraChen/ctxl/core/skillbundle"
)

func TestBuildBuiltinInjectionAndCustomSkills(t *testing.T) {
	root := t.TempDir()
	builtinSource := []byte("---\nname: base-skill\ndescription: Base instructions.\nmetadata:\n  owner: eric\n---\n\n# User body\n\nRead references/base.md.\n")
	customSource := []byte("---\nname: custom-one\ndescription: A custom skill.\n---\n\n# Exact custom body\n")
	writeFixture(t, root, "base-skill/SKILL.md", builtinSource, 0o644)
	writeFixture(t, root, "base-skill/references/base.md", []byte("base reference\n"), 0o644)
	writeFixture(t, root, "custom-one/SKILL.md", customSource, 0o644)
	writeFixture(t, root, "custom-one/scripts/run.sh", []byte("#!/bin/sh\necho custom\n"), 0o755)
	writeFixture(t, root, "custom-one/assets/blob.bin", []byte{0, 1, 2, 255}, 0o644)
	customTwoSource := []byte("---\nname: custom-two\ndescription: Another custom skill.\n---\n\nSECOND CUSTOM BODY\n")
	writeFixture(t, root, "custom-two/SKILL.md", customTwoSource, 0o644)

	s := mustSchema(t, `{
	  "name":"demo",
	  "cli":{"name":"democtl"},
	  "store":{"name":"demo-data"},
	  "skills":[
	    {"type":"builtin","directory":"base-skill","inject":"before","name":"demo-agent","description":"Generated demo instructions."},
	    {"type":"custom","directory":"custom-one"},
	    {"type":"custom","directory":"custom-two"}
	  ],
	  "entities":[{"name":"status","command":{"name":"current"},"kind":"singular","format":"markdown","path":"STATUS.md"}]
	}`)
	bundle, err := Build(s, root)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(bundle.Names(), ","); got != "demo-agent,custom-one,custom-two" {
		t.Fatalf("names = %s", got)
	}
	builtin, err := bundle.Markdown("demo-agent")
	if err != nil {
		t.Fatal(err)
	}
	generatedAt := bytes.Index(builtin, []byte("<!-- ctxl:generated:start -->"))
	userAt := bytes.Index(builtin, []byte("# User body"))
	if generatedAt < 0 || userAt < 0 || generatedAt > userAt {
		t.Fatalf("before injection order is wrong:\n%s", builtin)
	}
	for _, want := range []string{"name: demo-agent", "description: Generated demo instructions.", "democtl current show", ".demo-data/", "democtl skills path NAME"} {
		if !bytes.Contains(builtin, []byte(want)) {
			t.Fatalf("builtin missing %q:\n%s", want, builtin)
		}
	}
	custom := bundledSkill(t, bundle, "custom-one")
	assertEntry(t, custom.Entries, "SKILL.md", customSource, 0o644)
	assertEntry(t, custom.Entries, "scripts/run.sh", []byte("#!/bin/sh\necho custom\n"), 0o755)
	assertEntry(t, custom.Entries, "assets/blob.bin", []byte{0, 1, 2, 255}, 0o644)
	customTwo, err := bundle.Markdown("custom-two")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(customTwo, customTwoSource) {
		t.Fatalf("second custom SKILL.md was rewritten:\n%s", customTwo)
	}
}

func TestBuiltinAfterInjection(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "base-skill/SKILL.md", []byte("---\nname: base-skill\ndescription: Base instructions.\n---\n\nUSER BODY\n"), 0o644)
	s := mustSchema(t, `{
	  "name":"demo",
	  "skills":[{"type":"builtin","directory":"base-skill","inject":"after","name":"demo-skill"}],
	  "entities":[{"name":"status","kind":"singular","format":"markdown","path":"STATUS.md"}]
	}`)
	bundle, err := Build(s, root)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := bundle.Markdown("demo-skill")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Index(raw, []byte("USER BODY")) > bytes.Index(raw, []byte("<!-- ctxl:generated:start -->")) {
		t.Fatalf("after injection order is wrong:\n%s", raw)
	}
	if !bytes.HasPrefix(raw, []byte("---\n")) {
		t.Fatalf("frontmatter is not first:\n%s", raw)
	}
}

func TestBuildRejectsInvalidSkillDirectories(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "wrong-dir/SKILL.md", []byte("---\nname: another-name\ndescription: Mismatch.\n---\n"), 0o644)
	s := mustSchema(t, `{
	  "name":"demo",
	  "skills":[{"type":"custom","directory":"wrong-dir"}],
	  "entities":[{"name":"status","kind":"singular","format":"markdown","path":"STATUS.md"}]
	}`)
	if _, err := Build(s, root); err == nil || !strings.Contains(err.Error(), "must match directory name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustSchema(t *testing.T, raw string) schema.Schema {
	t.Helper()
	s, err := loader.Parse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func writeFixture(t *testing.T, root, rel string, data []byte, mode os.FileMode) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, mode); err != nil {
		t.Fatal(err)
	}
}

func assertEntry(t *testing.T, entries []skillbundle.Entry, path string, data []byte, mode uint32) {
	t.Helper()
	for _, entry := range entries {
		if entry.Path == path {
			if !bytes.Equal(entry.Data, data) || entry.Mode != mode {
				t.Fatalf("entry %s mode/data differ: mode=%o data=%v", path, entry.Mode, entry.Data)
			}
			return
		}
	}
	t.Fatalf("missing entry %s", path)
}

func bundledSkill(t *testing.T, bundle skillbundle.Bundle, name string) skillbundle.Skill {
	t.Helper()
	for _, skill := range bundle.Skills {
		if skill.Name == name {
			return skill
		}
	}
	t.Fatalf("missing bundled Skill %q", name)
	return skillbundle.Skill{}
}
