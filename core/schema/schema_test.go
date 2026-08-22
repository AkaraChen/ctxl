package schema

import "testing"

func TestParseDemoExample(t *testing.T) {
	raw := []byte(`{
	  "name": "demo",
	  "backend": {"type":"filesystem"},
	  "entities": [
	    {"name":"status","kind":"singular","format":"markdown","path":"STATUS.md","location":"root","fields":[{"name":"service","type":"string","required":true}]},
	    {"name":"log","kind":"plural","format":"ndjson","path":"events.log","fields":[{"name":"result","type":"string","required":true}]}
	  ]
	}`)
	s, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "demo" || len(s.Entities) != 2 {
		t.Fatalf("%+v", s)
	}
}

func TestRejectBadKind(t *testing.T) {
	_, err := Parse([]byte(`{"name":"x","backend":{"type":"filesystem"},"entities":[{"name":"a","kind":"maybe","format":"markdown","path":"a.md"}]}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSectionAndSymlink(t *testing.T) {
	s, err := Parse([]byte(`{
	  "name": "demo",
	  "backend": {"type":"filesystem"},
	  "entities": [
	    {"name":"guide","kind":"singular","format":"markdown","write":"section","section":"Context","path":"GUIDE.md"},
	    {"name":"alias","kind":"singular","format":"symlink","path":"ALIAS.md","target":"GUIDE.md"}
	  ]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if s.Entities[0].ResolvedWrite() != WriteSection {
		t.Fatal(s.Entities[0].Write)
	}
}

func TestRejectSectionWithoutHeading(t *testing.T) {
	_, err := Parse([]byte(`{"name":"x","backend":{"type":"filesystem"},"entities":[{"name":"a","kind":"singular","format":"markdown","write":"section","path":"a.md"}]}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRequireFilesystemBackend(t *testing.T) {
	if _, err := Parse([]byte(`{"name":"x","entities":[{"name":"a","kind":"singular","format":"markdown","path":"a.md"}]}`)); err == nil {
		t.Fatal("expected missing backend.type")
	}
	if _, err := Parse([]byte(`{"name":"x","backend":{"type":"mongodb"},"entities":[{"name":"a","kind":"singular","format":"markdown","path":"a.md"}]}`)); err == nil {
		t.Fatal("expected unsupported backend")
	}
}
