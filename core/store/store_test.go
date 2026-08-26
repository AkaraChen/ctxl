package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AkaraChen/ctxl/core/schema"
	"github.com/AkaraChen/ctxl/core/schema/loader"
)

// Section replacement and symlink behavior are not covered by the codegen
// end-to-end test; init ties both together.
func TestSectionAndSymlinkStores(t *testing.T) {
	s := storeTestSchema(t)
	root := t.TempDir()
	st, err := Open(s, schema.ScopeProject, root)
	if err != nil {
		t.Fatal(err)
	}
	guide := entity(t, s, "guide")
	if err := st.WriteSingular(guide, Record{Body: "first"}); err != nil {
		t.Fatal(err)
	}
	guidePath := filepath.Join(root, "GUIDE.md")
	if err := os.WriteFile(guidePath, []byte("# Intro\n\nkeep\n\n```\n# Context\nnot a heading\n```\n\n# Context\n\nold\n\n# Other\n\nstay\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := st.WriteSingular(guide, Record{Body: "installed"}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(guidePath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(raw)
	for _, want := range []string{"# Intro", "keep", "not a heading", "# Context", "installed", "# Other", "stay"} {
		if !strings.Contains(content, want) {
			t.Fatalf("section missing %q:\n%s", want, content)
		}
	}
	if strings.Contains(content, "## Context") || strings.Contains(content, "\nold\n") {
		t.Fatalf("section replacement changed the wrong content:\n%s", content)
	}

	alias := entity(t, s, "alias")
	if err := st.WriteSingular(alias, Record{}); err != nil {
		t.Fatal(err)
	}
	if err := st.WriteSingular(alias, Record{}); err != nil {
		t.Fatalf("idempotent symlink write: %v", err)
	}
	record, err := st.ReadSingular(alias)
	if err != nil || record.Fields["target"] != "GUIDE.md" {
		t.Fatalf("symlink record = %+v, err=%v", record, err)
	}
}

func TestSymlinkRequiresExistingTarget(t *testing.T) {
	s := storeTestSchema(t)
	st, err := Open(s, schema.ScopeProject, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := st.WriteSingular(entity(t, s, "alias"), Record{}); err == nil {
		t.Fatal("expected missing target error")
	}
}

func TestInitCreatesAndPreservesDeclaredPaths(t *testing.T) {
	s := storeTestSchema(t)
	root := t.TempDir()
	st, err := Open(s, schema.ScopeProject, root)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := st.Init(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != len(s.Entities) {
		t.Fatalf("initialized %d paths, want %d", len(rows), len(s.Entities))
	}
	for _, row := range rows {
		if row.Action != "created" {
			t.Fatalf("first init = %+v", rows)
		}
	}
	guidePath := filepath.Join(root, "GUIDE.md")
	if err := os.WriteFile(guidePath, []byte("# Context\n\nkeep me\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	again, err := st.Init(false)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range again {
		if row.Action != "skipped" {
			t.Fatalf("second init = %+v", again)
		}
	}
	kept, err := os.ReadFile(guidePath)
	if err != nil || !strings.Contains(string(kept), "keep me") {
		t.Fatalf("init overwrote existing file: %q, err=%v", kept, err)
	}
}

func storeTestSchema(t *testing.T) schema.Schema {
	t.Helper()
	s, err := loader.Parse([]byte(`{
	  "name":"store-test",
	  "entities":[
	    {"name":"status","kind":"singular","format":"markdown","path":"STATUS.md","location":"root","scope":"project","fields":[{"name":"service","type":"string"}]},
	    {"name":"guide","kind":"singular","format":"markdown","write":"section","section":"Context","body":"initial","path":"GUIDE.md","location":"root","scope":"project"},
	    {"name":"alias","kind":"singular","format":"symlink","path":"ALIAS.md","target":"GUIDE.md","location":"root","scope":"project"},
	    {"name":"events","kind":"plural","format":"ndjson","path":"events.ndjson","scope":"project","fields":[{"name":"result","type":"string"}]},
	    {"name":"notes","kind":"plural","format":"markdown","path":"docs/notes","location":"root","scope":"project","fields":[{"name":"id","type":"string"},{"name":"title","type":"string"}]}
	  ]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func entity(t *testing.T, s schema.Schema, name string) schema.Entity {
	t.Helper()
	for _, item := range s.Entities {
		if item.Name == name {
			return item
		}
	}
	t.Fatalf("missing test entity %q", name)
	return schema.Entity{}
}
