package ui

import "github.com/charmbracelet/lipgloss"

var (
	ColorGreen  = lipgloss.Color("#00CC66")
	ColorRed    = lipgloss.Color("#FF4455")
	ColorYellow = lipgloss.Color("#FFCC00")
	ColorGray   = lipgloss.Color("#6C7086")
	ColorPurple = lipgloss.Color("#CBA6F7")
	ColorBlue   = lipgloss.Color("#89B4FA")
	ColorWhite  = lipgloss.Color("#CDD6F4")

	StylePassed = lipgloss.NewStyle().Foreground(ColorGreen)
	StyleFailed = lipgloss.NewStyle().Foreground(ColorRed).Bold(true)
	StyleSkipped = lipgloss.NewStyle().Foreground(ColorGray).Italic(true)
	StyleRunning = lipgloss.NewStyle().Foreground(ColorYellow)
	StyleBold    = lipgloss.NewStyle().Bold(true)
	StyleDim     = lipgloss.NewStyle().Foreground(ColorGray)
	StyleJobName = lipgloss.NewStyle().Bold(true).Foreground(ColorPurple)
	StyleHeader  = lipgloss.NewStyle().Bold(true).Foreground(ColorWhite)

	StyleBreakpoint = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorYellow).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorYellow).
			Padding(0, 1)

	StyleSummaryTable = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorPurple).
				Padding(0, 1)

	StyleErrorBox = lipgloss.NewStyle().
			Foreground(ColorRed).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorRed).
			Padding(0, 1)
)

// NoColor disables all color styling.
func NoColor() {
	plain := lipgloss.NewStyle()
	StylePassed = plain
	StyleFailed = plain.Bold(true)
	StyleSkipped = plain
	StyleRunning = plain
	StyleBold = plain.Bold(true)
	StyleDim = plain
	StyleJobName = plain.Bold(true)
	StyleHeader = plain.Bold(true)
	StyleBreakpoint = plain.Bold(true)
	StyleSummaryTable = plain
	StyleErrorBox = plain
}
