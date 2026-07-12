package runner

import (
	"os/exec"

	"github.com/pedromendonka/runtui/parser"
)

// BuildCmd constructs an exec.Cmd for a task using the given RunContext.
// The script is the task/target name. args are additional arguments to
// pass to the task (typically collected from the arg prompt).
//
// Assembly:
//
//	<Binary> [Subcmd] <script> [ArgSeparator] [args...]
//
// For npm: "npm run test -- --coverage"
// For make: "make test --coverage"
func BuildCmd(ctx parser.RunContext, script string, args []string) *exec.Cmd {
	var cmdArgs []string
	if ctx.Subcmd != "" {
		cmdArgs = append(cmdArgs, ctx.Subcmd)
	}
	cmdArgs = append(cmdArgs, script)
	if len(args) > 0 {
		if ctx.ArgSeparator != "" {
			cmdArgs = append(cmdArgs, ctx.ArgSeparator)
		}
		cmdArgs = append(cmdArgs, args...)
	}
	return exec.Command(ctx.Binary, cmdArgs...)
}
