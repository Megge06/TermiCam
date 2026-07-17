package tui

import (
	tea "charm.land/bubbletea/v2"
)

// Define how the model updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if keyMsg.String() == "ctrl+c" || keyMsg.String() == "q" {
			return m, tea.Quit
		}
	}

	// Handle window size changes
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
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
func (m Model) updateSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (m Model) updateCamera(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "backspace":
			m.screen = screenSelect
		case "h":
			m.hideUI = !m.hideUI
		}
	}
	return m, nil
}
