package tui

import (
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Megge06/TermiCam/internal/video"
)

type screenState int

const (
	screenSettings screenState = iota
	screenSelect
	screenCamera
)

type Model struct {
	choices      []string
	cursor       int
	selected     int
	loading      bool
	hideUI       bool
	err          error
	screen       screenState
	termWidth    int
	termHeight   int
	videoSession *video.Session
	frameBuffer  []byte
	backBuffer   []byte
	videoWidth   int
	videoHeight  int
	color        bool
	detailed     bool
	fps          int
}

// Frame capture message types
type frameMsg struct{}
type frameErrMsg struct{ err error }

type devicesLoadedMsg []string
type errMsg struct{ err error }

// Empty initial model, filled with data upon command completion
func InitialModel() Model {
	return Model{
		selected: -1,
		loading:  true,
		hideUI:   false,
		color:    false,
		detailed: false,
		fps:      30,
	}
}

func (m Model) Init() tea.Cmd {
	return getDevicesCmd
}

// For linux, run the v4l2-ctl command to get a list of video devices
func getDevicesCmd() tea.Msg {
	out, err := exec.Command("v4l2-ctl", "--list-devices").Output()
	if err != nil {
		return errMsg{err}
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
