package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pedromendonka/runtui/parser"
)

// Shared fixtures for update tests.
var threeTasks = []parser.Task{
	{Name: "build", Command: "go build"},
	{Name: "dev", Command: "next dev"},
	{Name: "test", Command: "jest"},
}

var envTask = parser.Task{
	Name:    "env:set",
	Command: "dotenvx set",
	Args: []parser.ArgDef{
		{Name: "KEY", Required: true},
		{Name: "VALUE", Required: true},
	},
}

// newTestModel builds a Model with a reasonable layout so visibleLines > 0.
func newTestModel(tasks []parser.Task) Model {
	m := New(tasks, "test (npm)", npmCtx, false)
	m.layout.width = 120
	m.layout.height = 40
	return m
}

// sendKey sends a single key message to the current model and returns the
// updated model and any emitted command.
func sendKey(t *testing.T, m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	t.Helper()
	updated, cmd := m.Update(msg)
	mm, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned non-Model type %T", updated)
	}
	return mm, cmd
}

// --- List phase: navigation ---

func TestListDownMovesCursor(t *testing.T) {
	m := newTestModel(threeTasks)
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyDown})
	if m.list.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m.list.cursor)
	}
}

func TestListDownClampsAtEnd(t *testing.T) {
	m := newTestModel(threeTasks)
	for range 10 {
		m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyDown})
	}
	if m.list.cursor != len(threeTasks)-1 {
		t.Errorf("cursor = %d, want %d", m.list.cursor, len(threeTasks)-1)
	}
}

func TestListUpClampsAtZero(t *testing.T) {
	m := newTestModel(threeTasks)
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyUp})
	if m.list.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.list.cursor)
	}
}

// --- List phase: filtering ---

func TestListFilterByRune(t *testing.T) {
	m := newTestModel(threeTasks)
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})

	if m.list.filter != "d" {
		t.Errorf("filter = %q, want %q", m.list.filter, "d")
	}
	// "build" and "dev" both contain 'd'.
	if len(m.list.filtered) != 2 {
		t.Errorf("filtered = %d, want 2", len(m.list.filtered))
	}
}

func TestListBackspaceDeletesFilter(t *testing.T) {
	m := newTestModel(threeTasks)
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyBackspace})

	if m.list.filter != "d" {
		t.Errorf("filter = %q, want %q", m.list.filter, "d")
	}
}

func TestListEscClearsFilter(t *testing.T) {
	m := newTestModel(threeTasks)
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})

	m, cmd := sendKey(t, m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.list.filter != "" {
		t.Errorf("filter = %q, want empty", m.list.filter)
	}
	if m.quitting {
		t.Error("should not be quitting when esc clears a filter")
	}
	if cmd != nil {
		t.Error("no command should be emitted")
	}
}

func TestListEscQuitsWhenFilterEmpty(t *testing.T) {
	m := newTestModel(threeTasks)
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyEsc})
	if !m.quitting {
		t.Error("should be quitting when esc fires on empty filter")
	}
}

func TestListQQuitsWhenFilterEmpty(t *testing.T) {
	m := newTestModel(threeTasks)
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if !m.quitting {
		t.Error("'q' should quit with empty filter")
	}
}

func TestListQFiltersWhenFilterNonEmpty(t *testing.T) {
	m := newTestModel(threeTasks)
	// Start a filter, then type 'q' — it should append, not quit.
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	if m.quitting {
		t.Error("'q' with non-empty filter should not quit")
	}
	if m.list.filter != "dq" {
		t.Errorf("filter = %q, want %q", m.list.filter, "dq")
	}
}

