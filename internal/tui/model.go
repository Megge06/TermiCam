package tui

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"os"

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
	recording    bool
	recordFile   *os.File
	recordGzip   *gzip.Writer
	playbackMode bool
	playbackFile *os.File
	playbackGzip *gzip.Reader
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
	saved := LoadSettings()

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
	l.SetShowHelp(false)
	m := Model{
		selected:    -1,
		loading:     true,
		hideUI:      false,
		color:       saved.Color,
		detailed:    saved.Detailed,
		mirror:      saved.Mirror,
		fps:         saved.FPS,
		textInput:   ti,
		inputActive: false,
		deviceList:  l,
		recording:   false,
	}

	// Detect if file is being given as argument
	if len(os.Args) > 1 {
		filePath := os.Args[1]
		if err := m.loadRecording(filePath); err == nil {
			m.screen = screenCamera
			m.loading = false
		} else {
			m.err = fmt.Errorf("failed to load recording %q: %w", filePath, err)
		}
	}

	return m
}

func (m Model) Init() tea.Cmd {
	if m.playbackMode {
		return playbackTickCmd(m.fps)
	}
	return getDevicesCmd
}

func getDevicesCmd() tea.Msg {
	devices, err := video.ListDevices()
	if err != nil {
		return errMsg{err}
	}

	return devicesLoadedMsg(devices)
}

// Put atop recorded file to identify metadata
type fileHeader struct {
	Magic    [4]byte
	Width    int32
	Height   int32
	FPS      int32
	Color    byte
	Detailed byte
	Mirror   byte
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func byteToBool(b byte) bool {
	return b != 0
}

// Create writer, etc. to start recording
func (m *Model) startRecording(path string) error {
	// #nosec G304 -- Safe. Path is generated internally or supplied safely from command arguments
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	m.recordFile = file

	gw := gzip.NewWriter(file)
	m.recordGzip = gw

	// Save metadata atop of the file
	header := fileHeader{
		Magic:    [4]byte{'T', 'C', 'A', 'M'},
		Width:    int32(m.videoWidth),  // #nosec G115 -- Safe. Dimensions are small terminal resolutions
		Height:   int32(m.videoHeight), // #nosec G115 -- Safe. Dimensions are small terminal resolutions
		FPS:      int32(m.fps),         // #nosec G115 -- Safe. Target FPS is validated and low
		Color:    boolToByte(m.color),
		Detailed: boolToByte(m.detailed),
		Mirror:   boolToByte(m.mirror),
	}

	if err := binary.Write(gw, binary.LittleEndian, header); err != nil {
		_ = gw.Close()
		_ = file.Close()
		return err
	}

	m.recording = true
	return nil
}

// Because the data is streamed stopping is simple
func (m *Model) stopRecording() {
	m.recording = false
	if m.recordGzip != nil {
		_ = m.recordGzip.Close()
		m.recordGzip = nil
	}
	if m.recordFile != nil {
		_ = m.recordFile.Close()
		m.recordFile = nil
	}
}

// Start playback with metadata information from given file
func (m *Model) loadRecording(path string) error {
	// #nosec G304 G703 -- Safe. Path is passed safely from CLI arguments
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	m.playbackFile = file

	gr, err := gzip.NewReader(file)
	if err != nil {
		_ = file.Close()
		return err
	}
	m.playbackGzip = gr

	var header fileHeader
	if err := binary.Read(gr, binary.LittleEndian, &header); err != nil {
		_ = gr.Close()
		_ = file.Close()
		return err
	}

	if string(header.Magic[:]) != "TCAM" {
		_ = gr.Close()
		_ = file.Close()
		return fmt.Errorf("invalid recording file format")
	}

	// Change model information to fit playback
	m.videoWidth = int(header.Width)
	m.videoHeight = int(header.Height)
	m.fps = int(header.FPS)
	m.color = byteToBool(header.Color)
	m.detailed = byteToBool(header.Detailed)
	m.mirror = byteToBool(header.Mirror)
	m.playbackMode = true

	// Buffer holds exactly one frame
	m.frameBuffer = make([]byte, m.videoWidth*m.videoHeight*3)

	if err := m.readNextPlaybackFrame(); err != nil {
		_ = gr.Close()
		_ = file.Close()
		return fmt.Errorf("failed to read initial frame: %w", err)
	}

	return nil
}

func (m *Model) readNextPlaybackFrame() error {
	if m.playbackGzip == nil {
		return io.EOF
	}

	_, err := io.ReadFull(m.playbackGzip, m.frameBuffer)
	if err != nil {
		// End of file is reached, rewind to beginning
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return m.rewindPlayback()
		}
		return err
	}
	return nil
}

// Rewind to beginning of file by opening a new gzip reader
func (m *Model) rewindPlayback() error {
	if m.playbackFile == nil {
		return io.EOF
	}

	_ = m.playbackGzip.Close()
	_ = m.playbackFile.Close()

	// #nosec G304, G703 -- Safe. Path was validated on initial load.
	file, err := os.Open(m.playbackFile.Name())
	if err != nil {
		return err
	}
	m.playbackFile = file

	gr, err := gzip.NewReader(file)
	if err != nil {
		_ = file.Close()
		return err
	}
	m.playbackGzip = gr

	var header fileHeader
	if err := binary.Read(gr, binary.LittleEndian, &header); err != nil {
		_ = gr.Close()
		_ = file.Close()
		return err
	}

	_, err = io.ReadFull(m.playbackGzip, m.frameBuffer)
	return err
}

// Simply stops placback by closing the reader and file stream
func (m *Model) closePlayback() {
	m.playbackMode = false
	if m.playbackGzip != nil {
		_ = m.playbackGzip.Close()
		m.playbackGzip = nil
	}
	if m.playbackFile != nil {
		_ = m.playbackFile.Close()
		m.playbackFile = nil
	}
}
