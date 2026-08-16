package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AkaraChen/ctxl/schema"
)

func TestSingularAndNDJSON(t *testing.T) {
	s, err := schema.LoadFile("../examples/demo.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	st, err := Open(s, ScopeProject, dir)
	if err != nil {
		t.Fatal(err)
	}
	deploy, _ := st.Entity("status")
	if err := st.WriteSingular(deploy, Record{Fields: map[string]string{"service": "hermes", "port": "1", "start": "up", "stop": "down"}, Body: "ok"}); err != nil {
		t.Fatal(err)
	}
	got, err := st.ReadSingular(deploy)
	if err != nil || got.Fields["service"] != "hermes" {
		t.Fatalf("%v %+v", err, got)
	}
	if _, err := os.Stat(filepath.Join(dir, "STATUS.md")); err != nil {
		t.Fatal(err)
	}
	logE, _ := st.Entity("log")
	row, err := st.AppendNDJSON(logE, map[string]any{"result": "green", "cmd": "up"})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(row["id"]) != "1" {
		t.Fatalf("id %+v", row["id"])
	}
	all, err := st.ListNDJSON(logE)
	if err != nil || len(all) != 1 {
		t.Fatalf("%v %d", err, len(all))
	}
}

func TestMarkdownCollection(t *testing.T) {
	s, err := schema.LoadFile("../examples/demo.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	st, _ := Open(s, ScopeProject, dir)
	note, _ := st.Entity("note")
	if err := st.WriteMarkdownItem(note, "n1", Record{Fields: map[string]string{"title": "hello"}, Body: "body"}); err != nil {
		t.Fatal(err)
	}
	ids, err := st.ListMarkdownItems(note)
	if err != nil || len(ids) != 1 || ids[0] != "n1" {
		t.Fatalf("%v %v", err, ids)
	}
	got, err := st.ReadMarkdownItem(note, "n1")
	if err != nil || got.Fields["title"] != "hello" {
		t.Fatalf("%v %+v", err, got)
	}
}


func TestLastGreenFilledWhenEmpty(t *testing.T) {
	s, err := schema.LoadFile("../examples/demo.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	st, err := Open(s, ScopeProject, dir)
	if err != nil {
		t.Fatal(err)
	}
	deploy, _ := st.Entity("status")
	if err := st.WriteSingular(deploy, Record{Fields: map[string]string{"service": "hermes", "start": "up", "stop": "down", "last_green": ""}}); err != nil {
		t.Fatal(err)
	}
	got, err := st.ReadSingular(deploy)
	if err != nil {
		t.Fatal(err)
	}
	if got.Fields["last_green"] == "" {
		t.Fatal("last_green empty")
	}
	pub := got.Public()
	if _, ok := pub["Fields"]; ok {
		t.Fatal("Public still wraps Fields")
	}
	if pub["service"] != "hermes" {
		t.Fatalf("public %+v", pub)
	}
}

func TestFixedRowsDropsObjectFields(t *testing.T) {
	s, err := schema.LoadFile("../examples/demo.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	logE, err := s.Entity("log")
	if err != nil {
		t.Fatal(err)
	}
	rows := []map[string]any{{
		"id": 1, "ts": "t", "project": "p", "result": "green", "cmd": "up",
		"custom_data": map[string]any{"k": "v"},
	}}
	slim := FixedRows(logE, rows)
	if len(slim) != 1 {
		t.Fatalf("len %d", len(slim))
	}
	if _, ok := slim[0]["custom_data"]; ok {
		t.Fatal("custom_data still present")
	}
	if slim[0]["result"] != "green" {
		t.Fatalf("%+v", slim[0])
	}
}

func TestSectionWriteAndSymlink(t *testing.T) {
	s, err := schema.LoadFile("../examples/demo.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	st, err := Open(s, ScopeProject, dir)
	if err != nil {
		t.Fatal(err)
	}
	guide, _ := st.Entity("guide")
	if err := st.WriteSingular(guide, Record{Body: "first"}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "GUIDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(raw); !containsAll(got, "# Context", "first") {
		t.Fatalf("created: %s", got)
	}
	if err := os.WriteFile(filepath.Join(dir, "GUIDE.md"), []byte("# Intro\n\nkeep\n\n```\n# Context\nnot a heading\n```\n\n# Context\n\nold\n\n# Other\n\nstay\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := st.WriteSingular(guide, Record{Body: "installed"}); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile(filepath.Join(dir, "GUIDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(raw)
	if !containsAll(got, "# Intro", "keep", "not a heading", "# Context", "installed", "# Other", "stay") {
		t.Fatalf("section: %s", got)
	}
	if strings.Contains(got, "## Context") {
		t.Fatalf("heading level changed: %s", got)
	}
	if strings.Contains(got, "old") {
		t.Fatalf("old body remained: %s", got)
	}
	alias, _ := st.Entity("alias")
	if err := st.WriteSingular(alias, Record{}); err != nil {
		t.Fatal(err)
	}
	if err := st.WriteSingular(alias, Record{}); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "ALIAS.md")
	dest, err := os.Readlink(link)
	if err != nil {
		t.Fatal(err)
	}
	if dest != "GUIDE.md" {
		t.Fatalf("link %s", dest)
	}
}

func TestRootMarkdownCollection(t *testing.T) {
	s, err := schema.LoadFile("../examples/demo.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	st, _ := Open(s, ScopeProject, dir)
	note, _ := st.Entity("note")
	if err := st.WriteMarkdownItem(note, "n1", Record{Fields: map[string]string{"title": "hello"}, Body: "body"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "docs", "notes", "n1.md")); err != nil {
		t.Fatal(err)
	}
}

func TestSymlinkMissingTarget(t *testing.T) {
	s, err := schema.LoadFile("../examples/demo.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	st, _ := Open(s, ScopeProject, dir)
	alias, _ := st.Entity("alias")
	if err := st.WriteSingular(alias, Record{}); err == nil {
		t.Fatal("expected missing target")
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !stringContains(s, p) {
			return false
		}
	}
	return true
}

func stringContains(s, sub string) bool {
	return strings.Contains(s, sub)
}

func TestInitCreatesThenSkips(t *testing.T) {
	s, err := schema.LoadFile("../examples/demo.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	st, err := Open(s, ScopeProject, dir)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := st.Init(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 5 {
		t.Fatalf("rows %d", len(rows))
	}
	for _, r := range rows {
		if r.Action != "created" {
			t.Fatalf("%+v", r)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "STATUS.md")); err != nil {
		t.Fatal(err)
	}
	guide, err := os.ReadFile(filepath.Join(dir, "GUIDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(guide), "# Context") || !strings.Contains(string(guide), "Describe the current project here.") {
		t.Fatalf("guide %s", guide)
	}
	dest, err := os.Readlink(filepath.Join(dir, "ALIAS.md"))
	if err != nil || dest != "GUIDE.md" {
		t.Fatalf("alias %s %v", dest, err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".demo", "events.log")); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(filepath.Join(dir, "docs", "notes")); err != nil || !info.IsDir() {
		t.Fatalf("notes dir %v", err)
	}
	again, err := st.Init(false)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range again {
		if r.Action != "skipped" {
			t.Fatalf("second pass %+v", r)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "GUIDE.md"), []byte("# Context\n\nkeep me\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Init(false); err != nil {
		t.Fatal(err)
	}
	kept, _ := os.ReadFile(filepath.Join(dir, "GUIDE.md"))
	if !strings.Contains(string(kept), "keep me") {
		t.Fatalf("overwrote without force: %s", kept)
	}
}
