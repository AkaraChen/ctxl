package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/AkaraChen/ctxl/core/schema/loader"
	"github.com/AkaraChen/ctxl/core/skillsgen"
)

func TestPublicConstructionBuildsSpecializedCLI(t *testing.T) {
	s, err := loader.Parse([]byte(`{
	  "name":"demo",
	  "entities":[{"name":"status","command":{"name":"current"},"kind":"singular","format":"markdown","path":"STATUS.md"}]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := skillsgen.DefaultBundle(s)
	if err != nil {
		t.Fatal(err)
	}
	cmd := New(Options{Schema: s, Skills: bundle})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	help := output.String()
	for _, want := range []string{"Usage:", "current", "init", "skills"} {
		if !strings.Contains(help, want) {
			t.Fatalf("help missing %q:\n%s", want, help)
		}
	}
	if strings.Contains(help, "status") || strings.Contains(help, "--schema") || strings.Contains(help, "schema validate") {
		t.Fatalf("help exposes a non-specialized command:\n%s", help)
	}
}
