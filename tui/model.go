package tui

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pedromendonka/runtui/parser"
	"github.com/pedromendonka/runtui/runner"
)

type phase int

const (
	phaseList phase = iota
	phaseArgs
)

// argInput holds user input for a single argument field.
type argInput struct {
	def   parser.ArgDef
	value string
}

// execDoneMsg is sent when a subprocess finishes.
type execDoneMsg struct{ err error }

// execResult captures the outcome of the last command execution.
type execResult struct {
	taskName string
	command  string // full command as displayed (e.g. "npm run test -- --coverage")
	err      error
	exitCode int // -1 if not an ExitError
}

func newExecResult(taskName, runner string, args []string, err error) execResult {
	// Build the display command.
	parts := []string{runner, "run", taskName}
	if len(args) > 0 {
		parts = append(parts, "--")
		parts = append(parts, args...)
	}

	r := execResult{
		taskName: taskName,
		command:  strings.Join(parts, " "),
		err:      err,
		exitCode: -1,
	}

	if err != nil {
		// Go 1.26: generic type-safe error assertion.
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			r.exitCode = exitErr.ExitCode()
		}
	}

	return r
}

func (r execResult) ok() bool     { return r.err == nil }
func (r execResult) failed() bool { return r.err != nil }

func (r execResult) summary() string {
	if r.ok() {
		return fmt.Sprintf("%s completed successfully", r.taskName)
	}
	if r.exitCode >= 0 {
		return fmt.Sprintf("%s failed (exit %d)", r.taskName, r.exitCode)
	}
	return fmt.Sprintf("%s failed: %v", r.taskName, r.err)
}

// Model is the Bubble Tea model for the task list TUI.
type Model struct {
	tasks     []parser.Task
	filtered  []parser.Task
	filter    string
	cursor    int
	offset    int
	width     int
	height    int
	header    string
	nameWidth int
	descWidth int
	runner    string
	info      bool
	selected  *parser.Task
	quitting  bool
	lastRun   *execResult

	// Args phase state.
	phase          phase
	argInputs      []argInput
	argCursor      int
	simpleArg      string
	collectedArgs  []string
	showValidation bool
}

// New creates a Model ready to display the given tasks.
// The runner is the package manager command (e.g. "npm", "pnpm").
func New(tasks []parser.Task, header, runner string, info bool) Model {
	nw, dw := 0, 0
	for _, t := range tasks {
		if len(t.Name) > nw {
			nw = len(t.Name)
		}
		if len(t.Description) > dw {
			dw = len(t.Description)
		}
	}
	return Model{
		tasks:     tasks,
		filtered:  tasks,
		header:    header,
		runner:    runner,
		info:      info,
		nameWidth: nw,
		descWidth: dw,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) visibleLines() int {
	// banner box(7) + project info(1) + blank(1) + filter(1) + blank(1) + col header(1) + col sep(1) + tasks(N) + blank(1) + footer(1)
	overhead := 15
	if m.lastRun != nil {
		overhead += 4
	}
	lines := m.height - overhead
	if lines < 1 {
		lines = 1
	}
	return lines
}

func (m Model) isSimpleMode() bool {
	return m.selected != nil && len(m.selected.Args) == 0
}

func (m Model) collectConfigArgs() []string {
	var args []string
	for _, inp := range m.argInputs {
		if inp.value != "" {
			args = append(args, inp.value)
		}
	}
	return args
}

func (m Model) missingRequired() bool {
	for _, inp := range m.argInputs {
		if inp.def.Required && inp.value == "" {
			return true
		}
	}
	return false
}

func (m Model) argLabelWidth() int {
	w := 0
	for _, inp := range m.argInputs {
		if len(inp.def.Name) > w {
			w = len(inp.def.Name)
		}
	}
	return w
}

// collectSimpleArgs splits the free-form input into arguments with basic quote support.
// Supports double-quoted strings: --msg "hello world" → ["--msg", "hello world"]
func collectSimpleArgs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var args []string
	var current strings.Builder
	inQuote := false

	for _, r := range raw {
		switch {
		case r == '"':
			inQuote = !inQuote
		case r == ' ' && !inQuote:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

// execTask returns a tea.Cmd that runs the selected task as a subprocess.
func (m Model) execTask() tea.Cmd {
	cmd := runner.BuildCmd(m.runner, "run", m.selected.Name, m.collectedArgs)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return execDoneMsg{err: err}
	})
}

// cancelToList returns to the list without recording an execution.
func (m Model) cancelToList() Model {
	m.phase = phaseList
	m.selected = nil
	m.argInputs = nil
	m.argCursor = 0
	m.simpleArg = ""
	m.collectedArgs = nil
	m.showValidation = false
	return m
}

// resetToList clears args state, records the execution result, and returns to the list.
func (m Model) resetToList(err error) Model {
	result := newExecResult(m.selected.Name, m.runner, m.collectedArgs, err)
	m.lastRun = &result
	m.phase = phaseList
	m.selected = nil
	m.argInputs = nil
	m.argCursor = 0
	m.simpleArg = ""
	m.collectedArgs = nil
	m.showValidation = false
	return m
}
