package tui

import (
	"charm.land/bubbles/v2/textinput"
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
	devices      []video.Device
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
	textInput    textinput.Model
	inputActive  bool
}

// Frame capture message types
type frameMsg struct{}
type frameErrMsg struct{ err error }

type devicesLoadedMsg []video.Device
type errMsg struct{ err error }

// Empty initial model, filled with data upon command completion
func InitialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "30"
	ti.SetValue("30")
	ti.CharLimit = 3
	ti.SetWidth(5)

	return Model{
		selected:    -1,
		loading:     true,
		hideUI:      false,
		color:       false,
		detailed:    false,
		fps:         30,
		textInput:   ti,
		inputActive: false,
	}
}

func (m Model) Init() tea.Cmd {
	return getDevicesCmd
}

func getDevicesCmd() tea.Msg {
	devices, err := video.ListDevices()
	if err != nil {
		return errMsg{err}
	}

	return devicesLoadedMsg(devices)
}
