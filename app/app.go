// Package app holds the orchestration layer of runtui: flag parsing,
// project detection, parser selection, and TUI boot. It is separated
// from main.go so that end-to-end behavior is unit-testable via
// app.Run(args, stdout, stderr, version) without mounting a real TUI.
package app

import (
	"flag"
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pedromendonka/runtui/detector"
	"github.com/pedromendonka/runtui/tui"
)

// Exit codes returned by Run.
const (
	ExitOK    = 0
	ExitError = 1
)

// validRunners is the allow-list for the --runner override flag. It is
// intentionally strict to prevent arbitrary command execution.
var validRunners = map[string]bool{
	"npm": true, "yarn": true, "pnpm": true, "bun": true,
}

// Run is the process entry point. It parses args, detects the project,
// loads the parser, and launches the TUI. Errors are reported to stderr
// and returned as a non-zero exit code — the caller (main.go) is expected
// to pass the return value to os.Exit.
//
// Separating this from main lets tests exercise flag handling, error
// paths, and subcommand routing without touching os.Exit or os.Args.
func Run(args []string, stdout, stderr io.Writer, version string) int {
	fs := flag.NewFlagSet("runtui", flag.ContinueOnError)
	fs.SetOutput(stderr)

	runnerFlag := fs.String("runner", "", "override detected package manager (npm, yarn, pnpm, bun)")
	infoFlag := fs.Bool("info", false, "show full commands without truncation")
	versionFlag := fs.Bool("version", false, "print version and exit")

	if err := fs.Parse(args); err != nil {
		// flag package already printed the error via fs.SetOutput.
		return ExitError
	}

	if *versionFlag {
		fmt.Fprintln(stdout, "runtui "+version)
		return ExitOK
	}

	if *runnerFlag != "" && !validRunners[*runnerFlag] {
		fmt.Fprintf(stderr, "runtui: unsupported runner %q: must be one of npm, yarn, pnpm, bun\n", *runnerFlag)
		return ExitError
	}

	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "runtui: %v\n", err)
		return ExitError
	}

	projects, err := detector.Detect(dir)
	if err != nil {
		fmt.Fprintf(stderr, "runtui: %v\n", err)
		return ExitError
	}
	if len(projects) == 0 {
		fmt.Fprintln(stderr, "runtui: no supported project files found")
		return ExitError
	}

	project := projects[0]

	runner := project.Runner
	if *runnerFlag != "" {
		runner = *runnerFlag
	}

	p, err := parserFor(project.Type, runner)
	if err != nil {
		fmt.Fprintf(stderr, "runtui: %v\n", err)
		return ExitError
	}

	tasks, runCtx, err := p.Parse(project.Path)
	if err != nil {
		fmt.Fprintf(stderr, "runtui: %v\n", err)
		return ExitError
	}
	if len(tasks) == 0 {
		fmt.Fprintln(stderr, "runtui: no tasks found")
		return ExitError
	}

	header := fmt.Sprintf("%s (%s)", project.Type, runner)
	m := tui.New(tasks, header, runCtx, *infoFlag)

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintf(stderr, "runtui: %v\n", err)
		return ExitError
	}
	return ExitOK
}
