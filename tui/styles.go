package tui

import "github.com/charmbracelet/lipgloss"

// Brand palette — purple-to-teal gradient.
var (
	colorPurple    = lipgloss.Color("#7C3AED")
	colorIndigo    = lipgloss.Color("#6366F1")
	colorTeal      = lipgloss.Color("#14B8A6")
	colorTealLight = lipgloss.Color("#5EEAD4")
	colorLavender  = lipgloss.Color("#A78BFA")
	colorSuccess   = lipgloss.Color("#34D399")
	colorDanger    = lipgloss.Color("#F87171")
	colorSlate     = lipgloss.Color("#64748B")
	colorSlateLt   = lipgloss.Color("#94A3B8")
	colorSlateDk   = lipgloss.Color("#475569")
)

var (
	// Banner — ASCII art gradient lines (purple → indigo → teal).
	styleBannerL1 = lipgloss.NewStyle().Bold(true).Foreground(colorPurple)
	styleBannerL2 = lipgloss.NewStyle().Bold(true).Foreground(colorIndigo)
	styleBannerL3 = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#4DA6A8"))
	styleBannerL4 = lipgloss.NewStyle().Bold(true).Foreground(colorTeal)
	styleTagline  = lipgloss.NewStyle().Foreground(colorSlate).Italic(true)

	styleBannerBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSlateDk).
			Padding(0, 2)

	// Project info below banner.
	styleProjectInfo = lipgloss.NewStyle().Foreground(colorLavender).Bold(true)

	// Filter.
	styleFilterLabel = lipgloss.NewStyle().Foreground(colorSlate)
	styleFilterText  = lipgloss.NewStyle().Foreground(colorTeal)
	styleFilterCur   = lipgloss.NewStyle().Foreground(colorTeal)
	styleTaskCount   = lipgloss.NewStyle().Foreground(colorSlateDk)

	// Task list — column headers.
	styleColHeader = lipgloss.NewStyle().Foreground(colorSlateDk).Bold(true)
	styleColSep    = lipgloss.NewStyle().Foreground(colorSlateDk)

	// Task list — rows.
	styleCursor  = lipgloss.NewStyle().Foreground(colorTeal).Bold(true)
	styleSelName = lipgloss.NewStyle().Bold(true).Foreground(colorTeal)
	styleName    = lipgloss.NewStyle().Bold(true)
	styleCmd     = lipgloss.NewStyle().Foreground(colorSlate)
	styleDesc    = lipgloss.NewStyle().Foreground(colorSlateDk).Italic(true)
	styleDescSel = lipgloss.NewStyle().Foreground(colorTealLight).Italic(true)
	styleEmpty   = lipgloss.NewStyle().Foreground(colorSlate).Italic(true)

	// Footer.
	styleFooter    = lipgloss.NewStyle().Foreground(colorSlateDk)
	styleFooterKey = lipgloss.NewStyle().Foreground(colorSlateLt).Bold(true)
	styleFooterSep = lipgloss.NewStyle().Foreground(colorSlateDk)

	// Execution result table.
	styleResultBoxOK = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorSuccess).
				Padding(0, 2)

	styleResultBoxFail = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorDanger).
				Padding(0, 2)

	styleResultValue = lipgloss.NewStyle().Foreground(colorSlateLt)
	styleResultOK    = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	styleResultFail  = lipgloss.NewStyle().Foreground(colorDanger).Bold(true)

	// Args phase.
	styleSubHeader = lipgloss.NewStyle().Foreground(colorSlate).Italic(true)
	styleArgLabel  = lipgloss.NewStyle().Bold(true).Foreground(colorSlateLt)
	styleArgActive = lipgloss.NewStyle().Bold(true).Foreground(colorTeal)
	styleRequired  = lipgloss.NewStyle().Foreground(colorDanger).Bold(true)
	styleHint      = lipgloss.NewStyle().Foreground(colorSlateDk).Italic(true)
	styleError     = lipgloss.NewStyle().Foreground(colorDanger).Bold(true)
)
