package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/AkaraChen/ctxl/core/schema"
	"github.com/AkaraChen/ctxl/core/skillbundle"
	"github.com/spf13/cobra"
)

func TestSingletonSkillsOmitListAndAllowOmittedName(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	root := newTestCLI(singletonBundle())

	if hasSkillsList(t, root) {
		t.Fatal("expected skills list to be omitted from a singleton CLI")
	}

	body := execute(t, root, "skills", "get")
	if body != singletonMarkdown {
		t.Fatalf("skills get = %q", body)
	}

	named := execute(t, newTestCLI(singletonBundle()), "skills", "get", "demo-agent")
	if named != singletonMarkdown {
		t.Fatalf("named skills get = %q", named)
	}

	if _, err := executeErr(t, newTestCLI(singletonBundle()), "skills", "get", "missing"); err == nil || !strings.Contains(err.Error(), `unknown skill "missing"`) {
		t.Fatalf("wrong name error = %v", err)
	}

	if out, err := executeErr(t, newTestCLI(singletonBundle()), "skills", "list"); err == nil {
		t.Fatalf("expected singleton skills list to fail, got %q", out)
	}

	path := strings.TrimSpace(execute(t, newTestCLI(singletonBundle()), "skills", "path"))
	raw, err := os.ReadFile(path + "/SKILL.md")
	if err != nil || string(raw) != singletonMarkdown {
		t.Fatalf("materialized SKILL.md = %q err=%v path=%q", raw, err, path)
	}
}

func TestMultiSkillSkillsRequireNameAndKeepList(t *testing.T) {
	root := newTestCLI(multiBundle())
	if !hasSkillsList(t, root) {
		t.Fatal("expected skills list on a multi-skill CLI")
	}

	listed := execute(t, root, "skills", "list")
	if !strings.Contains(listed, "demo-agent") || !strings.Contains(listed, "custom-one") {
		t.Fatalf("skills list = %s", listed)
	}

	if _, err := executeErr(t, newTestCLI(multiBundle()), "skills", "get"); err == nil {
		t.Fatal("expected multi-skill get without a name to fail")
	}

	got := execute(t, newTestCLI(multiBundle()), "skills", "get", "custom-one")
	if got != customMarkdown {
		t.Fatalf("skills get custom-one = %q", got)
	}
}

func newTestCLI(bundle skillbundle.Bundle) *cobra.Command {
	return New(Options{
		Schema: schema.Schema{
			CLI:   schema.NameOverride{Name: "democtl"},
			Store: schema.NameOverride{Name: "demo"},
			Entities: []schema.Entity{{
				Name:    "status",
				Kind:    schema.KindSingular,
				Format:  schema.FormatMarkdown,
				Command: schema.NameOverride{Name: "status"},
			}},
		},
		Skills: bundle,
	})
}

func hasSkillsList(t *testing.T, root *cobra.Command) bool {
	t.Helper()
	skills, _, err := root.Find([]string{"skills"})
	if err != nil {
		t.Fatal(err)
	}
	for _, cmd := range skills.Commands() {
		if cmd.Name() == "list" {
			return true
		}
	}
	return false
}

func execute(t *testing.T, root *cobra.Command, args ...string) string {
	t.Helper()
	out, err := executeErr(t, root, args...)
	if err != nil {
		t.Fatalf("%s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return out
}

func executeErr(t *testing.T, root *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

const singletonMarkdown = "---\nname: demo-agent\ndescription: Demo.\n---\n\ninstructions\n"

const customMarkdown = "---\nname: custom-one\ndescription: Custom.\n---\n\nCUSTOM\n"

func singletonBundle() skillbundle.Bundle {
	return skillbundle.Bundle{Skills: []skillbundle.Skill{{
		Name:    "demo-agent",
		Entries: []skillbundle.Entry{{Path: "SKILL.md", Mode: 0o644, Data: []byte(singletonMarkdown)}},
	}}}
}

func multiBundle() skillbundle.Bundle {
	return skillbundle.Bundle{Skills: []skillbundle.Skill{
		{Name: "demo-agent", Entries: []skillbundle.Entry{{Path: "SKILL.md", Mode: 0o644, Data: []byte(singletonMarkdown)}}},
		{Name: "custom-one", Entries: []skillbundle.Entry{{Path: "SKILL.md", Mode: 0o644, Data: []byte(customMarkdown)}}},
	}}
}
