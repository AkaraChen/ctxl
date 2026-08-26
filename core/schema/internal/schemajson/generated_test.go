package schemajson

import (
	"encoding/json"
	"testing"
)

func TestGeneratedDecoderAppliesLiteralDefaults(t *testing.T) {
	var document Document
	err := json.Unmarshal([]byte(`{
	  "name":"demo",
	  "entities":[{"name":"status","kind":"singular","format":"markdown","path":"STATUS.md"}]
	}`), &document)
	if err != nil {
		t.Fatal(err)
	}
	entity := document.Entities[0]
	if document.Generation.Mode != GenerationModeStandalone || entity.Location != EntityLocationStore || entity.Scope != EntityScopeBoth || entity.Id != "id" || entity.Write != EntityWriteReplace {
		t.Fatalf("generated defaults were not applied: %+v %+v", document.Generation, entity)
	}

	var builtin BuiltinSkill
	if err := json.Unmarshal([]byte(`{"type":"builtin"}`), &builtin); err != nil {
		t.Fatal(err)
	}
	if builtin.Inject != BuiltinSkillInjectAfter {
		t.Fatalf("builtin inject default = %q", builtin.Inject)
	}
}
