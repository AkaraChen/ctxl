package codegen

import (
	"bytes"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateStandaloneBuildAndRuntime(t *testing.T) {
	root := t.TempDir()
	repo := repositoryRoot(t)
	writeTestFile(t, root, "custom-one/SKILL.md", []byte("---\nname: custom-one\ndescription: Custom instructions.\n---\n\nCUSTOM BODY\n"), 0o644)
	writeTestFile(t, root, "custom-one/scripts/run.sh", []byte("#!/bin/sh\necho ok\n"), 0o755)
	schemaPath := writeTestFile(t, root, "context.schema.json", []byte(`{
	  "name":"demo",
	  "description":"Demo generated CLI.",
	  "generation":{"output":"out","module":"example.com/demo"},
	  "cli":{"name":"democtl"},
	  "store":{"name":"demo-data"},
	  "skills":[{"type":"builtin","name":"demo-agent"},{"type":"custom","directory":"custom-one"}],
	  "entities":[
	    {"name":"status","command":{"name":"current"},"kind":"singular","format":"markdown","path":"STATUS.md","location":"root","scope":"project","fields":[{"name":"service","type":"string","required":true}]},
	    {"name":"events","kind":"plural","format":"ndjson","path":"events.ndjson","scope":"project","fields":[{"name":"result","type":"string","required":true},{"name":"details","type":"object"}]}
	  ]
	}`), 0o644)

	result, err := generate(schemaPath, runtimeDependency{version: "v0.0.0", replace: repo})
	if err != nil {
		t.Fatal(err)
	}
	if result != filepath.Join(root, "out") {
		t.Fatalf("output = %q", result)
	}
	mainSource, err := os.ReadFile(filepath.Join(result, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "main.go", mainSource, parser.AllErrors); err != nil {
		t.Fatalf("generated source is not valid Go: %v", err)
	}
	if !bytes.Contains(mainSource, []byte("schema.Schema")) || !bytes.Contains(mainSource, []byte("skillbundle.Bundle")) || bytes.Contains(mainSource, []byte("schema.Parse")) || bytes.Contains(mainSource, []byte("--schema")) {
		t.Fatalf("unexpected generated source:\n%s", mainSource)
	}
	generatedModule, err := os.ReadFile(filepath.Join(result, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	for _, generatorOnly := range []string{"go-playground/validator", "santhosh-tekuri/jsonschema", "gopkg.in/yaml"} {
		if bytes.Contains(generatedModule, []byte(generatorOnly)) {
			t.Fatalf("generated runtime depends on generator-only module %q:\n%s", generatorOnly, generatedModule)
		}
	}

	binary := filepath.Join(root, "democtl")
	runCommand(t, result, "go", "build", "-o", binary, ".")
	help := runCommand(t, root, binary, "--help")
	for _, want := range []string{"Demo generated CLI.", "current", "skills"} {
		if !strings.Contains(help, want) {
			t.Fatalf("help missing %q:\n%s", want, help)
		}
	}
	if strings.Contains(help, "--schema") || strings.Contains(help, "schema validate") {
		t.Fatalf("runtime help exposes removed schema commands:\n%s", help)
	}
	runCommand(t, root, binary, "current", "write", "--service", "api")
	show := runCommand(t, root, binary, "current", "show")
	if !strings.Contains(show, `"service": "api"`) {
		t.Fatalf("show output = %s", show)
	}
	if _, err := os.Stat(filepath.Join(root, "STATUS.md")); err != nil {
		t.Fatal(err)
	}
	runCommand(t, root, binary, "events", "append", "--result", "green", "--details", `{"attempt":1}`)
	listed := runCommand(t, root, binary, "events", "list")
	if strings.Contains(listed, "details") || !strings.Contains(listed, `"result": "green"`) {
		t.Fatalf("default event list = %s", listed)
	}
	full := runCommand(t, root, binary, "events", "list", "--full")
	if !strings.Contains(full, "details") {
		t.Fatalf("full event list = %s", full)
	}
	skills := runCommand(t, root, binary, "skills", "list")
	if !strings.Contains(skills, "demo-agent") || !strings.Contains(skills, "custom-one") {
		t.Fatalf("skill list = %s", skills)
	}
	custom := runCommand(t, root, binary, "skills", "get", "custom-one")
	if custom != "---\nname: custom-one\ndescription: Custom instructions.\n---\n\nCUSTOM BODY\n" {
		t.Fatalf("custom SKILL.md was rewritten:\n%s", custom)
	}
	skillPath := strings.TrimSpace(runCommandWithEnv(t, root, []string{"HOME=" + filepath.Join(root, "home")}, binary, "skills", "path", "custom-one"))
	asset, err := os.ReadFile(filepath.Join(skillPath, "scripts", "run.sh"))
	if err != nil || string(asset) != "#!/bin/sh\necho ok\n" {
		t.Fatalf("materialized script = %q, err=%v", asset, err)
	}

	stale := filepath.Join(result, "stale.txt")
	if err := os.WriteFile(stale, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := generate(schemaPath, runtimeDependency{version: "v0.0.0", replace: repo}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale generated file survived: %v", err)
	}
}

func TestGenerateRequiredOnlySchemaUsesDerivedStandaloneDefaults(t *testing.T) {
	root := t.TempDir()
	repo := repositoryRoot(t)
	schemaPath := writeTestFile(t, root, "schema.json", []byte(`{
	  "name":"demo",
	  "entities":[{"name":"status","kind":"singular","format":"markdown","path":"STATUS.md"}]
	}`), 0o644)
	result, err := generate(schemaPath, runtimeDependency{version: "v0.0.0", replace: repo})
	if err != nil {
		t.Fatal(err)
	}
	if result != filepath.Join(root, "generated", "demo") {
		t.Fatalf("output = %q", result)
	}
	goMod, err := os.ReadFile(filepath.Join(result, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(goMod, []byte("module demo")) {
		t.Fatalf("module was not derived from name:\n%s", goMod)
	}
	runCommand(t, result, "go", "build", ".")
}

func TestGenerateFailurePreservesPreviousOutput(t *testing.T) {
	root := t.TempDir()
	repo := repositoryRoot(t)
	skillPath := writeTestFile(t, root, "custom-one/SKILL.md", []byte("---\nname: custom-one\ndescription: Custom.\n---\n"), 0o644)
	schemaPath := writeTestFile(t, root, "schema.json", []byte(`{
	  "name":"demo",
	  "generation":{"output":"out"},
	  "skills":[{"type":"custom","directory":"custom-one"}],
	  "entities":[{"name":"status","kind":"singular","format":"markdown","path":"STATUS.md"}]
	}`), 0o644)
	if _, err := generate(schemaPath, runtimeDependency{version: "v0.0.0", replace: repo}); err != nil {
		t.Fatal(err)
	}
	mainPath := filepath.Join(root, "out", "main.go")
	before, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(skillPath); err != nil {
		t.Fatal(err)
	}
	if _, err := generate(schemaPath, runtimeDependency{version: "v0.0.0", replace: repo}); err == nil {
		t.Fatal("expected invalid Skill to fail generation")
	}
	after, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("failed generation changed previous output")
	}
}

func TestGenerateRejectsUnmarkedOutput(t *testing.T) {
	root := t.TempDir()
	repo := repositoryRoot(t)
	writeTestFile(t, root, "out/user.txt", []byte("mine"), 0o644)
	schemaPath := writeTestFile(t, root, "schema.json", []byte(`{
	  "name":"demo",
	  "generation":{"output":"out"},
	  "entities":[{"name":"status","kind":"singular","format":"markdown","path":"STATUS.md"}]
	}`), 0o644)
	_, err := generate(schemaPath, runtimeDependency{version: "v0.0.0", replace: repo})
	if err == nil || !strings.Contains(err.Error(), "unmarked non-empty") {
		t.Fatalf("unexpected error: %v", err)
	}
	if raw, readErr := os.ReadFile(filepath.Join(root, "out", "user.txt")); readErr != nil || string(raw) != "mine" {
		t.Fatalf("unmarked output changed: %q, %v", raw, readErr)
	}
}

func TestGenerateExistingModuleDoesNotEditModuleFiles(t *testing.T) {
	root := t.TempDir()
	repo := repositoryRoot(t)
	goMod := []byte("module example.com/host\n\ngo 1.24.0\n\nrequire github.com/AkaraChen/ctxl v0.0.0\n\nreplace github.com/AkaraChen/ctxl => " + filepath.ToSlash(repo) + "\n")
	goModPath := writeTestFile(t, root, "go.mod", goMod, 0o644)
	precondition := writeTestFile(t, root, "precondition.go", []byte("package host\n\nimport _ \"github.com/AkaraChen/ctxl/cli\"\n"), 0o644)
	runCommand(t, root, "go", "mod", "tidy")
	if err := os.Remove(precondition); err != nil {
		t.Fatal(err)
	}
	goMod, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatal(err)
	}
	goSumPath := filepath.Join(root, "go.sum")
	goSum, err := os.ReadFile(goSumPath)
	if err != nil {
		t.Fatal(err)
	}
	schemaPath := writeTestFile(t, root, "schema.json", []byte(`{
	  "name":"hostctl",
	  "generation":{"mode":"existing-module","output":"cmd/hostctl","ctxl_version":"v0.0.0"},
	  "entities":[{"name":"status","kind":"singular","format":"markdown","path":"STATUS.md"}]
	}`), 0o644)
	result, err := Generate(schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(goMod, after) {
		t.Fatalf("generator edited go.mod:\n%s", after)
	}
	afterSum, err := os.ReadFile(goSumPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(goSum, afterSum) {
		t.Fatal("generator edited go.sum")
	}
	runCommand(t, root, "go", "build", "-o", filepath.Join(root, "hostctl"), "./cmd/hostctl")
	if result != filepath.Join(root, "cmd", "hostctl") {
		t.Fatalf("output = %q", result)
	}
}

func TestExistingModuleMissingDependencyHasRemediation(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "go.mod", []byte("module example.com/host\n\ngo 1.24.0\n"), 0o644)
	schemaPath := writeTestFile(t, root, "schema.json", []byte(`{
	  "name":"hostctl",
	  "generation":{"mode":"existing-module","ctxl_version":"v1.2.3"},
	  "entities":[{"name":"status","kind":"singular","format":"markdown","path":"STATUS.md"}]
	}`), 0o644)
	_, err := Generate(schemaPath)
	if err == nil || !strings.Contains(err.Error(), "go get github.com/AkaraChen/ctxl@v1.2.3") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func writeTestFile(t *testing.T, root, rel string, data []byte, mode os.FileMode) string {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, mode); err != nil {
		t.Fatal(err)
	}
	return path
}

func runCommand(t *testing.T, dir, name string, args ...string) string {
	t.Helper()
	return runCommandWithEnv(t, dir, nil, name, args...)
}

func runCommandWithEnv(t *testing.T, dir string, extraEnv []string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), append([]string{"GOWORK=off"}, extraEnv...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed: %v\n%s", name, strings.Join(args, " "), err, out)
	}
	return string(out)
}
