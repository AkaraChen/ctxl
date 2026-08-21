package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AkaraChen/ctxl/core/schema"
)

func runCLI(t *testing.T, dir string, args ...string) string {
	t.Helper()
	s, err := schema.LoadFile("../examples/demo.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	cmd := New(Options{Name: "ctxl", Schema: s})
	cmd.SetArgs(args)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	errRun := cmd.Execute()
	w.Close()
	os.Stdout = stdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	if errRun != nil {
		t.Fatalf("%v\n%s", errRun, buf.String())
	}
	return buf.String()
}

func TestShowAndLogList(t *testing.T) {
	dir := t.TempDir()
	schemaPath, err := filepath.Abs("../examples/demo.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	runCLI(t, dir, "--schema", schemaPath, "status", "write", "--service", "hermes", "--start", "up", "--stop", "down")
	out := runCLI(t, dir, "--schema", schemaPath, "status", "show")
	var rec map[string]any
	if err := json.Unmarshal([]byte(out), &rec); err != nil {
		t.Fatal(err)
	}
	if _, ok := rec["Fields"]; ok {
		t.Fatalf("internal wrapper: %s", out)
	}
	if rec["service"] != "hermes" {
		t.Fatalf("service %v", rec["service"])
	}
	if rec["last_green"] == "" || rec["last_green"] == nil {
		t.Fatalf("last_green empty: %s", out)
	}
	runCLI(t, dir, "--schema", schemaPath, "log", "append", "--result", "green", "--cmd", "up", "--custom_data", `{"k":"v"}`)
	listed := runCLI(t, dir, "--schema", schemaPath, "log", "list")
	if strings.Contains(listed, "custom_data") {
		t.Fatalf("list leaked custom_data: %s", listed)
	}
	full := runCLI(t, dir, "--schema", schemaPath, "log", "list", "--full")
	if !strings.Contains(full, "custom_data") {
		t.Fatalf("full missing custom_data: %s", full)
	}
}
