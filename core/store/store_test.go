package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AkaraChen/ctxl/core/schema"
	"github.com/AkaraChen/ctxl/core/schema/loader"
)

func TestSingularAndNDJSONStores(t *testing.T) {
	s := storeTestSchema(t)
	root := t.TempDir()
	st, err := Open(s, schema.ScopeProject, root)
	if err != nil {
		t.Fatal(err)
	}
	status := entity(t, s, "status")
	if err := st.WriteSingular(status, Record{Fields: map[string]string{"service": "api"}, Body: "healthy"}); err != nil {
		t.Fatal(err)
	}
	record, err := st.ReadSingular(status)
	if err != nil {
		t.Fatal(err)
	}
	if record.Fields["service"] != "api" || record.Body != "healthy" {
		t.Fatalf("record = %+v", record)
	}
	public := record.Public()
	if public["service"] != "api" || public["body"] != "healthy" {
		t.Fatalf("public record = %+v", public)
	}

	events := entity(t, s, "events")
	row, err := st.AppendNDJSON(events, map[string]any{"result": "green", "details": map[string]any{"attempt": 1}})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(row["id"]) != "1" || row["ts"] == "" {
		t.Fatalf("generated fields = %+v", row)
	}
	got, err := st.GetNDJSON(events, "1")
	if err != nil || got["result"] != "green" {
		t.Fatalf("get = %+v, err=%v", got, err)
	}
	rows, err := st.ListNDJSON(events)
	if err != nil || len(rows) != 1 {
		t.Fatalf("list = %+v, err=%v", rows, err)
	}
}

func TestMarkdownCollectionLifecycle(t *testing.T) {
	s := storeTestSchema(t)
	root := t.TempDir()
	st, err := Open(s, schema.ScopeProject, root)
	if err != nil {
		t.Fatal(err)
	}
	notes := entity(t, s, "notes")
	if err := st.WriteMarkdownItem(notes, "n1", Record{Fields: map[string]string{"title": "first"}, Body: "body"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "notes", "n1.md")); err != nil {
		t.Fatal(err)
	}
	if err := st.WriteMarkdownItem(notes, "n1", Record{Fields: map[string]string{"title": "updated"}, Body: "new body"}); err != nil {
		t.Fatal(err)
	}
	ids, err := st.ListMarkdownItems(notes)
	if err != nil || len(ids) != 1 || ids[0] != "n1" {
		t.Fatalf("list = %v, err=%v", ids, err)
	}
	record, err := st.ReadMarkdownItem(notes, "n1")
	if err != nil || record.Fields["title"] != "updated" || record.Body != "new body" {
		t.Fatalf("record = %+v, err=%v", record, err)
	}
	if err := st.DeleteMarkdownItem(notes, "n1"); err != nil {
		t.Fatal(err)
	}
	if _, err := st.ReadMarkdownItem(notes, "n1"); err == nil {
		t.Fatal("deleted note is still readable")
	}
}

func TestFixedRowsDropsObjectFields(t *testing.T) {
	events := entity(t, storeTestSchema(t), "events")
	rows := []map[string]any{{
		"id": 1, "ts": "t", "result": "green", "details": map[string]any{"attempt": 1},
	}}
	slim := FixedRows(events, rows)
	if len(slim) != 1 || slim[0]["result"] != "green" {
		t.Fatalf("fixed rows = %+v", slim)
	}
	if _, ok := slim[0]["details"]; ok {
		t.Fatalf("object field leaked: %+v", slim[0])
	}
}

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

func TestStoreOverrideAndEntityScope(t *testing.T) {
	s, err := loader.Parse([]byte(`{
	  "name":"demo",
	  "store":{"name":"custom-data"},
	  "entities":[{"name":"status","kind":"singular","format":"markdown","path":"STATUS.md","scope":"project"}]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	project, err := Open(s, schema.ScopeProject, root)
	if err != nil {
		t.Fatal(err)
	}
	dir, err := project.StoreDir()
	if err != nil {
		t.Fatal(err)
	}
	if dir != filepath.Join(root, ".custom-data") {
		t.Fatalf("store dir = %q", dir)
	}
	global, err := Open(s, schema.ScopeGlobal, root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := global.EntityPath(s.Entities[0]); err == nil {
		t.Fatal("project-only entity accepted global scope")
	}
	if _, err := Open(s, "", root); err == nil {
		t.Fatal("empty scope was accepted")
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
	    {"name":"events","kind":"plural","format":"ndjson","path":"events.ndjson","scope":"project","fields":[{"name":"id","type":"int"},{"name":"ts","type":"string"},{"name":"result","type":"string"},{"name":"details","type":"object"}]},
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
