package runner

import (
	"testing"
)

func TestBuildCmdNoArgs(t *testing.T) {
	cmd := BuildCmd("npm", "run", "dev", nil)

	if cmd.Path == "" {
		t.Fatal("expected non-empty path")
	}

	args := cmd.Args[1:] // skip the binary name
	expected := []string{"run", "dev"}
	if len(args) != len(expected) {
		t.Fatalf("args = %v, want %v", args, expected)
	}
	for i, want := range expected {
		if args[i] != want {
			t.Errorf("args[%d] = %q, want %q", i, args[i], want)
		}
	}
}

func TestBuildCmdWithArgs(t *testing.T) {
	cmd := BuildCmd("pnpm", "run", "test", []string{"--coverage", "--watch"})

	args := cmd.Args[1:]
	expected := []string{"run", "test", "--", "--coverage", "--watch"}
	if len(args) != len(expected) {
		t.Fatalf("args = %v, want %v", args, expected)
	}
	for i, want := range expected {
		if args[i] != want {
			t.Errorf("args[%d] = %q, want %q", i, args[i], want)
		}
	}
}

func TestBuildCmdNoSubcmd(t *testing.T) {
	cmd := BuildCmd("make", "", "build", nil)

	args := cmd.Args[1:]
	expected := []string{"build"}
	if len(args) != len(expected) {
		t.Fatalf("args = %v, want %v", args, expected)
	}
	if args[0] != "build" {
		t.Errorf("args[0] = %q, want %q", args[0], "build")
	}
}

func TestBuildCmdDifferentRunners(t *testing.T) {
	for _, runner := range []string{"npm", "yarn", "pnpm", "bun"} {
		cmd := BuildCmd(runner, "run", "build", nil)
		if cmd.Args[0] != runner {
			t.Errorf("runner %s: Args[0] = %q", runner, cmd.Args[0])
		}
	}
}
