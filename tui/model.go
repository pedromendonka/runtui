package tui

import (
	"errors"
	"fmt"
	"os/exec"
	"slices"
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

// Vertical space budget. Each constant is the number of terminal rows the
// corresponding block consumes when rendered. visibleLines sums these to
// derive how many task rows fit in the remaining space — replacing the
// previous magic number 15.
const (
	rowsBanner     = 7 // bordered banner box
	rowsProject    = 1 // project info line
	rowsBlank1     = 1 // blank between project info and header
	rowsColHeader  = 1
	rowsColSep     = 1
	rowsBlank2     = 1 // blank before filter
	rowsFilter     = 1
	rowsBlank3     = 1 // blank between filter and footer
	rowsFooter     = 1
	rowsExecResult = 4 // bordered result box + spacing

	rowsListOverhead = rowsBanner + rowsProject + rowsBlank1 +
		rowsColHeader + rowsColSep + rowsBlank2 +
		rowsFilter + rowsBlank3 + rowsFooter + 2 // +2 safety margin
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

func newExecResult(taskName string, ctx parser.RunContext, args []string, err error) execResult {
	parts := slices.Clone(ctx.DisplayPrefix)
	parts = append(parts, taskName)
	if len(args) > 0 {
		if ctx.ArgSeparator != "" {
			parts = append(parts, ctx.ArgSeparator)
		}
		parts = append(parts, args...)
	}

	r := execResult{
		taskName: taskName,
		command:  strings.Join(parts, " "),
		err:      err,
		exitCode: -1,
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
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

// listState holds browse/filter state for the task list phase.
type listState struct {
	tasks    []parser.Task
	filtered []parser.Task
	filter   string
	cursor   int
	offset   int
	lastRun  *execResult
}

// argsState holds input state for the arguments phase.
type argsState struct {
	selected       *parser.Task
	inputs         []argInput
	cursor         int
	simpleArg      string
	collected      []string
	showValidation bool
}

// isSimpleMode reports whether the current task uses the free-form prompt
// (i.e. has no structured args config).
func (a argsState) isSimpleMode() bool {
	return a.selected != nil && len(a.selected.Args) == 0
}

// collectConfigArgs returns non-empty values from structured fields,
// in declaration order.
func (a argsState) collectConfigArgs() []string {
	var args []string
	for _, inp := range a.inputs {
		if inp.value != "" {
			args = append(args, inp.value)
		}
	}
	return args
}

// missingRequired reports whether any required field is still empty.
func (a argsState) missingRequired() bool {
	for _, inp := range a.inputs {
		if inp.def.Required && inp.value == "" {
			return true
		}
	}
	return false
}

// labelWidth returns the widest arg label, used to align labels in the form.
func (a argsState) labelWidth() int {
	w := 0
	for _, inp := range a.inputs {
		if len(inp.def.Name) > w {
			w = len(inp.def.Name)
		}
	}
	return w
}

// layout holds rendering-time terminal dimensions and precomputed widths.
type layout struct {
	width     int
	height    int
	nameWidth int
	descWidth int
}

// Model is the Bubble Tea model for the task list TUI.
type Model struct {
	list     listState
	args     argsState
	layout   layout
	header   string
	runCtx   parser.RunContext
	info     bool
	phase    phase
	quitting bool
}

// New creates a Model ready to display the given tasks.
// The runCtx fully describes how to invoke a selected task.
func New(tasks []parser.Task, header string, runCtx parser.RunContext, info bool) Model {
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
		list: listState{
			tasks:    tasks,
			filtered: tasks,
		},
		layout: layout{
			nameWidth: nw,
			descWidth: dw,
		},
		header: header,
		runCtx: runCtx,
		info:   info,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// visibleLines returns how many task rows fit in the current viewport.
// Derived from rowsListOverhead + an extra allotment when the last-run
// result box is visible.
func (m Model) visibleLines() int {
	overhead := rowsListOverhead
	if m.list.lastRun != nil {
		overhead += rowsExecResult
	}
	lines := m.layout.height - overhead
	if lines < 1 {
		lines = 1
	}
	return lines
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
	cmd := runner.BuildCmd(m.runCtx, m.args.selected.Name, m.args.collected)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return execDoneMsg{err: err}
	})
}

// cancelToList returns to the list without recording an execution.
func (m Model) cancelToList() Model {
	m.phase = phaseList
	m.args = argsState{}
	return m
}

// resetToList clears args state, records the execution result, and returns to the list.
func (m Model) resetToList(err error) Model {
	result := newExecResult(m.args.selected.Name, m.runCtx, m.args.collected, err)
	m.list.lastRun = &result
	m.phase = phaseList
	m.args = argsState{}
	return m
}
