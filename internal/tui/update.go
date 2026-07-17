package tui

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"

	tea "charm.land/bubbletea/v2"
	"github.com/Megge06/TermiCam/internal/video"
)

// Define how the model updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if keyMsg.String() == "ctrl+c" || keyMsg.String() == "q" {
			if m.videoSession != nil {
				_ = m.videoSession.Close()
			}
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
		m.err = msg.err
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
				var device string
				for idx := range m.selected {
					device = m.choices[idx]
					break
				}

				if device != "" {
					// Query native camera resolution
					nativeW, nativeH, err := getDeviceResolution(device)
					if err != nil || nativeW == 0 || nativeH == 0 {
						nativeW, nativeH = 640, 480 // Fallback
					}

					// Calculate aspect ratio
					aspectRatio := float64(nativeW) / float64(nativeH)

					// Cap maximum width to 640px to keep CPU load low
					maxWidth := 640
					if nativeW < maxWidth {
						maxWidth = nativeW
					}

					// Scale height proportionally to maintain exact aspect ratio
					m.videoWidth = maxWidth
					m.videoHeight = int(float64(maxWidth) / aspectRatio)

					fps := 30

					session, err := video.NewSession(device, m.videoWidth, m.videoHeight, fps)
					if err != nil {
						m.err = err
						return m, nil
					}

					m.videoSession = session
					m.frameBuffer = make([]byte, m.videoWidth*m.videoHeight*3)
					m.backBuffer = make([]byte, m.videoWidth*m.videoHeight*3)
					m.screen = screenCamera

					return m, readFrameCmd(m.videoSession, m.backBuffer)
				}
			}
		}
	}
	return m, nil
}

// Update camera screen
func (m Model) updateCamera(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case frameMsg:
		m.frameBuffer, m.backBuffer = m.backBuffer, m.frameBuffer
		// Fetch the next frame into the back buffer
		return m, readFrameCmd(m.videoSession, m.backBuffer)

	case frameErrMsg:
		m.err = msg.err
		if m.videoSession != nil {
			_ = m.videoSession.Close()
			m.videoSession = nil
		}
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "backspace":
			if m.videoSession != nil {
				_ = m.videoSession.Close()
				m.videoSession = nil
			}
			m.screen = screenSelect
		case "h":
			m.hideUI = !m.hideUI
		}
	}
	return m, nil

}

// Read a frame from the video session and update the buffers
func readFrameCmd(s *video.Session, buf []byte) tea.Cmd {
	return func() tea.Msg {
		err := s.ReadFrame(buf)
		if err != nil {
			return frameErrMsg{err}
		}
		return frameMsg{}
	}
}

// Query device resolution to preserve aspect ratio
func getDeviceResolution(device string) (int, int, error) {
	out, err := exec.Command("v4l2-ctl", "-d", device, "--get-fmt-video").CombinedOutput()
	if err != nil {
		out, err = exec.Command("v4l2-ctl", "-d", device, "--get-fmt-video-out").CombinedOutput()
		if err != nil {
			return 640, 480, err
		}
	}

	// Look for pattern: Width/Height : 1280/720
	re := regexp.MustCompile(`Width/Height\s*:\s*(\d+)/(\d+)`)
	matches := re.FindStringSubmatch(string(out))

	if len(matches) < 3 {
		return 640, 480, fmt.Errorf("could not parse resolution from output")
	}

	w, _ := strconv.Atoi(matches[1])
	h, _ := strconv.Atoi(matches[2])

	return w, h, nil
}
