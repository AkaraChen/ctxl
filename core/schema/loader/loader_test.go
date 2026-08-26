package loader

import (
	"testing"

	"github.com/AkaraChen/ctxl/core/schema"
)

// Guards the hand-written default/derivation step that replaced the generated
// decoder: literal JSON Schema defaults, name-derived identities, and the
// implicit built-in Skill.
func TestParseAppliesSchemaDefaults(t *testing.T) {
	s, err := Parse([]byte(`{
	  "name": "demo",
	  "skills": [{"type":"custom","directory":"custom-one"}],
	  "entities": [
	    {"name":"log","kind":"plural","format":"ndjson","path":"events.log","fields":[{"name":"result","type":"string","required":true}]}
	  ]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if s.CLI.Name != "demo" || s.Store.Name != "demo" || s.Generation.Mode != schema.GenerationStandalone || s.Generation.Module != "demo" || s.Generation.Output != "generated/demo" {
		t.Fatalf("defaults not derived from name: %+v", s)
	}
	if len(s.Skills) != 2 || s.Skills[0].Type != schema.SkillBuiltin || s.Skills[0].Name != "demo" || s.Skills[0].Inject != schema.InjectAfter || s.Skills[1].Type != schema.SkillCustom {
		t.Fatalf("implicit builtin missing: %+v", s.Skills)
	}
	log := s.Entities[0]
	if log.Command.Name != "log" || log.Location != schema.LocationStore || log.Scope != schema.ScopeBoth || log.ID != "id" || log.Write != schema.WriteReplace {
		t.Fatalf("entity defaults not normalized: %+v", log)
	}
}

func TestParseRejectsContractViolations(t *testing.T) {
	for name, raw := range map[string]string{
		"invalid entity kind": `{"name":"x","entities":[{"name":"a","kind":"maybe","format":"markdown","path":"a.md"}]}`,
		"unknown root key":    `{"name":"x","surprise":true,"entities":[{"name":"a","kind":"singular","format":"markdown","path":"a.md"}]}`,
		"multiple builtins":   `{"name":"x","skills":[{"type":"builtin"},{"type":"builtin"}],"entities":[{"name":"a","kind":"singular","format":"markdown","path":"a.md"}]}`,
	} {
		if _, err := Parse([]byte(raw)); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}
