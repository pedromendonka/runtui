package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pedromendonka/runtui/parser"
)

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.phase {
		case phaseList:
			return m.handleListKey(msg)
		case phaseArgs:
			return m.handleArgsKey(msg)
		}
	case tea.WindowSizeMsg:
		m.layout.width = msg.Width
		m.layout.height = msg.Height
	case execDoneMsg:
		m = m.resetToList(msg.err)
	}
	return m, nil
}

// --- List phase ---

func (m Model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		m.quitting = true
		return m, tea.Quit

	case tea.KeyEsc:
		if m.list.filter != "" {
			m.list.filter = ""
			m.list.filtered = m.list.tasks
			m.list.cursor = 0
			m.list.offset = 0
		} else {
			m.quitting = true
			return m, tea.Quit
		}

	case tea.KeyUp:
		if m.list.cursor > 0 {
			m.list.cursor--
			if m.list.cursor < m.list.offset {
				m.list.offset = m.list.cursor
			}
		}

	case tea.KeyDown:
		if m.list.cursor < len(m.list.filtered)-1 {
			m.list.cursor++
			if vis := m.visibleLines(); m.list.cursor >= m.list.offset+vis {
				m.list.offset = m.list.cursor - vis + 1
			}
		}

	case tea.KeyEnter:
		if len(m.list.filtered) > 0 && m.list.cursor < len(m.list.filtered) {
			task := m.list.filtered[m.list.cursor]
			m.args.selected = &task
			m.list.lastRun = nil
			m.phase = phaseArgs
			m.args.showValidation = false

			if len(task.Args) > 0 {
				m.args.inputs = make([]argInput, len(task.Args))
				for i, def := range task.Args {
					m.args.inputs[i] = argInput{def: def, value: def.Default}
				}
				m.args.cursor = 0
			} else {
				m.args.simpleArg = ""
			}
		}

	case tea.KeyBackspace:
		if len(m.list.filter) > 0 {
			m.list.filter = m.list.filter[:len(m.list.filter)-1]
			m.list.filtered = filterTasks(m.list.tasks, m.list.filter)
			m.list.cursor = 0
			m.list.offset = 0
		}

	case tea.KeyRunes:
		ch := string(msg.Runes)
		if ch == "q" && m.list.filter == "" {
			m.quitting = true
			return m, tea.Quit
		}
		m.list.filter += ch
		m.list.filtered = filterTasks(m.list.tasks, m.list.filter)
		m.list.cursor = 0
		m.list.offset = 0
	}

	return m, nil
}

// --- Args phase ---

func (m Model) handleArgsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC {
		m.quitting = true
		return m, tea.Quit
	}

	if msg.Type == tea.KeyEsc {
		m = m.cancelToList()
		return m, nil
	}

	if m.args.isSimpleMode() {
		return m.handleSimpleArgsKey(msg)
	}
	return m.handleConfigArgsKey(msg)
}

func (m Model) handleSimpleArgsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		m.args.collected = collectSimpleArgs(m.args.simpleArg)
		return m, m.execTask()

	case tea.KeyBackspace:
		if len(m.args.simpleArg) > 0 {
			m.args.simpleArg = m.args.simpleArg[:len(m.args.simpleArg)-1]
		}

	case tea.KeyRunes:
		m.args.simpleArg += string(msg.Runes)

	case tea.KeySpace:
		m.args.simpleArg += " "
	}

	return m, nil
}

func (m Model) handleConfigArgsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		if m.args.missingRequired() {
			m.args.showValidation = true
			return m, nil
		}
		m.args.collected = m.args.collectConfigArgs()
		return m, m.execTask()

	case tea.KeyTab, tea.KeyDown:
		m.args.showValidation = false
		if m.args.cursor < len(m.args.inputs)-1 {
			m.args.cursor++
		}

	case tea.KeyShiftTab, tea.KeyUp:
		m.args.showValidation = false
		if m.args.cursor > 0 {
			m.args.cursor--
		}

	case tea.KeyBackspace:
		v := m.args.inputs[m.args.cursor].value
		if len(v) > 0 {
			m.args.inputs[m.args.cursor].value = v[:len(v)-1]
		}

	case tea.KeyRunes:
		m.args.inputs[m.args.cursor].value += string(msg.Runes)
		m.args.showValidation = false

	case tea.KeySpace:
		m.args.inputs[m.args.cursor].value += " "
		m.args.showValidation = false
	}

	return m, nil
}

func filterTasks(tasks []parser.Task, filter string) []parser.Task {
	if filter == "" {
		return tasks
	}
	lower := strings.ToLower(filter)
	var result []parser.Task
	for _, t := range tasks {
		if strings.Contains(strings.ToLower(t.Name), lower) {
			result = append(result, t)
		}
	}
	return result
}
