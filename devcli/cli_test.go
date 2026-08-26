package devcli

import "testing"

func TestDeveloperCLIOnlyExposesGeneration(t *testing.T) {
	root := New()
	commands := map[string]bool{}
	for _, cmd := range root.Commands() {
		commands[cmd.Name()] = true
	}
	if !commands["generate"] || commands["completion"] || commands["status"] || commands["skills"] || commands["schema"] || commands["init"] {
		t.Fatalf("developer commands = %#v", commands)
	}
	if root.PersistentFlags().Lookup("schema") != nil {
		t.Fatal("developer CLI exposes --schema")
	}
}
