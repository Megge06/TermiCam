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
		s := errStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n" + mutedStyle.Render("Press q to exit.")
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
	s := logoStyle.Render(logo) + "\n\n"
	s += titleStyle.Render("--- Settings ---") + "\n\n"

	fpsVal := fmt.Sprintf("%d", m.fps)
	if m.inputActive {
		fpsVal = m.textInput.View()
	}

	settings := []struct {
		label     string
		isToggle  bool
		toggleVal bool
		value     string
	}{
		{"Color Mode", true, m.color, ""},
		{"Detailed Palette", true, m.detailed, ""},
		{"Target FPS", false, false, fpsVal},
		{"Proceed to Device Selection", false, false, ""},
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

		if item.isToggle {
			s += fmt.Sprintf("%s %s: %s\n", cursor, labelStr, renderToggle(item.toggleVal))
		} else if item.value != "" {
			s += fmt.Sprintf("%s %s: %s\n", cursor, labelStr, checkedStyle.Render(item.value))
		} else {
			s += fmt.Sprintf("%s %s\n", cursor, labelStr)
		}
	}

	if m.inputActive {
		s += mutedStyle.Render("[Enter] Confirm FPS  [ESC] Cancel Editing")
	} else {
		s += mutedStyle.Render("[Space] Edit [Enter] Proceed  [q] Quit")
	}

	v := tea.NewView(s)
	v.AltScreen = true
	return v
}

func (m Model) viewSelect() tea.View {
	if m.loading {
		s := logoStyle.Render(logo) + "\n\n" + subtitleStyle.Render("Scanning system for video devices...")
		v := tea.NewView(s)
		return v
	}

	if len(m.devices) == 0 {
		s := logoStyle.Render(logo) + "\n\n"
		s += errStyle.Render("No compatible video devices found.") + "\n\n" + mutedStyle.Render("Press q to exit.")
		v := tea.NewView(s)
		v.AltScreen = true
		return v
	}

	// Render the Logo Header followed by the interactive search list
	s := logoStyle.Render(logo) + "\n\n"
	s += m.deviceList.View()

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
		reservedHeight = 9
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

	// Render Header and Meta Badges
	if !m.hideUI {
		headerText := titleStyle.Render("--- Camera Screen ---")

		// Visual status badges indicating session information
		fpsBadge := badgeStyle.Render(fmt.Sprintf(" %d FPS ", m.fps))

		var colorMode string
		if m.color {
			colorMode = " COLOR "
		} else {
			colorMode = " B&W "
		}
		colorBadge := badgeStyle.Render(colorMode)

		var detailMode string
		if m.detailed {
			detailMode = " DETAILED "
		} else {
			detailMode = " SIMPLE "
		}
		detailBadge := badgeStyle.Render(detailMode)

		badges := lipgloss.JoinHorizontal(lipgloss.Center, fpsBadge, " ", colorBadge, " ", detailBadge)

		s = lipgloss.PlaceHorizontal(m.termWidth-2, lipgloss.Center, headerText) + "\n"
		s += lipgloss.PlaceHorizontal(m.termWidth-2, lipgloss.Center, badges) + "\n\n"
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

// renderToggle draws the active/inactive visual toggle pill representation
func renderToggle(enabled bool) string {
	if enabled {
		return activeToggleStyle.Render(" ON ") + inactiveToggleStyle.Render(" OFF ")
	}
	activeMuted := activeToggleStyle.Background(gray).Foreground(fg)
	return inactiveToggleStyle.Render(" ON ") + activeMuted.Render(" OFF ")
}
