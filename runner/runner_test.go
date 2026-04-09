package runner

import (
	"testing"

	"github.com/pedromendonka/runtui/parser"
)

var (
	npmCtx = parser.RunContext{
		Binary:        "npm",
		Subcmd:        "run",
		ArgSeparator:  "--",
		DisplayPrefix: []string{"npm", "run"},
	}
	makeCtx = parser.RunContext{
		Binary:        "make",
		Subcmd:        "",
		ArgSeparator:  "",
		DisplayPrefix: []string{"make"},
	}
)

func TestBuildCmdNoArgs(t *testing.T) {
	cmd := BuildCmd(npmCtx, "dev", nil)

	if cmd.Path == "" {
		t.Fatal("expected non-empty path")
	}

	args := cmd.Args[1:] // skip the binary name
	expected := []string{"run", "dev"}
	assertArgs(t, args, expected)
}

func TestBuildCmdWithArgs(t *testing.T) {
	ctx := parser.RunContext{
		Binary:        "pnpm",
		Subcmd:        "run",
		ArgSeparator:  "--",
		DisplayPrefix: []string{"pnpm", "run"},
	}
	cmd := BuildCmd(ctx, "test", []string{"--coverage", "--watch"})

	args := cmd.Args[1:]
	expected := []string{"run", "test", "--", "--coverage", "--watch"}
	assertArgs(t, args, expected)
}

func TestBuildCmdNoSubcmd(t *testing.T) {
	cmd := BuildCmd(makeCtx, "build", nil)

	args := cmd.Args[1:]
	expected := []string{"build"}
	assertArgs(t, args, expected)
}

func TestBuildCmdMakeWithArgs(t *testing.T) {
	// Make has no ArgSeparator — args are passed directly.
	cmd := BuildCmd(makeCtx, "build", []string{"VERBOSE=1"})

	args := cmd.Args[1:]
	expected := []string{"build", "VERBOSE=1"}
	assertArgs(t, args, expected)
}

func TestBuildCmdDifferentRunners(t *testing.T) {
	for _, runner := range []string{"npm", "yarn", "pnpm", "bun"} {
		ctx := parser.RunContext{
			Binary:        runner,
			Subcmd:        "run",
			ArgSeparator:  "--",
			DisplayPrefix: []string{runner, "run"},
		}
		cmd := BuildCmd(ctx, "build", nil)
		if cmd.Args[0] != runner {
			t.Errorf("runner %s: Args[0] = %q", runner, cmd.Args[0])
		}
	}
}

func assertArgs(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("args = %v, want %v", got, want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("args[%d] = %q, want %q", i, got[i], w)
		}
	}
}
