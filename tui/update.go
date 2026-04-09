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
		m.width = msg.Width
		m.height = msg.Height
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
		if m.filter != "" {
			m.filter = ""
			m.filtered = m.tasks
			m.cursor = 0
			m.offset = 0
		} else {
			m.quitting = true
			return m, tea.Quit
		}

	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.offset {
				m.offset = m.cursor
			}
		}

	case tea.KeyDown:
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
			if vis := m.visibleLines(); m.cursor >= m.offset+vis {
				m.offset = m.cursor - vis + 1
			}
		}

	case tea.KeyEnter:
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			task := m.filtered[m.cursor]
			m.selected = &task
			m.lastRun = nil
			m.phase = phaseArgs
			m.showValidation = false

			if len(task.Args) > 0 {
				m.argInputs = make([]argInput, len(task.Args))
				for i, def := range task.Args {
					m.argInputs[i] = argInput{def: def, value: def.Default}
				}
				m.argCursor = 0
			} else {
				m.simpleArg = ""
			}
		}

	case tea.KeyBackspace:
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.filtered = filterTasks(m.tasks, m.filter)
			m.cursor = 0
			m.offset = 0
		}

	case tea.KeyRunes:
		ch := string(msg.Runes)
		if ch == "q" && m.filter == "" {
			m.quitting = true
			return m, tea.Quit
		}
		m.filter += ch
		m.filtered = filterTasks(m.tasks, m.filter)
		m.cursor = 0
		m.offset = 0
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

	if m.isSimpleMode() {
		return m.handleSimpleArgsKey(msg)
	}
	return m.handleConfigArgsKey(msg)
}

func (m Model) handleSimpleArgsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		m.collectedArgs = collectSimpleArgs(m.simpleArg)
		return m, m.execTask()

	case tea.KeyBackspace:
		if len(m.simpleArg) > 0 {
			m.simpleArg = m.simpleArg[:len(m.simpleArg)-1]
		}

	case tea.KeyRunes:
		m.simpleArg += string(msg.Runes)

	case tea.KeySpace:
		m.simpleArg += " "
	}

	return m, nil
}

func (m Model) handleConfigArgsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		if m.missingRequired() {
			m.showValidation = true
			return m, nil
		}
		m.collectedArgs = m.collectConfigArgs()
		return m, m.execTask()

	case tea.KeyTab, tea.KeyDown:
		m.showValidation = false
		if m.argCursor < len(m.argInputs)-1 {
			m.argCursor++
		}

	case tea.KeyShiftTab, tea.KeyUp:
		m.showValidation = false
		if m.argCursor > 0 {
			m.argCursor--
		}

	case tea.KeyBackspace:
		v := m.argInputs[m.argCursor].value
		if len(v) > 0 {
			m.argInputs[m.argCursor].value = v[:len(v)-1]
		}

	case tea.KeyRunes:
		m.argInputs[m.argCursor].value += string(msg.Runes)
		m.showValidation = false

	case tea.KeySpace:
		m.argInputs[m.argCursor].value += " "
		m.showValidation = false
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
