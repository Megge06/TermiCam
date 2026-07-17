package tui

import (
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Megge06/TermiCam/internal/video"
)

type screenState int

const (
	screenSelect screenState = iota
	screenCamera
)

type Model struct {
	choices      []string
	cursor       int
	selected     map[int]struct{}
	loading      bool
	hideUI       bool
	err          error
	screen       screenState
	termWidth    int
	termHeight   int
	videoSession *video.Session
	frameBuffer  []byte
}

type devicesLoadedMsg []string
type errMsg error

// Empty initial model, filled with data upon command completion
func InitialModel() Model {
	return Model{
		selected: make(map[int]struct{}),
		loading:  true,
		hideUI:   false,
	}
}

func (m Model) Init() tea.Cmd {
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
