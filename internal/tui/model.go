package tui

import (
	"charm.land/bubbles/v2/list"
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
	deviceList   list.Model
	mirror       bool
}

// Frame capture message types
type frameMsg struct{}
type frameErrMsg struct{ err error }

type devicesLoadedMsg []video.Device
type errMsg struct{ err error }

// List item for video devices
type deviceItem struct {
	device   video.Device
	selected bool
}

func (i deviceItem) Title() string {
	if i.selected {
		return "[x] " + i.device.String()
	}
	return "[ ] " + i.device.String()
}

func (i deviceItem) Description() string {
	return i.device.ID
}

func (i deviceItem) FilterValue() string {
	return i.device.String()
}

// placeholderItem serves as a visual indicator at the end of the list
type placeholderItem struct{}

func (p placeholderItem) Title() string {
	return "• (No other devices detected)"
}

func (p placeholderItem) Description() string {
	return "This is the end of the available video devices list."
}

func (p placeholderItem) FilterValue() string {
	return "No other devices detected"
}

func (m *Model) updateListItems() {
	items := make([]list.Item, 0, len(m.devices)+1)
	for i, d := range m.devices {
		items = append(items, deviceItem{
			device:   d,
			selected: m.selected == i,
		})
	}

	items = append(items, placeholderItem{})

	m.deviceList.SetItems(items)
}

// Empty initial model, filled with data upon command completion
func InitialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "30"
	ti.SetValue("30")
	ti.CharLimit = 3
	ti.SetWidth(5)

	// Configure and initialize the default list delegate using your existing tui color scheme
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(magenta).BorderLeftForeground(magenta)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(gray).BorderLeftForeground(magenta)

	l := list.New(nil, delegate, 0, 0)
	l.Title = "Select video devices to configure"
	l.SetShowHelp(false) // Custom helper layouts can still be rendered around it if desired

	return Model{
		selected:    -1,
		loading:     true,
		hideUI:      false,
		color:       false,
		detailed:    false,
		fps:         30,
		textInput:   ti,
		inputActive: false,
		deviceList:  l, // <-- NEW ASSIGNMENT
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
