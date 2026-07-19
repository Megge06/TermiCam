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
		{"Mirror Image", true, m.mirror, ""},
		{"Target FPS", false, false, fpsVal},
		{"Proceed to Device Selection", false, false, ""},
	}

	for i, item := range settings {
		cursor := " "
		if m.cursor == i {
			cursor = cursorStyle.Render(">")
		}

		var labelStr string
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
	if (m.videoSession == nil && !m.playbackMode) || len(m.frameBuffer) == 0 {
		v := tea.NewView(errStyle.Render("Camera session is not initialized."))
		v.AltScreen = true
		return v
	}

	// Get the dimensions of the image
	imgWidth := m.videoWidth
	imgHeight := m.videoHeight

	// Allocate screen layout width
	var reservedWidth int
	if !m.hideUI {
		reservedWidth = 40
	} else {
		reservedWidth = 4
	}

	targetWidth := m.termWidth - reservedWidth
	if targetWidth <= 0 {
		targetWidth = 80
	}

	// Calculate the maximum height for the ASCII art based on the terminal height
	var reservedHeight int
	if !m.hideUI {
		reservedHeight = 4
	} else {
		reservedHeight = 2
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

	// Calculate the exact height of the ASCII artwork using the same formula as ascii.ConvertRGB24ToASCII
	targetHeight := int(float64(targetWidth*imgHeight/imgWidth) * 0.45)
	if targetHeight <= 0 {
		targetHeight = 1
	}

	// Pass the loaded frame into the raw RGB24 converter
	asciiArt, err := ascii.ConvertRGB24ToASCII(m.frameBuffer, imgWidth, imgHeight, targetWidth, m.color, m.detailed, m.mirror)
	if err != nil {
		asciiArt = errStyle.Render(fmt.Sprintf("Error converting image to ASCII: %v", err))
	}

	// If UI is hidden, render raw centered ASCII stream directly
	if m.hideUI {
		centeredRaw := lipgloss.Place(m.termWidth, m.termHeight, lipgloss.Center, lipgloss.Center, asciiArt)
		v := tea.NewView(centeredRaw)
		v.AltScreen = true
		return v
	}

	// Render dynamic sidebar content using active settings
	var colorMode string
	if m.color {
		colorMode = "COLOR"
	} else {
		colorMode = "B&W"
	}

	var detailMode string
	if m.detailed {
		detailMode = "DETAILED"
	} else {
		detailMode = "SIMPLE"
	}

	// Change recording/playback status
	var statusStr string
	if m.playbackMode {
		statusStr = badgeStyle.Render(" PLAYBACK ")
	} else if m.recording {
		statusStr = errStyle.Render(" RECORDING ")
	} else {
		statusStr = badgeStyle.Render(" LIVE ")
	}

	var recKeymap string
	if !m.playbackMode {
		recKeymap = mutedStyle.Render("[r]      Toggle Record\n")
	}

	hudContent := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("MONITOR HUD"),
		"",
		fmt.Sprintf("Status:     %s", statusStr),
		fmt.Sprintf("Resolution: %dx%d", m.videoWidth, m.videoHeight),
		fmt.Sprintf("Target FPS: %s", badgeStyle.Render(fmt.Sprintf(" %d ", m.fps))),
		fmt.Sprintf("Color Mode: %s", badgeStyle.Render(colorMode)),
		fmt.Sprintf("Palette:    %s", badgeStyle.Render(detailMode)),
		"",
		mutedStyle.Render("Terminal Scale:"),
		mutedStyle.Render(fmt.Sprintf("%dx%d", targetWidth, targetHeight)),
		"",
		mutedStyle.Render("Terminal Window:"),
		mutedStyle.Render(fmt.Sprintf("%dx%d", m.termWidth, m.termHeight)),
		titleStyle.Render("--- KEYMAP ---"),
		mutedStyle.Render("[h]      Toggle HUD"),
		recKeymap+
			mutedStyle.Render("[ESC]    Go Back"),
		mutedStyle.Render("[q]      Exit Program"),
	)

	// Combine the viewfinder monitor box and the HUD side-panel horizontally
	var monitor string
	if m.recording {
		monitor = recordingStyle.Render(asciiArt)
	} else {
		monitor = monitorStyle.Render(asciiArt)
	}
	sidebar := hudStyle.Render(hudContent)
	compositeView := lipgloss.JoinHorizontal(lipgloss.Top, monitor, "   ", sidebar)

	// Center-align the entire structured panel in the active terminal window
	centeredComposite := lipgloss.Place(m.termWidth, m.termHeight, lipgloss.Center, lipgloss.Center, compositeView)

	v := tea.NewView(centeredComposite)
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
