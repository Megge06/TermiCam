package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/Megge06/TermiCam/internal/video"
)

type playbackTickMsg struct{}

// Define how the model updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		// Do not quit on "q" if an input field is active or the list is filtering
		isTyping := m.inputActive || (m.screen == screenSelect && m.deviceList.FilterState() == list.Filtering)

		if keyMsg.String() == "ctrl+c" || (keyMsg.String() == "q" && !isTyping) {
			if m.videoSession != nil {
				_ = m.videoSession.Close()
			}
			m.stopRecording()
			m.closePlayback()
			return m, tea.Quit
		}
	}

	// Handle window size changes
	switch msg := msg.(type) {
	// In case of playing back a recording
	case playbackTickMsg:
		if m.playbackMode {
			if err := m.readNextPlaybackFrame(); err != nil {
				m.err = err
				m.closePlayback()
				return m, nil
			}
			return m, playbackTickCmd(m.fps)
		}

	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height

		// Reserve some height for the logo block
		h := msg.Height - 8
		if h < 0 {
			h = 0
		}
		m.deviceList.SetSize(msg.Width, h) //

	case devicesLoadedMsg:
		m.loading = false
		m.devices = []video.Device(msg)
		m.choices = make([]string, len(m.devices))
		for i, device := range m.devices {
			m.choices[i] = device.String()
		}
		m.updateListItems()
		return m, nil
	case errMsg:
		m.loading = false
		m.err = msg.err
		return m, nil
	}

	switch m.screen {
	case screenSelect:
		return m.updateSelect(msg)
	case screenCamera:
		return m.updateCamera(msg)
	case screenSettings:
		return m.updateSettings(msg)
	}

	return m, nil
}

func (m Model) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.inputActive {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)

		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "enter":
				// Confirm the change
				val := strings.TrimSpace(m.textInput.Value())
				if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
					m.fps = parsed
					m.persistCurrentSettings()
				} else {
					// Fallback if the entered value is invalid
					m.textInput.SetValue(strconv.Itoa(m.fps))
				}
				m.textInput.Blur()
				m.inputActive = false
			case "esc":
				// Revert changes and exit edit mode
				m.textInput.SetValue(strconv.Itoa(m.fps))
				m.textInput.Blur()
				m.inputActive = false
			}
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	// Handle keyboard inputs
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			// There are 5 items (0: Color, 1: Detailed, 2: Mirror, 3: FPS, 4: Proceed)
			if m.cursor < 4 {
				m.cursor++
			}
		case "space", "left", "right":
			switch m.cursor {
			case 0:
				m.color = !m.color
				m.persistCurrentSettings()
			case 1:
				m.detailed = !m.detailed
				m.persistCurrentSettings()
			case 2:
				m.mirror = !m.mirror
				m.persistCurrentSettings()
			case 3:
				// Activate input on Space
				m.inputActive = true
				m.textInput.Focus()
				m.textInput.SetValue(strconv.Itoa(m.fps))
				return m, textinput.Blink
			case 4:
				m.persistCurrentSettings()
				m.screen = screenSelect
				m.cursor = 0
				return m, nil
			}

		case "enter":
			m.persistCurrentSettings()
			m.screen = screenSelect
			m.cursor = 0
			return m, nil

		}
	}
	return m, nil
}

// Update select screen
func (m Model) updateSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	wasFiltering := m.deviceList.FilterState() == list.Filtering

	var cmd tea.Cmd
	m.deviceList, cmd = m.deviceList.Update(msg)

	switch msg := msg.(type) {

	// Handle keyboard inputs
	case tea.KeyPressMsg:
		if m.loading || m.err != nil {
			return m, nil
		}

		// Prevent custom actions if the user is typing query filters inside the list searchbox
		if wasFiltering {
			return m, cmd
		}

		switch msg.String() {
		case "space":
			selectedItem := m.deviceList.SelectedItem()
			if selectedItem != nil {
				if devItem, ok := selectedItem.(deviceItem); ok {
					targetIdx := -1
					for i, d := range m.devices {
						if d.ID == devItem.device.ID {
							targetIdx = i
							break
						}
					}
					if targetIdx != -1 {
						if m.selected == targetIdx {
							m.selected = -1
						} else {
							m.selected = targetIdx
						}
						m.updateListItems()
					}
				}
			}
		case "esc", "backspace":
			m.screen = screenSettings
			m.cursor = 0
			return m, nil
		case "enter":
			idx := m.selected

			// If no device was explicitly checked with Space, fall back to the currently focused device item
			if idx == -1 {
				selectedItem := m.deviceList.SelectedItem()
				if selectedItem != nil {
					if devItem, ok := selectedItem.(deviceItem); ok {
						for i, d := range m.devices {
							if d.ID == devItem.device.ID {
								idx = i
								break
							}
						}
					}
				}
			}

			if idx >= 0 && idx < len(m.devices) {
				device := m.devices[idx]

				if device.ID != "" {
					// Query native camera resolution
					nativeW, nativeH, err := video.GetDeviceResolution(device)
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

					session, err := video.NewSession(device, nativeW, nativeH, m.videoWidth, m.videoHeight, m.fps)
					if err != nil {
						m.err = err
						return m, nil
					}

					m.videoSession = session
					m.frameBuffer = make([]byte, m.videoWidth*m.videoHeight*3)
					m.backBuffer = make([]byte, m.videoWidth*m.videoHeight*3)
					m.screen = screenCamera

					m.cursor = 0
					return m, readFrameCmd(m.videoSession, m.backBuffer)
				}
			}
		}
	}
	return m, cmd
}

// Update camera screen
func (m Model) updateCamera(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case frameMsg:
		m.frameBuffer, m.backBuffer = m.backBuffer, m.frameBuffer
		// Reads recorded frame if a recording is played
		if m.recording && m.recordGzip != nil {
			_, err := m.recordGzip.Write(m.frameBuffer)
			if err != nil {
				m.err = fmt.Errorf("failed to write recorded frame: %w", err)
				m.stopRecording()
			}
		}
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
			if m.playbackMode {
				m.closePlayback()
				return m, tea.Quit
			}
			if m.videoSession != nil {
				_ = m.videoSession.Close()
				m.videoSession = nil
			}
			m.stopRecording()
			m.screen = screenSelect
		case "h":
			m.hideUI = !m.hideUI
		case "r":
			if !m.playbackMode {
				if m.recording {
					m.stopRecording()
				} else {
					filename := fmt.Sprintf("rec_%s.tcam", time.Now().Format("2006-01-02_15-04-05"))
					if err := m.startRecording(filename); err != nil {
						m.err = err
					}
				}
			}
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

// Paces playback to privded fps
func playbackTickCmd(fps int) tea.Cmd {
	if fps <= 0 {
		fps = 30
	}
	return tea.Tick(time.Second/time.Duration(fps), func(t time.Time) tea.Msg {
		return playbackTickMsg{}
	})
}

// persistCurrentSettings saves the current settings to disk
func (m Model) persistCurrentSettings() {
	settings := PersistedSettings{
		Color:    m.color,
		Detailed: m.detailed,
		Mirror:   m.mirror,
		FPS:      m.fps,
	}
	_ = SaveSettings(settings)
}
