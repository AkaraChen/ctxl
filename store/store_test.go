package store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/AkaraChen/ctxlayer/schema"
)

func TestSingularAndNDJSON(t *testing.T) {
	s, err := schema.LoadFile("../examples/yoi.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	st, err := Open(s, ScopeProject, dir)
	if err != nil {
		t.Fatal(err)
	}
	deploy, _ := st.Entity("deploy")
	if err := st.WriteSingular(deploy, Record{Fields: map[string]string{"service": "hermes", "port": "1", "start": "up", "stop": "down"}, Body: "ok"}); err != nil {
		t.Fatal(err)
	}
	got, err := st.ReadSingular(deploy)
	if err != nil || got.Fields["service"] != "hermes" {
		t.Fatalf("%v %+v", err, got)
	}
	if _, err := os.Stat(filepath.Join(dir, "DEPLOY.md")); err != nil {
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
	s, err := schema.LoadFile("../examples/notes.schema.json")
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

