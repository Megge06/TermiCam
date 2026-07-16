package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

// Render the view
func (m Model) View() tea.View {
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

func (m Model) viewSelect() tea.View {
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

func (m Model) viewCamera() tea.View {
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
