package store

import (
	"errors"
	"strings"
	"testing"

	"github.com/AkaraChen/ctxl/core/schema"
)

func TestSelectAndTree(t *testing.T) {
	s, err := schema.LoadFile("../../examples/demo.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	st, err := Open(s, ScopeProject, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	all, err := st.Select(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 5 {
		t.Fatalf("all %d", len(all))
	}
	nodes := Tree(all)
	if len(nodes) != 5 || nodes[0].Entity != "status" || nodes[4].Entity != "note" {
		t.Fatalf("%+v", nodes)
	}
	if nodes[3].Path != "events.log" {
		t.Fatalf("path %+v", nodes[3])
	}
	picked, err := st.Select([]string{"note", "note", "log"})
	if err != nil || len(picked) != 2 || picked[0].Name != "note" || picked[1].Name != "log" {
		t.Fatalf("%v %+v", err, picked)
	}
	if _, err := st.Select([]string{"missing"}); err == nil {
		t.Fatal("expected unknown entity")
	}
}

func TestGrepIdentityAndContent(t *testing.T) {
	s, err := schema.LoadFile("../../examples/demo.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	st, err := Open(s, ScopeProject, dir)
	if err != nil {
		t.Fatal(err)
	}
	ents, err := st.Select(nil)
	if err != nil {
		t.Fatal(err)
	}
	hits, err := st.Grep(ents, "note", GrepOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !hasHit(hits, "note", MatchIdentity, "note") {
		t.Fatalf("entity name: %+v", hits)
	}
	if !hasHit(hits, "note", MatchIdentity, "docs/notes") {
		t.Fatalf("path: %+v", hits)
	}

	note, _ := st.Entity("note")
	if err := st.WriteMarkdownItem(note, "n1", Record{Fields: map[string]string{"title": "hello"}, Body: "alpha\nbeta alpha\n"}); err != nil {
		t.Fatal(err)
	}
	status, _ := st.Entity("status")
	if err := st.WriteSingular(status, Record{Fields: map[string]string{"service": "hermes", "start": "up", "stop": "down"}}); err != nil {
		t.Fatal(err)
	}
	logE, _ := st.Entity("log")
	if _, err := st.AppendNDJSON(logE, map[string]any{"result": "green", "cmd": "up"}); err != nil {
		t.Fatal(err)
	}

	hits, err = st.Grep(ents, "alpha", GrepOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var content int
	for _, h := range hits {
		if h.Entity == "note" && h.ID == "n1" && h.Kind == MatchContent {
			content++
		}
	}
	if content != 2 {
		t.Fatalf("content hits %d %+v", content, hits)
	}

	hits, err = st.Grep(ents, "n1", GrepOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !hasHit(hits, "note", MatchIdentity, "n1") {
		t.Fatalf("item id: %+v", hits)
	}

	hits, err = st.Grep(ents, "HERMES", GrepOptions{IgnoreCase: true})
	if err != nil || !hasKind(hits, MatchContent) {
		t.Fatalf("case fold %v %+v", err, hits)
	}
	hits, err = st.Grep(ents, "gr[ae]en", GrepOptions{Regexp: true})
	if err != nil || !hasHit(hits, "log", MatchContent, "") {
		t.Fatalf("regex %v %+v", err, hits)
	}
	if _, err := st.Grep(ents, "(", GrepOptions{Regexp: true}); err == nil {
		t.Fatal("expected invalid regexp")
	}
	hits, err = st.Grep(ents, "zzz-no-such", GrepOptions{})
	if err != nil || len(hits) != 0 {
		t.Fatalf("empty %v %+v", err, hits)
	}
}

func TestSelectWrongScope(t *testing.T) {
	s, err := schema.LoadFile("../../examples/demo.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	st, err := Open(s, ScopeGlobal, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	all, err := st.Select(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || all[0].Name != "note" {
		t.Fatalf("global available %+v", all)
	}
	if _, err := st.Select([]string{"status"}); err == nil {
		t.Fatal("expected locked entity")
	}
}

func TestErrNoMatchSentinel(t *testing.T) {
	if !errors.Is(ErrNoMatch, ErrNoMatch) {
		t.Fatal("sentinel")
	}
	if !strings.Contains(ErrNoMatch.Error(), "no matches") {
		t.Fatal(ErrNoMatch)
	}
}

func hasHit(hits []Hit, entity, kind, snippet string) bool {
	for _, h := range hits {
		if h.Entity != entity || h.Kind != kind {
			continue
		}
		if snippet == "" || h.Snippet == snippet || strings.Contains(h.Snippet, snippet) {
			return true
		}
	}
	return false
}

func hasKind(hits []Hit, kind string) bool {
	for _, h := range hits {
		if h.Kind == kind {
			return true
		}
	}
	return false
}
