package schema

import "testing"

func TestParseYoiExample(t *testing.T) {
	raw := []byte(`{
	  "name": "yoi",
	  "entities": [
	    {"name":"deploy","kind":"singular","format":"markdown","path":"DEPLOY.md","location":"root","fields":[{"name":"service","type":"string","required":true}]},
	    {"name":"log","kind":"plural","format":"ndjson","path":"deploy.log","fields":[{"name":"result","type":"string","required":true}]}
	  ]
	}`)
	s, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "yoi" || len(s.Entities) != 2 {
		t.Fatalf("%+v", s)
	}
}

func TestRejectBadKind(t *testing.T) {
	_, err := Parse([]byte(`{"name":"x","entities":[{"name":"a","kind":"maybe","format":"markdown","path":"a.md"}]}`))
	if err == nil {
		t.Fatal("expected error")
	}
}
