package tui

import (
	"fmt"
	"strings"

	"github.com/pedromendonka/runtui/parser"
)

const (
	bannerArtL1 = "╦═╗  ╦ ╦  ╔╗ ╔  ╔╦╗  ╦ ╦  ╦"
	bannerArtL2 = "╠╦╝  ║ ║  ║╚╗║   ║   ║ ║  ║"
	bannerArtL3 = "║╚╗  ║ ║  ║ ╚║   ║   ║ ║  ║"
	bannerArtL4 = "╩ ╚  ╚═╝  ╝  ╚   ╩   ╚═╝  ╩"
)

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting || m.layout.width == 0 || m.layout.height == 0 {
		return ""
	}
	switch m.phase {
	case phaseArgs:
		return m.viewArgs()
	default:
		return m.viewList()
	}
}

// renderBanner writes the banner (full art on wide terminals, short label
// on narrow ones) followed by a newline. Shared between list and args views.
func (m Model) renderBanner(b *strings.Builder, withTagline bool) {
	if m.layout.width >= 40 {
		lines := []string{
			styleBannerL1.Render(bannerArtL1),
			styleBannerL2.Render(bannerArtL2),
			styleBannerL3.Render(bannerArtL3),
			styleBannerL4.Render(bannerArtL4),
		}
		if withTagline {
			lines = append(lines, styleTagline.Render("One TUI to run them all."))
		}
		b.WriteString(styleBannerBox.Render(strings.Join(lines, "\n")))
		b.WriteString("\n")
	} else {
		b.WriteString("  ")
		b.WriteString(styleBannerL3.Render("runtui"))
		b.WriteString("\n")
	}
}

// --- List view ---

func (m Model) viewList() string {
	var b strings.Builder

	m.renderBanner(&b, true)

	// Project info.
	b.WriteString("  ")
	b.WriteString(styleProjectInfo.Render(m.header))
	b.WriteString("\n\n")

	// Column headers.
	hasDesc := m.layout.descWidth > 0
	descW, cmdW := m.columnWidths()

	b.WriteString("    ")
	b.WriteString(styleColHeader.Render(fmt.Sprintf("%-*s", m.layout.nameWidth, "Task")))
	if hasDesc {
		b.WriteString("  ")
		b.WriteString(styleColHeader.Render(fmt.Sprintf("%-*s", descW, "Description")))
	}
	b.WriteString("  ")
	b.WriteString(styleColHeader.Render("Command"))
	b.WriteString("\n")

	sepLen := m.layout.width - 4
	if sepLen < 20 {
		sepLen = 20
	}
	b.WriteString("    ")
	b.WriteString(styleColSep.Render(strings.Repeat("─", sepLen)))
	b.WriteString("\n")

	// Task rows.
	if len(m.list.filtered) == 0 {
		b.WriteString("    ")
		b.WriteString(styleEmpty.Render("no matching tasks"))
		b.WriteString("\n")
	} else {
		vis := m.visibleLines()
		end := m.list.offset + vis
		if end > len(m.list.filtered) {
			end = len(m.list.filtered)
		}
		for i := m.list.offset; i < end; i++ {
			m.renderTask(&b, m.list.filtered[i], i == m.list.cursor, descW, cmdW)
		}
	}

	// Filter.
	b.WriteString("\n  ")
	b.WriteString(styleFilterLabel.Render("/ "))
	b.WriteString(styleFilterText.Render(m.list.filter))
	b.WriteString(styleFilterCur.Render("│"))
	if len(m.list.filter) > 0 {
		b.WriteString(styleTaskCount.Render(fmt.Sprintf("  %d/%d", len(m.list.filtered), len(m.list.tasks))))
	}

	// Footer.
	b.WriteString("\n  ")
	b.WriteString(renderFooter("↑↓", "navigate", "/", "filter", "enter", "run", "q", "quit"))

	// Last execution result.
	if m.list.lastRun != nil {
		b.WriteString("\n\n")
		m.renderExecResult(&b)
	}

	return b.String()
}

// columnWidths computes the description and command column widths for the current terminal.
func (m Model) columnWidths() (descW, cmdW int) {
	if m.layout.descWidth == 0 {
		// No descriptions — all remaining space goes to command.
		cmdW = m.layout.width - 8 - m.layout.nameWidth
		return 0, cmdW
	}

	descW = m.layout.descWidth
	if descW > 60 {
		descW = 60
	}
	// cursor(4) + name(nameWidth) + gap(2) + desc(descW) + gap(2) + cmd
	cmdW = m.layout.width - 10 - m.layout.nameWidth - descW
	if cmdW < 15 {
		// Squeeze the description column to give command at least 15 chars.
		descW = m.layout.width - 10 - m.layout.nameWidth - 15
		if descW < 0 {
			descW = 0
		}
		cmdW = m.layout.width - 10 - m.layout.nameWidth - descW
	}
	return descW, cmdW
}

func (m Model) renderTask(b *strings.Builder, t parser.Task, selected bool, descW, cmdW int) {
	cursor := "    "
	nameStyle := styleName
	dStyle := styleDesc
	if selected {
		cursor = "  " + styleCursor.Render("→") + " "
		nameStyle = styleSelName
		dStyle = styleDescSel
	}

	name := nameStyle.Render(fmt.Sprintf("%-*s", m.layout.nameWidth, t.Name))

	cmd := t.Command
	if !m.info && cmdW > 0 {
		cmd = truncate(cmd, cmdW)
	}

	b.WriteString(cursor)
	b.WriteString(name)

	if descW > 0 {
		desc := truncate(t.Description, descW)
		b.WriteString("  ")
		b.WriteString(dStyle.Render(fmt.Sprintf("%-*s", descW, desc)))
	}

	b.WriteString("  ")
	b.WriteString(styleCmd.Render(cmd))
	b.WriteString("\n")
}

