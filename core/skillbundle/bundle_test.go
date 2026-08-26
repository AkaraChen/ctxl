package skillbundle

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestMaterializeCompleteSkillConcurrently(t *testing.T) {
	cache := t.TempDir()
	bundle := Bundle{Skills: []Skill{{
		Name: "demo-skill",
		Entries: []Entry{
			{Path: "SKILL.md", Mode: 0o644, Data: []byte("---\nname: demo-skill\ndescription: Demo.\n---\n")},
			{Path: "scripts", Mode: 0o755, Directory: true},
			{Path: "scripts/run.sh", Mode: 0o755, Data: []byte("#!/bin/sh\n")},
			{Path: "assets/blob.bin", Mode: 0o640, Data: []byte{0, 1, 255}},
			{Path: "empty", Mode: 0o700, Directory: true},
		},
	}}}

	const callers = 12
	paths := make(chan string, callers)
	errs := make(chan error, callers)
	var wg sync.WaitGroup
	for range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			path, err := bundle.materializeAt("demo-skill", cache)
			paths <- path
			errs <- err
		}()
	}
	wg.Wait()
	close(paths)
	close(errs)
	var expected string
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	for path := range paths {
		if expected == "" {
			expected = path
		}
		if path != expected {
			t.Fatalf("materialized paths differ: %q and %q", expected, path)
		}
	}
	assertFile(t, filepath.Join(expected, "assets", "blob.bin"), []byte{0, 1, 255}, 0o640)
	assertFile(t, filepath.Join(expected, "scripts", "run.sh"), []byte("#!/bin/sh\n"), 0o755)
	if info, err := os.Stat(filepath.Join(expected, "empty")); err != nil || !info.IsDir() || info.Mode().Perm() != 0o700 {
		t.Fatalf("empty directory was not preserved: info=%v err=%v", info, err)
	}
}

func TestMaterializeRejectsUnsafeEntry(t *testing.T) {
	bundle := Bundle{Skills: []Skill{{Name: "bad", Entries: []Entry{{Path: "../escape", Data: []byte("x")}}}}}
	if _, err := bundle.materializeAt("bad", t.TempDir()); err == nil {
		t.Fatal("expected unsafe path error")
	}
}

func TestMaterializeRejectsUnsafeSkillName(t *testing.T) {
	bundle := Bundle{Skills: []Skill{{Name: "../bad", Entries: []Entry{{Path: "SKILL.md", Data: []byte("x")}}}}}
	if _, err := bundle.materializeAt("../bad", t.TempDir()); err == nil {
		t.Fatal("expected unsafe skill name error")
	}
}

func assertFile(t *testing.T, path string, data []byte, mode os.FileMode) {
	t.Helper()
	actual, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, data) {
		t.Fatalf("%s data = %v", path, actual)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != mode {
		t.Fatalf("%s mode = %o, want %o", path, info.Mode().Perm(), mode)
	}
}
