package store

import (
	"fmt"
	"os"
	"path/filepath"
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
