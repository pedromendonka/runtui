package parser

import (
	"testing"
)

func TestMakefileParse(t *testing.T) {
	p := &MakefileParser{}
	tasks, ctx, err := p.Parse("../testdata/Makefile")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.Binary != "make" {
		t.Errorf("Binary = %q, want %q", ctx.Binary, "make")
	}
	if ctx.Subcmd != "" {
		t.Errorf("Subcmd = %q, want empty", ctx.Subcmd)
	}
	if ctx.ArgSeparator != "" {
		t.Errorf("ArgSeparator = %q, want empty", ctx.ArgSeparator)
	}

	// Fixture has 6 documented targets: all, build, clean, deploy, lint, test.
	// Undocumented targets (vet, fmt-check) are excluded because documented ones exist.
	want := []struct {
		name string
		desc string
	}{
		{"all", "Build everything (default)"},
		{"build", "Compile the binary"},
		{"clean", "Remove build artifacts"},
		{"deploy", "Deploy the application"},
		{"lint", "Run all linters"},
		{"test", "Run all tests"},
	}

	if len(tasks) != len(want) {
		t.Fatalf("got %d tasks, want %d: %v", len(tasks), len(want), taskNames(tasks))
	}

	for i, w := range want {
		if tasks[i].Name != w.name {
			t.Errorf("tasks[%d].Name = %q, want %q", i, tasks[i].Name, w.name)
		}
		if tasks[i].Description != w.desc {
			t.Errorf("tasks[%d].Description = %q, want %q", i, tasks[i].Description, w.desc)
		}
		wantCmd := "make " + w.name
		if tasks[i].Command != wantCmd {
			t.Errorf("tasks[%d].Command = %q, want %q", i, tasks[i].Command, wantCmd)
		}
	}
}

func TestMakefileParseUndocumented(t *testing.T) {
	// When no targets have ## descriptions, show all targets.
	data := []byte("build:\n\tgo build .\n\ntest:\n\tgo test ./...\n")
	p := &MakefileParser{}
	tasks, _, err := p.parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2: %v", len(tasks), taskNames(tasks))
	}
	if tasks[0].Name != "build" {
		t.Errorf("tasks[0].Name = %q, want %q", tasks[0].Name, "build")
	}
	if tasks[1].Name != "test" {
		t.Errorf("tasks[1].Name = %q, want %q", tasks[1].Name, "test")
	}
}

func TestMakefileParseSkipsVariables(t *testing.T) {
	data := []byte("FOO := bar\nBAR ?= baz\nQUX += quux\nbuild:\n\techo ok\n")
	p := &MakefileParser{}
	tasks, _, err := p.parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1: %v", len(tasks), taskNames(tasks))
	}
	if tasks[0].Name != "build" {
		t.Errorf("tasks[0].Name = %q, want %q", tasks[0].Name, "build")
	}
}

func TestMakefileParseSkipsDotTargets(t *testing.T) {
	data := []byte(".PHONY: build\n.DEFAULT_GOAL := build\nbuild: ## Build\n\tgo build .\n")
	p := &MakefileParser{}
	tasks, _, err := p.parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1: %v", len(tasks), taskNames(tasks))
	}
	if tasks[0].Name != "build" {
		t.Errorf("tasks[0].Name = %q, want %q", tasks[0].Name, "build")
	}
}

func TestMakefileParseEmpty(t *testing.T) {
	p := &MakefileParser{}
	tasks, ctx, err := p.parse([]byte(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("got %d tasks, want 0", len(tasks))
	}
	if ctx.Binary != "make" {
		t.Errorf("Binary = %q, want %q", ctx.Binary, "make")
	}
}

func TestMakefileParseTargetNames(t *testing.T) {
	// Targets can contain hyphens, dots, and underscores.
	data := []byte("foo-bar: ## Hyphen\nfoo.bar: ## Dot\nfoo_bar: ## Underscore\n")
	p := &MakefileParser{}
	tasks, _, err := p.parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tasks) != 3 {
		t.Fatalf("got %d tasks, want 3: %v", len(tasks), taskNames(tasks))
	}
	// Sorted: foo-bar, foo.bar, foo_bar
	if tasks[0].Name != "foo-bar" {
		t.Errorf("tasks[0].Name = %q, want %q", tasks[0].Name, "foo-bar")
	}
	if tasks[1].Name != "foo.bar" {
		t.Errorf("tasks[1].Name = %q, want %q", tasks[1].Name, "foo.bar")
	}
	if tasks[2].Name != "foo_bar" {
		t.Errorf("tasks[2].Name = %q, want %q", tasks[2].Name, "foo_bar")
	}
}

func TestMakefileParseFileNotFound(t *testing.T) {
	p := &MakefileParser{}
	_, _, err := p.Parse("/nonexistent/Makefile")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestMakefileParseFromFixture(t *testing.T) {
	// Verify we can read the testdata fixture via the public Parse method.
	p := &MakefileParser{}
	tasks, _, err := p.Parse("../testdata/Makefile")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) == 0 {
		t.Fatal("expected tasks from fixture")
	}
}

func TestMakefileRunContext(t *testing.T) {
	p := &MakefileParser{}
	_, ctx, _ := p.parse([]byte("build:\n\tgo build .\n"))
	if ctx.Binary != "make" {
		t.Errorf("Binary = %q, want %q", ctx.Binary, "make")
	}
	if ctx.Subcmd != "" {
		t.Errorf("Subcmd = %q, want empty", ctx.Subcmd)
	}
	if ctx.ArgSeparator != "" {
		t.Errorf("ArgSeparator = %q, want empty", ctx.ArgSeparator)
	}
	if len(ctx.DisplayPrefix) != 1 || ctx.DisplayPrefix[0] != "make" {
		t.Errorf("DisplayPrefix = %v, want [make]", ctx.DisplayPrefix)
	}
}

// taskNames is a test helper that returns task names for error messages.
func taskNames(tasks []Task) []string {
	names := make([]string, len(tasks))
	for i, t := range tasks {
		names[i] = t.Name
	}
	return names
}
