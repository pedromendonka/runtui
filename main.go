package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pedromendonka/runtui/detector"
	"github.com/pedromendonka/runtui/parser"
	"github.com/pedromendonka/runtui/tui"
)

var version = "dev"

var validRunners = map[string]bool{
	"npm": true, "yarn": true, "pnpm": true, "bun": true,
}

func fatal(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, "runtui: "+format+"\n", args...)
	os.Exit(1)
}

func main() {
	runnerFlag := flag.String("runner", "", "override detected package manager (npm, yarn, pnpm, bun)")
	infoFlag := flag.Bool("info", false, "show full commands without truncation")
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println("runtui " + version)
		return
	}

	dir, err := os.Getwd()
	if err != nil {
		fatal("%v", err)
	}

	projects, err := detector.Detect(dir)
	if err != nil {
		fatal("%v", err)
	}

	if len(projects) == 0 {
		fatal("no supported project files found")
	}

	project := projects[0]

	runner := project.Runner
	if *runnerFlag != "" {
		if !validRunners[*runnerFlag] {
			fatal("unsupported runner %q: must be one of npm, yarn, pnpm, bun", *runnerFlag)
		}
		runner = *runnerFlag
	}

	var p parser.Parser
	switch project.Type {
	case detector.TypePackageJSON:
		p = parser.NewPackageJSON(runner)
	default:
		fatal("unsupported project type: %s", project.Type)
	}

	tasks, runCtx, err := p.Parse(project.Path)
	if err != nil {
		fatal("%v", err)
	}

	if len(tasks) == 0 {
		fatal("no tasks found")
	}

	header := fmt.Sprintf("%s (%s)", project.Type, runner)
	m := tui.New(tasks, header, runCtx, *infoFlag)

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fatal("%v", err)
	}
}
