package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type screenState int

const (
	screenSelect screenState = iota
	screenCamera
)

// Define the Palette
var (
	magenta     = lipgloss.Color("#a4326b") // Main Magenta
	darkMagenta = lipgloss.Color("#4a0033") // Deeper background accent
	cyan        = lipgloss.Color("#00f5d4") // Cyan for selectors
	fg          = lipgloss.Color("#f8f8f2") // Wh8ite
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
)

type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
	loading  bool
	err      error
	screen   screenState
}

type devicesLoadedMsg []string
type errMsg error

// Main entry point
func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

// Empty initial model, filled with data upon command completion
func initialModel() model {
	return model{
		selected: make(map[int]struct{}),
		loading:  true,
	}
}

func (m model) Init() tea.Cmd {
	return getDevicesCmd
}

// For linux, run the v4l2-ctl command to get a list of video devices
func getDevicesCmd() tea.Msg {
	out, err := exec.Command("v4l2-ctl", "--list-devices").Output()
	if err != nil {
		return errMsg(err)
	}

	// Parse output to get device names
	lines := strings.Split(string(out), "\n")
	var devices []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "/dev/video") {
			devices = append(devices, trimmed)
		}
	}

	return devicesLoadedMsg(devices)
}

// Define how the model updates
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if keyMsg.String() == "ctrl+c" || keyMsg.String() == "q" {
			return m, tea.Quit
		}
	}

	switch m.screen {
	case screenSelect:
		return m.updateSelect(msg)
	case screenCamera:
		return m.updateCamera(msg)
	}

	return m, nil
}

// Update select screen
func (m model) updateSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case devicesLoadedMsg:
		m.loading = false
		m.choices = msg
		return m, nil

	case errMsg:
		m.loading = false
		m.err = msg
		return m, nil

	// Handle keyboard inputs
	case tea.KeyPressMsg:
		if m.loading || m.err != nil {
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "space":
			if _, ok := m.selected[m.cursor]; ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		case "enter":
			if len(m.selected) > 0 {
				m.screen = screenCamera
			}
		}
	}
	return m, nil
}

// Update camera screen
func (m model) updateCamera(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "backspace":
			m.screen = screenSelect
		}
	}
	return m, nil
}

// Render the view
func (m model) View() tea.View {
	if m.err != nil {
		s := errStyle.Render(fmt.Sprintf("Error running v4l2-ctl: %v", m.err)) + "\n\n" + mutedStyle.Render("Press q to exit.")
		v := tea.NewView(s)
		v.AltScreen = true
		return v
	}

	switch m.screen {
	case screenSelect:
		return m.viewSelect()
	case screenCamera:
		return m.viewCamera()
	default:
		v := tea.NewView(errStyle.Render("Error: state machine desync."))
		v.AltScreen = true
		return v
	}
}

func (m model) viewSelect() tea.View {
	if m.loading {
		v := tea.NewView(subtitleStyle.Render("Scanning system for video devices..."))
		v.AltScreen = true
		return v
	}

	if len(m.choices) == 0 {
		s := errStyle.Render("No compatible video devices found.") + "\n\n" + mutedStyle.Render("Press q to exit.")
		v := tea.NewView(s)
		v.AltScreen = true
		return v
	}

	// Format Title Header
	s := titleStyle.Render("Select video devices to configure") + " " + subtitleStyle.Render("(Space to check, Enter to proceed):") + "\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = cursorStyle.Render(">")
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = checkedStyle.Render("x")
		}

		// Apply distinct styling to the row currently under the cursor
		renderedChoice := choiceStyle.Render(choice)
		if m.cursor == i {
			renderedChoice = currentChoiceStyle.Render(choice)
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, renderedChoice)
	}

	s += "\n" + mutedStyle.Render("Press q to quit.") + "\n"

	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

func (m model) viewCamera() tea.View {
	s := titleStyle.Render("--- Camera Screen ---") + "\n\n"

	// Print configured devices
	s += choiceStyle.Render("Active Feeds:") + "\n"
	for idx := range m.selected {
		s += fmt.Sprintf("  %s %s\n", cursorStyle.Render("•"), choiceStyle.Render(m.choices[idx]))
	}

	s += "\n" + mutedStyle.Render("Press ESC to go back, or q to quit.") + "\n"

	v := tea.NewView(s)
	v.AltScreen = true
	return v
}
