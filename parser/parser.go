package parser

// Task represents a runnable task extracted from a project config file.
type Task struct {
	Name        string
	Command     string
	Description string
	Args        []ArgDef
}

// ArgDef defines a structured argument that a task accepts.
type ArgDef struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Hint     string `json:"hint,omitempty"`
	Default  string `json:"default,omitempty"`
}

// RunContext describes how to invoke a task discovered by a parser.
// It fully captures the shape of the command line so that the TUI and
// runner do not need to know whether the project is npm, make, cargo, etc.
//
// Example for npm:
//
//	RunContext{Binary: "npm", Subcmd: "run", ArgSeparator: "--", DisplayPrefix: []string{"npm", "run"}}
//
// Example for make:
//
//	RunContext{Binary: "make", Subcmd: "", ArgSeparator: "", DisplayPrefix: []string{"make"}}
type RunContext struct {
	// Binary is the executable to invoke (e.g. "npm", "make", "cargo").
	Binary string
	// Subcmd is an optional subcommand placed before the script/target name.
	// "" means no subcommand (e.g. make). "run" for npm/yarn/pnpm/bun.
	Subcmd string
	// ArgSeparator is placed between the task name and user arguments.
	// "--" for npm family (so args reach the script, not the runner).
	// "" for make and most other tools.
	ArgSeparator string
	// DisplayPrefix is what the execution result box shows before the task
	// name. Usually []string{Binary} or []string{Binary, Subcmd}.
	DisplayPrefix []string
}

// Parser extracts tasks from a project config file and returns the
// RunContext describing how to invoke those tasks.
type Parser interface {
	Parse(path string) ([]Task, RunContext, error)
}
