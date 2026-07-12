package parser

import (
	"reflect"
	"testing"
)

func newPnp() *PackageJSON { return NewPackageJSON("npm") }

func TestParseScripts(t *testing.T) {
	input := `{
		"scripts": {
			"dev": "next dev",
			"build": "tsc",
			"test": "jest"
		}
	}`

	tasks, ctx, err := newPnp().parse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}

	// Tasks should be sorted by name.
	expected := []struct{ name, cmd string }{
		{"build", "tsc"},
		{"dev", "next dev"},
		{"test", "jest"},
	}
	for i, want := range expected {
		if tasks[i].Name != want.name {
			t.Errorf("task[%d].Name = %q, want %q", i, tasks[i].Name, want.name)
		}
		if tasks[i].Command != want.cmd {
			t.Errorf("task[%d].Command = %q, want %q", i, tasks[i].Command, want.cmd)
		}
	}

	// RunContext should reflect npm defaults.
	wantCtx := RunContext{
		Binary:        "npm",
		Subcmd:        "run",
		ArgSeparator:  "--",
		DisplayPrefix: []string{"npm", "run"},
	}
	if !reflect.DeepEqual(ctx, wantCtx) {
		t.Errorf("ctx = %+v, want %+v", ctx, wantCtx)
	}
}

func TestParseWithRuntuiConfig(t *testing.T) {
	input := `{
		"scripts": {
			"dev": "next dev",
			"env:set": "dotenvx set"
		},
		"runtui": {
			"dev": {
				"description": "Start dev server"
			},
			"env:set": {
				"description": "Set env var",
				"args": [
					{"name": "KEY", "required": true, "hint": "e.g. DATABASE_URL"},
					{"name": "VALUE", "required": true}
				]
			}
		}
	}`

	tasks, _, err := newPnp().parse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}

	// dev: has description, no args.
	dev := tasks[0]
	if dev.Name != "dev" {
		t.Fatalf("expected dev, got %s", dev.Name)
	}
	if dev.Description != "Start dev server" {
		t.Errorf("dev.Description = %q, want %q", dev.Description, "Start dev server")
	}
	if len(dev.Args) != 0 {
		t.Errorf("dev.Args should be empty, got %d", len(dev.Args))
	}

	// env:set: has description and 2 args.
	envSet := tasks[1]
	if envSet.Name != "env:set" {
		t.Fatalf("expected env:set, got %s", envSet.Name)
	}
	if len(envSet.Args) != 2 {
		t.Fatalf("env:set should have 2 args, got %d", len(envSet.Args))
	}
	if envSet.Args[0].Name != "KEY" || !envSet.Args[0].Required {
		t.Errorf("arg[0] = %+v, want KEY required", envSet.Args[0])
	}
	if envSet.Args[0].Hint != "e.g. DATABASE_URL" {
		t.Errorf("arg[0].Hint = %q, want %q", envSet.Args[0].Hint, "e.g. DATABASE_URL")
	}
}

func TestParseEmptyScripts(t *testing.T) {
	tasks, _, err := newPnp().parse([]byte(`{"scripts": {}}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tasks != nil {
		t.Errorf("expected nil tasks for empty scripts, got %d", len(tasks))
	}
}

func TestParseNoScripts(t *testing.T) {
	tasks, _, err := newPnp().parse([]byte(`{"name": "my-project"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tasks != nil {
		t.Errorf("expected nil tasks, got %d", len(tasks))
	}
}

func TestParseMalformedJSON(t *testing.T) {
	_, _, err := newPnp().parse([]byte(`{invalid`))
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestParseRuntuiConfigForUnknownScript(t *testing.T) {
	input := `{
		"scripts": {"dev": "next dev"},
		"runtui": {
			"nonexistent": {"description": "ghost"}
		}
	}`

	tasks, _, err := newPnp().parse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Description != "" {
		t.Errorf("dev should have no description, got %q", tasks[0].Description)
	}
}

func TestParseDifferentRunners(t *testing.T) {
	for _, runner := range []string{"npm", "yarn", "pnpm", "bun"} {
		p := NewPackageJSON(runner)
		_, ctx, err := p.parse([]byte(`{"scripts": {"dev": "next dev"}}`))
		if err != nil {
			t.Fatalf("runner %s: %v", runner, err)
		}
		if ctx.Binary != runner {
			t.Errorf("runner %s: ctx.Binary = %q", runner, ctx.Binary)
		}
	}
}
