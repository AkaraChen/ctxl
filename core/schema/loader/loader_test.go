package loader

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AkaraChen/ctxl/core/schema"
)

func TestParseAppliesSchemaDefaults(t *testing.T) {
	raw := []byte(`{
	  "name": "demo",
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
	if s.CLI.Name != "demo" || s.Store.Name != "demo" || s.Generation.Module != "demo" || s.Generation.Output != "generated/demo" {
		t.Fatalf("defaults not derived from name: %+v", s)
	}
	if len(s.Skills) != 1 || s.Skills[0].Type != schema.SkillBuiltin || s.Skills[0].Name != "demo" {
		t.Fatalf("implicit builtin missing: %+v", s.Skills)
	}
	log := s.Entities[1]
	if log.Command.Name != "log" || log.Location != schema.LocationStore || log.Scope != schema.ScopeBoth || log.ID != "id" || log.Write != schema.WriteReplace {
		t.Fatalf("entity defaults not normalized: %+v", log)
	}
}

func TestLoadFileReportsStructuredSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.schema.json")
	if err := os.WriteFile(path, []byte(`{"name":"x","unknown":true,"entities":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadFile(path)
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("error is not structured: %v", err)
	}
	if validation.Source != path || validation.Path != "" {
		t.Fatalf("validation error = %+v", validation)
	}
}

func TestParseOverrides(t *testing.T) {
	s, err := Parse([]byte(`{
	  "name":"demo",
	  "generation":{"mode":"standalone","output":"tools/command","module":"example.com/demo","ctxl_version":"v1.2.3"},
	  "cli":{"name":"democtl"},
	  "store":{"name":"demo-data"},
	  "skills":[{"type":"builtin","name":"demo-agent","description":"Use demo."}],
	  "entities":[{"name":"status","command":{"name":"current"},"kind":"singular","format":"markdown","path":"STATUS.md"}]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if s.CLI.Name != "democtl" || s.Store.Name != "demo-data" || s.Generation.Output != "tools/command" {
		t.Fatalf("identity overrides lost: %+v", s)
	}
	if s.Entities[0].Command.Name != "current" || s.Skills[0].Name != "demo-agent" {
		t.Fatalf("command or skill override lost: %+v", s)
	}
}

func TestParsePrependsImplicitBuiltinToCustomSkills(t *testing.T) {
	s, err := Parse([]byte(`{
	  "name":"demo",
	  "skills":[{"type":"custom","directory":"custom-one"}],
	  "entities":[{"name":"status","kind":"singular","format":"markdown","path":"STATUS.md"}]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Skills) != 2 || s.Skills[0].Type != schema.SkillBuiltin || s.Skills[0].Inject != schema.InjectAfter || s.Skills[1].Type != schema.SkillCustom {
		t.Fatalf("effective Skills = %+v", s.Skills)
	}
}

func TestSchemaContractRejections(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "invalid entity kind",
			raw:  `{"name":"x","entities":[{"name":"a","kind":"maybe","format":"markdown","path":"a.md"}]}`,
			want: "kind",
		},
		{
			name: "section missing heading",
			raw:  `{"name":"x","entities":[{"name":"a","kind":"singular","format":"markdown","write":"section","path":"a.md"}]}`,
			want: "section",
		},
		{
			name: "unknown root key",
			raw:  `{"name":"x","surprise":true,"entities":[{"name":"a","kind":"singular","format":"markdown","path":"a.md"}]}`,
			want: "additional properties",
		},
		{
			name: "custom skill extra key",
			raw:  `{"name":"x","skills":[{"type":"custom","directory":"skill","name":"nope"}],"entities":[{"name":"a","kind":"singular","format":"markdown","path":"a.md"}]}`,
			want: "oneOf",
		},
		{
			name: "multiple builtins",
			raw:  `{"name":"x","skills":[{"type":"builtin"},{"type":"builtin"}],"entities":[{"name":"a","kind":"singular","format":"markdown","path":"a.md"}]}`,
			want: "contains",
		},
		{
			name: "unsafe output",
			raw:  `{"name":"x","generation":{"output":"../outside"},"entities":[{"name":"a","kind":"singular","format":"markdown","path":"a.md"}]}`,
			want: "allOf",
		},
		{
			name: "non-string metadata",
			raw:  `{"name":"x","skills":[{"type":"builtin","metadata":{"count":1}}],"entities":[{"name":"a","kind":"singular","format":"markdown","path":"a.md"}]}`,
			want: "type",
		},
		{
			name: "existing module with unused module override",
			raw:  `{"name":"x","generation":{"mode":"existing-module","module":"example.com/x"},"entities":[{"name":"a","kind":"singular","format":"markdown","path":"a.md"}]}`,
			want: "not",
		},
		{
			name: "symlink with unused fields",
			raw:  `{"name":"x","entities":[{"name":"a","kind":"singular","format":"symlink","path":"a","target":"b","fields":[]}]}`,
			want: "not",
		},
		{
			name: "section with unused fields",
			raw:  `{"name":"x","entities":[{"name":"a","kind":"singular","format":"markdown","path":"a.md","write":"section","section":"A","fields":[]}]}`,
			want: "not",
		},
		{
			name: "global entity at project root",
			raw:  `{"name":"x","entities":[{"name":"a","kind":"singular","format":"markdown","path":"a.md","scope":"global","location":"root"}]}`,
			want: "must be 'store'",
		},
		{
			name: "invalid derived builtin name",
			raw:  `{"name":"Not A Skill","cli":{"name":"not-a-skill"},"store":{"name":"not-a-skill"},"generation":{"mode":"existing-module"},"entities":[{"name":"a","kind":"singular","format":"markdown","path":"a.md"}]}`,
			want: "does not match pattern",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.raw))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error %v does not contain %q", err, tt.want)
			}
		})
	}
}
