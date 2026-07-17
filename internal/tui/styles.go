package tui

import "charm.land/lipgloss/v2"

// Define the Palette
var (
	magenta     = lipgloss.Color("#a4326b") // Main Magenta
	darkMagenta = lipgloss.Color("#4a0033") // Deeper background accent
	cyan        = lipgloss.Color("#00f5d4") // Cyan for selectors
	fg          = lipgloss.Color("#f8f8f2") // White
	gray        = lipgloss.Color("#6272a4") // Gray
	errorRed    = lipgloss.Color("#ff3333") // Error Red

	// Styles
	titleStyle    = lipgloss.NewStyle().Foreground(magenta).Bold(true)
	subtitleStyle = lipgloss.NewStyle().Foreground(gray).Italic(true)
	cursorStyle   = lipgloss.NewStyle().Foreground(cyan).Bold(true)
	checkedStyle  = lipgloss.NewStyle().Foreground(cyan).Bold(true)
	choiceStyle   = lipgloss.NewStyle().Foreground(fg)

	// Highlight the currently hovered list item
	currentChoiceStyle = lipgloss.NewStyle().Foreground(magenta).Bold(true).Underline(true)

	mutedStyle = lipgloss.NewStyle().Foreground(gray)
	errStyle   = lipgloss.NewStyle().Foreground(errorRed).Bold(true)

	logoStyle = lipgloss.NewStyle().Foreground(magenta).Bold(true)

	activeToggleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1a1a1a")).
				Background(cyan).
				Padding(0, 1).
				Bold(true)

	inactiveToggleStyle = lipgloss.NewStyle().
				Foreground(gray).
				Background(darkMagenta).
				Padding(0, 1)

	badgeStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Background(darkMagenta).
			Padding(0, 1).
			Bold(true)
)

// ASCII art wordmark, shown atop the settings and device-selection screens
const logo = "  ______                    _ ______\n /_  __/__  _________ ___  (_) ____/___ _____ ___\n  / / / _ \\/ ___/ __ `__ \\/ / /   / __ `/ __ `__ \\\n / / /  __/ /  / / / / / / / /___/ /_/ / / / / / /\n/_/  \\___/_/  /_/ /_/ /_/_/\\____/\\__,_/_/ /_/ /_/"
