package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Megge06/TermiCam/internal/ascii"
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
	case screenSettings:
		return m.viewSettings()
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

func (m Model) viewSettings() tea.View {
	s := titleStyle.Render("--- Settings ---") + "\n\n"

	settings := []struct {
		label string
		value string
	}{
		{"Color Mode", fmt.Sprintf("%t", m.color)},
		{"Detailed Palette", fmt.Sprintf("%t", m.detailed)},
		{"Target FPS", fmt.Sprintf("%d", m.fps)},
		{"Proceed to Device Selection", ""},
	}

	for i, item := range settings {
		cursor := " "
		if m.cursor == i {
			cursor = cursorStyle.Render(">")
		}

		labelStr := item.label
		if m.cursor == i {
			labelStr = currentChoiceStyle.Render(item.label)
		} else {
			labelStr = choiceStyle.Render(item.label)
		}

		if item.value != "" {
			s += fmt.Sprintf("%s %s: %s\n", cursor, labelStr, checkedStyle.Render(item.value))
		} else {
			s += fmt.Sprintf("%s %s\n", cursor, labelStr)
		}
	}

	s += "\n" + mutedStyle.Render("[Space] Toggle/Cycle  [Enter] Proceed  [q] Quit")

	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

func (m Model) viewSelect() tea.View {
	if m.loading {
		v := tea.NewView(subtitleStyle.Render("Scanning system for video devices..."))
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
		if m.selected == i {
			checked = checkedStyle.Render("x")
		}

		renderedChoice := choiceStyle.Render(choice)
		if m.cursor == i {
			renderedChoice = currentChoiceStyle.Render(choice)
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, renderedChoice)
	}
	s += "\n" + mutedStyle.Render("[Space] Toggle/Cycle  [Enter] Proceed  [ESC] Go Back [q] Quit") + "\n"

	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

func (m Model) viewCamera() tea.View {
	var s string

	if m.videoSession == nil || len(m.frameBuffer) == 0 {
		s += errStyle.Render("Camera session is not initialized.")
		v := tea.NewView(s)
		v.AltScreen = true
		return v
	}

	// Get the dimensions of the image
	imgWidth := m.videoWidth
	imgHeight := m.videoHeight

	targetWidth := m.termWidth - 4
	if targetWidth <= 0 {
		targetWidth = 80
	}

	// Calculate the maximum height for the ASCII art based on the terminal height
	var reservedHeight int
	if !m.hideUI {
		reservedHeight = 7
	} else {
		reservedHeight = 0
	}
	maxHeight := m.termHeight - reservedHeight
	if maxHeight <= 0 {
		maxHeight = 10
	}

	if imgWidth > 0 && imgHeight > 0 {
		// Calculate the maximum width that will keep the height within the maxHeight limit
		aspectRatio := float64(imgWidth) / float64(imgHeight)
		vFitWidth := int(float64(maxHeight) * aspectRatio / 0.45)

		// If fitting vertically requires a smaller width, scale down targetWidth
		if vFitWidth < targetWidth {
			targetWidth = vFitWidth
		}

		if targetWidth > imgWidth {
			targetWidth = imgWidth
		}
		if targetWidth <= 0 {
			targetWidth = 1
		}
	}
	// Pass the loaded frame into the raw RGB24 converter
	asciiArt, err := ascii.ConvertRGB24ToASCII(m.frameBuffer, imgWidth, imgHeight, targetWidth, m.color, m.detailed)
	if err != nil {
		s += errStyle.Render(fmt.Sprintf("Error converting image to ASCII: %v", err))
	}

	// Render Header
	if !m.hideUI {
		s = lipgloss.PlaceHorizontal(m.termWidth-2, lipgloss.Center, titleStyle.Render("--- Camera Screen ---")) + "\n\n"
	}

	// Centered ASCII Art
	centeredAscii := lipgloss.PlaceHorizontal(m.termWidth-2, lipgloss.Center, asciiArt)
	s += centeredAscii

	// Render Footer
	if !m.hideUI {
		s += "\n\n" + lipgloss.PlaceHorizontal(m.termWidth-2, lipgloss.Center, mutedStyle.Render("[h] Hide UI"))
		s += "\n" + lipgloss.PlaceHorizontal(m.termWidth-2, lipgloss.Center, mutedStyle.Render("[Space] Toggle/Cycle  [Enter] Proceed  [ESC] Go Back [q] Quit"))
	}

	v := tea.NewView(s)
	v.AltScreen = true
	return v
}
