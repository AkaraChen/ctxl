package skillsgen

import (
	"strings"
	"testing"

	"github.com/AkaraChen/ctxl/core/schema"
)

func TestInstructionsUseZeroArgSkillsCommandsForSingleton(t *testing.T) {
	s := schema.Schema{
		CLI:    schema.NameOverride{Name: "democtl"},
		Store:  schema.NameOverride{Name: "demo"},
		Skills: []schema.Skill{{Type: schema.SkillBuiltin, Name: "demo-agent"}},
		Entities: []schema.Entity{{
			Name:    "status",
			Kind:    schema.KindSingular,
			Format:  schema.FormatMarkdown,
			Command: schema.NameOverride{Name: "status"},
		}},
	}
	text := instructions(s)
	for _, want := range []string{"`democtl skills get`", "`democtl skills path`"} {
		if !strings.Contains(text, want) {
			t.Fatalf("singleton instructions missing %q:\n%s", want, text)
		}
	}
	for _, refuse := range []string{"skills list", "skills get NAME", "skills path NAME"} {
		if strings.Contains(text, refuse) {
			t.Fatalf("singleton instructions still document %q:\n%s", refuse, text)
		}
	}
}

func TestInstructionsKeepNamedSkillsCommandsForMultipleSkills(t *testing.T) {
	s := schema.Schema{
		CLI:   schema.NameOverride{Name: "democtl"},
		Store: schema.NameOverride{Name: "demo"},
		Skills: []schema.Skill{
			{Type: schema.SkillBuiltin, Name: "demo-agent"},
			{Type: schema.SkillCustom, Directory: "custom-one"},
		},
		Entities: []schema.Entity{{
			Name:    "status",
			Kind:    schema.KindSingular,
			Format:  schema.FormatMarkdown,
			Command: schema.NameOverride{Name: "status"},
		}},
	}
	text := instructions(s)
	for _, want := range []string{"`democtl skills list`", "`democtl skills get NAME`", "`democtl skills path NAME`"} {
		if !strings.Contains(text, want) {
			t.Fatalf("multi-skill instructions missing %q:\n%s", want, text)
		}
	}
}
