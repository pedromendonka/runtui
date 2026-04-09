package runner

import (
	"os/exec"
)

// BuildCmd constructs an exec.Cmd for the given task.
// The runner is the binary (e.g. "npm", "make").
// The subcmd is the runner subcommand (e.g. "run" for npm, empty for make).
// The script is the task/target name.
// args are additional arguments to pass after the script name.
func BuildCmd(runner, subcmd, script string, args []string) *exec.Cmd {
	var cmdArgs []string
	if subcmd != "" {
		cmdArgs = append(cmdArgs, subcmd)
	}
	cmdArgs = append(cmdArgs, script)
	if len(args) > 0 {
		cmdArgs = append(cmdArgs, "--")
		cmdArgs = append(cmdArgs, args...)
	}
	return exec.Command(runner, cmdArgs...)
}
