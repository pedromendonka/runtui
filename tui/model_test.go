package tui

import (
	"errors"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/pedromendonka/runtui/parser"
)

var npmCtx = parser.RunContext{
	Binary:        "npm",
	Subcmd:        "run",
	ArgSeparator:  "--",
	DisplayPrefix: []string{"npm", "run"},
}

var makeCtx = parser.RunContext{
	Binary:        "make",
	Subcmd:        "",
	ArgSeparator:  "",
	DisplayPrefix: []string{"make"},
}

// --- collectSimpleArgs ---

func TestCollectSimpleArgs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", nil},
		{"whitespace only", "   ", nil},
		{"single token", "dev", []string{"dev"}},
		{"multiple tokens", "--coverage --watch", []string{"--coverage", "--watch"}},
		{"quoted token", `--msg "hello world"`, []string{"--msg", "hello world"}},
		{"multiple quoted", `"a b" "c d"`, []string{"a b", "c d"}},
		{"leading/trailing space", "  dev  ", []string{"dev"}},
		// Documented current behavior: an unterminated quote keeps reading
		// until EOL. The trailing content becomes one arg.
		{"unterminated quote", `--msg "hello world`, []string{"--msg", "hello world"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectSimpleArgs(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("collectSimpleArgs(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// --- filterTasks ---

func TestFilterTasks(t *testing.T) {
	tasks := []parser.Task{
		{Name: "dev"},
		{Name: "build"},
		{Name: "test"},
		{Name: "test:watch"},
		{Name: "DEPLOY"},
	}

	tests := []struct {
		filter string
		want   []string
	}{
		{"", []string{"dev", "build", "test", "test:watch", "DEPLOY"}},
		{"test", []string{"test", "test:watch"}},
		{"TEST", []string{"test", "test:watch"}}, // case-insensitive
		{"deploy", []string{"DEPLOY"}},           // case-insensitive
		{"nope", nil},
	}

	for _, tt := range tests {
		t.Run(tt.filter, func(t *testing.T) {
			got := filterTasks(tasks, tt.filter)
			var gotNames []string
			for _, task := range got {
				gotNames = append(gotNames, task.Name)
			}
			if !reflect.DeepEqual(gotNames, tt.want) {
				t.Errorf("filterTasks(%q) = %v, want %v", tt.filter, gotNames, tt.want)
			}
		})
	}
}

// --- newExecResult ---

func TestNewExecResultSuccess(t *testing.T) {
	r := newExecResult("test", npmCtx, []string{"--coverage"}, nil)

	if !r.ok() {
		t.Error("expected ok()")
	}
	if r.failed() {
		t.Error("expected !failed()")
	}
	if r.command != "npm run test -- --coverage" {
		t.Errorf("command = %q, want %q", r.command, "npm run test -- --coverage")
	}
	if r.exitCode != -1 {
		t.Errorf("exitCode = %d, want -1", r.exitCode)
	}
	if !strings.Contains(r.summary(), "completed successfully") {
		t.Errorf("summary = %q", r.summary())
	}
}

func TestNewExecResultMake(t *testing.T) {
	// Make has no ArgSeparator — display should show "make build VERBOSE=1".
	r := newExecResult("build", makeCtx, []string{"VERBOSE=1"}, nil)

	if r.command != "make build VERBOSE=1" {
		t.Errorf("command = %q, want %q", r.command, "make build VERBOSE=1")
	}
}

func TestNewExecResultExitError(t *testing.T) {
	// Force an ExitError by running a command that fails predictably.
	cmd := exec.Command("sh", "-c", "exit 42")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected err from sh -c 'exit 42'")
	}

	r := newExecResult("test", npmCtx, nil, err)

	if !r.failed() {
		t.Error("expected failed()")
	}
	if r.exitCode != 42 {
		t.Errorf("exitCode = %d, want 42", r.exitCode)
	}
	if !strings.Contains(r.summary(), "exit 42") {
		t.Errorf("summary = %q, want to contain 'exit 42'", r.summary())
	}
}

func TestNewExecResultGenericError(t *testing.T) {
	// Non-ExitError path: exitCode stays at -1.
	r := newExecResult("test", npmCtx, nil, errors.New("launch failed"))

	if !r.failed() {
		t.Error("expected failed()")
	}
	if r.exitCode != -1 {
		t.Errorf("exitCode = %d, want -1", r.exitCode)
	}
	if !strings.Contains(r.summary(), "launch failed") {
		t.Errorf("summary = %q, want to contain 'launch failed'", r.summary())
	}
}

// --- argsState helpers ---

func TestArgsStateIsSimpleMode(t *testing.T) {
	noArgs := &parser.Task{Name: "dev"}
	withArgs := &parser.Task{Name: "env:set", Args: []parser.ArgDef{{Name: "KEY"}}}

	if !(argsState{selected: noArgs}).isSimpleMode() {
		t.Error("task with no Args should be simple mode")
	}
	if (argsState{selected: withArgs}).isSimpleMode() {
		t.Error("task with Args should not be simple mode")
	}
	if (argsState{}).isSimpleMode() {
		t.Error("nil selected should not be simple mode")
	}
}

func TestArgsStateCollectConfigArgs(t *testing.T) {
	a := argsState{
		inputs: []argInput{
			{def: parser.ArgDef{Name: "KEY"}, value: "DATABASE_URL"},
			{def: parser.ArgDef{Name: "VALUE"}, value: ""}, // skipped
			{def: parser.ArgDef{Name: "PORT"}, value: "5432"},
		},
	}
	got := a.collectConfigArgs()
	want := []string{"DATABASE_URL", "5432"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("collectConfigArgs() = %v, want %v", got, want)
	}
}

func TestArgsStateMissingRequired(t *testing.T) {
	tests := []struct {
		name  string
		state argsState
		want  bool
	}{
		{
			"all required filled",
			argsState{inputs: []argInput{
				{def: parser.ArgDef{Name: "KEY", Required: true}, value: "x"},
			}},
			false,
		},
		{
			"required empty",
			argsState{inputs: []argInput{
				{def: parser.ArgDef{Name: "KEY", Required: true}, value: ""},
			}},
			true,
		},
		{
			"optional empty is ok",
			argsState{inputs: []argInput{
				{def: parser.ArgDef{Name: "KEY", Required: false}, value: ""},
			}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.missingRequired(); got != tt.want {
				t.Errorf("missingRequired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArgsStateLabelWidth(t *testing.T) {
	a := argsState{
		inputs: []argInput{
			{def: parser.ArgDef{Name: "KEY"}},
			{def: parser.ArgDef{Name: "ENVIRONMENT"}},
			{def: parser.ArgDef{Name: "TAG"}},
		},
	}
	if got := a.labelWidth(); got != len("ENVIRONMENT") {
		t.Errorf("labelWidth() = %d, want %d", got, len("ENVIRONMENT"))
	}
}

// --- New model constructor ---

func TestNewModelComputesWidths(t *testing.T) {
	tasks := []parser.Task{
		{Name: "dev", Description: "run dev server"},
		{Name: "very-long-task-name"},
		{Name: "test", Description: "this is a longer description"},
	}
	m := New(tasks, "header", npmCtx, false)

	if m.layout.nameWidth != len("very-long-task-name") {
		t.Errorf("nameWidth = %d, want %d", m.layout.nameWidth, len("very-long-task-name"))
	}
	if m.layout.descWidth != len("this is a longer description") {
		t.Errorf("descWidth = %d, want %d", m.layout.descWidth, len("this is a longer description"))
	}
	if m.runCtx.Binary != "npm" {
		t.Errorf("runCtx.Binary = %q, want npm", m.runCtx.Binary)
	}
	if len(m.list.tasks) != 3 || len(m.list.filtered) != 3 {
		t.Errorf("tasks/filtered mismatch: %d/%d", len(m.list.tasks), len(m.list.filtered))
	}
}

// visibleLines sanity check — make sure it derives from the row constants
// rather than reintroducing a magic number.
func TestVisibleLinesFormula(t *testing.T) {
	m := Model{layout: layout{height: 40}}
	base := m.visibleLines()
	expected := 40 - rowsListOverhead
	if base != expected {
		t.Errorf("visibleLines() = %d, want %d", base, expected)
	}

	// With an exec result block visible, overhead grows by rowsExecResult.
	m.list.lastRun = &execResult{}
	withResult := m.visibleLines()
	if withResult != expected-rowsExecResult {
		t.Errorf("visibleLines() with lastRun = %d, want %d", withResult, expected-rowsExecResult)
	}

	// Never go below 1.
	m2 := Model{layout: layout{height: 1}}
	if m2.visibleLines() < 1 {
		t.Errorf("visibleLines() = %d, want >= 1", m2.visibleLines())
	}
}

// Sanity: if newExecResult ever stops using the DisplayPrefix, this fails.
func TestNewExecResultUsesDisplayPrefix(t *testing.T) {
	customCtx := parser.RunContext{
		Binary:        "cargo",
		Subcmd:        "",
		ArgSeparator:  "--",
		DisplayPrefix: []string{"cargo"},
	}
	r := newExecResult("test", customCtx, []string{"--release"}, nil)
	want := "cargo test -- --release"
	if r.command != want {
		t.Errorf("command = %q, want %q", r.command, want)
	}

	// Guard against a regression that might look "close enough" but differ.
	if strings.HasPrefix(r.command, "npm") {
		t.Errorf("command should not start with npm: %q", r.command)
	}
}