// --- Args view ---

func (m Model) viewArgs() string {
	var b strings.Builder

	m.renderBanner(&b, false)

	// Task name + command.
	b.WriteString("  ")
	b.WriteString(styleProjectInfo.Render(m.args.selected.Name))
	b.WriteString("  ")
	b.WriteString(styleSubHeader.Render(m.args.selected.Command))
	b.WriteString("\n\n")

	if m.args.isSimpleMode() {
		m.renderSimpleArgs(&b)
	} else {
		m.renderConfigArgs(&b)
	}

	return b.String()
}

func (m Model) renderSimpleArgs(b *strings.Builder) {
	b.WriteString("  ")
	b.WriteString(styleArgActive.Render("Arguments: "))
	b.WriteString(styleFilterText.Render(m.args.simpleArg))
	b.WriteString(styleFilterCur.Render("│"))
	b.WriteString("\n\n  ")
	b.WriteString(renderFooter("enter", "run (empty to skip)", "esc", "back"))
}

func (m Model) renderConfigArgs(b *strings.Builder) {
	labelW := m.args.labelWidth()

	for i, inp := range m.args.inputs {
		active := i == m.args.cursor

		label := fmt.Sprintf("%-*s", labelW, inp.def.Name)
		b.WriteString("  ")
		if active {
			b.WriteString(styleArgActive.Render(label))
		} else {
			b.WriteString(styleArgLabel.Render(label))
		}

		if inp.def.Required {
			b.WriteString(styleRequired.Render(" *"))
		} else {
			b.WriteString("  ")
		}

		b.WriteString("  ")
		if active {
			b.WriteString(styleFilterText.Render(inp.value))
			b.WriteString(styleFilterCur.Render("│"))
		} else {
			b.WriteString(inp.value)
		}

		if m.args.showValidation && inp.def.Required && inp.value == "" {
			b.WriteString("  ")
			b.WriteString(styleError.Render("required"))
		}

		b.WriteString("\n")

		if inp.def.Hint != "" {
			pad := labelW + 4
			b.WriteString(strings.Repeat(" ", 2+pad))
			b.WriteString(styleHint.Render(inp.def.Hint))
			b.WriteString("\n")
		}

		b.WriteString("\n")
	}

	b.WriteString("  ")
	b.WriteString(renderFooter("tab", "next", "enter", "run", "esc", "back"))
}

func (m Model) renderExecResult(b *strings.Builder) {
	r := m.list.lastRun

	status := "✓ Done"
	statusStyle := styleResultOK
	if r.failed() {
		status = "✗ Failed"
		statusStyle = styleResultFail
	}

	code := "—"
	if r.exitCode >= 0 {
		code = fmt.Sprintf("%d", r.exitCode)
	} else if r.ok() {
		code = "0"
	}

	var content strings.Builder

	// Column headers.
	content.WriteString(styleColHeader.Render(fmt.Sprintf("%-10s", "Status")))
	content.WriteString("  ")
	content.WriteString(styleColHeader.Render(fmt.Sprintf("%-6s", "Code")))
	content.WriteString("  ")
	content.WriteString(styleColHeader.Render(fmt.Sprintf("%-*s", m.layout.nameWidth, "Task")))
	content.WriteString("  ")
	content.WriteString(styleColHeader.Render("Command"))
	content.WriteString("\n")

	// Separator.
	sepLen := 10 + 2 + 6 + 2 + m.layout.nameWidth + 2 + len(r.command)
	if sepLen > m.layout.width-8 {
		sepLen = m.layout.width - 8
	}
	if sepLen < 20 {
		sepLen = 20
	}
	content.WriteString(styleColSep.Render(strings.Repeat("─", sepLen)))
	content.WriteString("\n")

	// Data row.
	content.WriteString(statusStyle.Render(fmt.Sprintf("%-10s", status)))
	content.WriteString("  ")
	if r.failed() {
		content.WriteString(styleResultFail.Render(fmt.Sprintf("%-6s", code)))
	} else {
		content.WriteString(styleResultOK.Render(fmt.Sprintf("%-6s", code)))
	}
	content.WriteString("  ")
	content.WriteString(styleResultValue.Render(fmt.Sprintf("%-*s", m.layout.nameWidth, r.taskName)))
	content.WriteString("  ")
	content.WriteString(styleResultValue.Render(r.command))

	// Error detail.
	if r.failed() && r.exitCode < 0 {
		content.WriteString("\n")
		content.WriteString(styleError.Render(r.err.Error()))
	}

	// Wrap in bordered box.
	box := styleResultBoxOK
	if r.failed() {
		box = styleResultBoxFail
	}
	b.WriteString(box.Render(content.String()))
	b.WriteString("\n")
}

func renderFooter(pairs ...string) string {
	var parts []string
	for i := 0; i+1 < len(pairs); i += 2 {
		key := styleFooterKey.Render(pairs[i])
		action := styleFooter.Render(" " + pairs[i+1])
		parts = append(parts, key+action)
	}
	return strings.Join(parts, styleFooterSep.Render("  ·  "))
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 1 {
		return "…"
	}
	return string(r[:max-1]) + "…"
}