func TestListCtrlCQuits(t *testing.T) {
	m := newTestModel(threeTasks)
	m, cmd := sendKey(t, m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if !m.quitting {
		t.Error("ctrl-c should quit")
	}
	if cmd == nil {
		t.Error("ctrl-c should emit tea.Quit command")
	}
}

// --- List phase: Enter transitions to args phase ---

func TestListEnterTransitionsToSimpleArgs(t *testing.T) {
	m := newTestModel(threeTasks)
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.phase != phaseArgs {
		t.Errorf("phase = %v, want phaseArgs", m.phase)
	}
	if m.args.selected == nil {
		t.Fatal("args.selected should be set")
	}
	if m.args.selected.Name != "build" {
		t.Errorf("selected.Name = %q, want build", m.args.selected.Name)
	}
	if len(m.args.inputs) != 0 {
		t.Error("simple mode should have no inputs")
	}
}

func TestListEnterTransitionsToConfigArgsWithDefaults(t *testing.T) {
	taskWithDefault := parser.Task{
		Name: "deploy",
		Args: []parser.ArgDef{
			{Name: "ENV", Required: true, Default: "staging"},
			{Name: "TAG", Required: false, Default: "latest"},
		},
	}
	m := newTestModel([]parser.Task{taskWithDefault})
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	if len(m.args.inputs) != 2 {
		t.Fatalf("inputs = %d, want 2", len(m.args.inputs))
	}
	if m.args.inputs[0].value != "staging" {
		t.Errorf("inputs[0].value = %q, want staging", m.args.inputs[0].value)
	}
	if m.args.inputs[1].value != "latest" {
		t.Errorf("inputs[1].value = %q, want latest", m.args.inputs[1].value)
	}
}

// --- Args phase: simple mode ---

func TestSimpleArgsTypesAndBackspaces(t *testing.T) {
	m := newTestModel(threeTasks)
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyEnter}) // enter args

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	if m.args.simpleArg != "--v" {
		t.Errorf("simpleArg = %q, want --v", m.args.simpleArg)
	}

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyBackspace})
	if m.args.simpleArg != "--" {
		t.Errorf("after backspace: simpleArg = %q, want --", m.args.simpleArg)
	}

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeySpace})
	if m.args.simpleArg != "-- " {
		t.Errorf("after space: simpleArg = %q, want '-- '", m.args.simpleArg)
	}
}

func TestSimpleArgsEscReturnsToList(t *testing.T) {
	m := newTestModel(threeTasks)
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.phase != phaseList {
		t.Errorf("phase = %v, want phaseList", m.phase)
	}
	if m.args.selected != nil {
		t.Error("selected should be cleared")
	}
	if m.args.simpleArg != "" {
		t.Errorf("simpleArg = %q, want empty", m.args.simpleArg)
	}
}

func TestSimpleArgsEnterExecutes(t *testing.T) {
	m := newTestModel(threeTasks)
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyEnter}) // → args
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	m, cmd := sendKey(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	// Enter in args phase should emit an exec cmd.
	if cmd == nil {
		t.Error("expected a tea.Cmd from Enter in args phase")
	}
	if len(m.args.collected) != 1 || m.args.collected[0] != "x" {
		t.Errorf("collected = %v, want [x]", m.args.collected)
	}
}

// --- Args phase: config mode ---

func TestConfigArgsEnterBlockedByMissingRequired(t *testing.T) {
	m := newTestModel([]parser.Task{envTask})
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyEnter}) // → args

	m, cmd := sendKey(t, m, tea.KeyMsg{Type: tea.KeyEnter}) // try to run

	if cmd != nil {
		t.Error("Enter with missing required should not emit a command")
	}
	if !m.args.showValidation {
		t.Error("showValidation should be true")
	}
	if m.phase != phaseArgs {
		t.Error("should still be in args phase")
	}
}

func TestConfigArgsTabMovesCursor(t *testing.T) {
	m := newTestModel([]parser.Task{envTask})
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyEnter}) // → args

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyTab})
	if m.args.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m.args.cursor)
	}

	// Tab past the end clamps.
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyTab})
	if m.args.cursor != 1 {
		t.Errorf("cursor after second tab = %d, want clamped at 1", m.args.cursor)
	}

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.args.cursor != 0 {
		t.Errorf("cursor after shift-tab = %d, want 0", m.args.cursor)
	}
}

func TestConfigArgsTypingFillsActiveField(t *testing.T) {
	m := newTestModel([]parser.Task{envTask})
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyEnter}) // → args

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("K")})
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("E")})
	if m.args.inputs[0].value != "KE" {
		t.Errorf("inputs[0].value = %q, want KE", m.args.inputs[0].value)
	}
	if m.args.inputs[1].value != "" {
		t.Errorf("inputs[1].value = %q, want empty", m.args.inputs[1].value)
	}
}

func TestConfigArgsBackspaceDeletesActiveField(t *testing.T) {
	m := newTestModel([]parser.Task{envTask})
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("K")})
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("E")})

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyBackspace})
	if m.args.inputs[0].value != "K" {
		t.Errorf("value after backspace = %q, want K", m.args.inputs[0].value)
	}
}

func TestConfigArgsFullFlowRuns(t *testing.T) {
	m := newTestModel([]parser.Task{envTask})
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyEnter}) // → args

	// Fill KEY
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("K")})
	// Tab to VALUE
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyTab})
	// Fill VALUE
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("V")})

	_, cmd := sendKey(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("expected exec cmd after filling all required fields")
	}
}

// --- Window size + exec done ---

func TestWindowSizeUpdatesLayout(t *testing.T) {
	m := newTestModel(threeTasks)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	mm := updated.(Model)

	if mm.layout.width != 80 {
		t.Errorf("width = %d, want 80", mm.layout.width)
	}
	if mm.layout.height != 24 {
		t.Errorf("height = %d, want 24", mm.layout.height)
	}
}
