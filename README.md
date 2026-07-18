# TermiCam

A real-time ASCII camera for your terminal.

TermiCam is a Go TUI application that reads camera frames through FFmpeg and renders them as ASCII art in your terminal. It supports Linux, macOS, and Windows camera capture through platform-specific FFmpeg backends.

## Requirements

- **Go 1.26.5+** to build the app.
- **FFmpeg** for camera capture. `ffmpeg` must be available in your `PATH`.
- A terminal with enough space to render the camera view.

Platform-specific requirements:

- **Linux:** install `v4l2-ctl` from `v4l-utils`.
- **macOS:** grant camera permission to the terminal app you use to run TermiCam.
- **Windows:** use an FFmpeg build with DirectShow support. Most common Windows FFmpeg builds include this.

## Local Setup

Clone the repository:

```sh
git clone https://github.com/Megge06/TermiCam.git
cd TermiCam
```

Install Go dependencies:

```sh
go mod download
```

Run the app from source:

```sh
go run ./cmd/termicam
```

Build a local binary:

```sh
go build -o termicam ./cmd/termicam
```

Run the built binary on Linux or macOS:

```sh
./termicam
```

On Windows, build and run from PowerShell:

```powershell
go build -o termicam.exe ./cmd/termicam
.\termicam.exe
```

## Installing Platform Dependencies

### macOS

Install FFmpeg:

```sh
brew install ffmpeg
```

When you first run TermiCam, macOS may ask for camera access. Allow access for the terminal application you are using, such as Terminal, iTerm2, or VS Code.

If camera access was previously denied, update it in **System Settings > Privacy & Security > Camera**.

### Linux

Install FFmpeg and `v4l-utils`.

Debian or Ubuntu:

```sh
sudo apt update
sudo apt install ffmpeg v4l-utils
```

Fedora:

```sh
sudo dnf install ffmpeg v4l-utils
```

Arch Linux:

```sh
sudo pacman -S ffmpeg v4l-utils
```

Confirm that cameras are visible:

```sh
v4l2-ctl --list-devices
```

### Windows

Install Go and FFmpeg with `winget` from PowerShell:

```powershell
winget install --id GoLang.Go -e
winget install --id Gyan.FFmpeg -e
```

Restart PowerShell after installation so the updated `PATH` is loaded.

Confirm FFmpeg is available:

```powershell
ffmpeg -version
```

Confirm DirectShow can see video devices:

```powershell
ffmpeg -hide_banner -f dshow -list_devices true -i dummy
```

TermiCam uses FFmpeg's DirectShow input on Windows. If no camera appears, check Windows camera privacy settings and make sure desktop apps are allowed to access the camera.

## Usage

Start the app:

```sh
go run ./cmd/termicam
```

The app starts on a settings screen. Configure the display options, proceed to device selection, choose a camera, and press Enter to start the ASCII camera view.

Keybindings:

- `up` / `down` or `k` / `j`: move through settings.
- `space`: toggle settings or select a camera.
- `enter`: continue or start the selected camera.
- `esc` / `backspace`: go back from camera view or device selection.
- `h`: toggle the camera HUD.
- `q` or `ctrl+c`: quit.

## Development Commands

Run package checks:

```sh
go test ./...
```

Format Go files:

```sh
gofmt -w .
```

Build for the current platform:

```sh
go build -o termicam ./cmd/termicam
```

## Troubleshooting

### `ffmpeg` Not Found

Make sure FFmpeg is installed and available in your `PATH`:

```sh
ffmpeg -version
```

### No Camera Devices Found

Check that your operating system can see the camera first.

Linux:

```sh
v4l2-ctl --list-devices
```

macOS:

```sh
ffmpeg -hide_banner -f avfoundation -list_devices true -i ""
```

Windows:

```powershell
ffmpeg -hide_banner -f dshow -list_devices true -i dummy
```

### Permission Errors

- **macOS:** allow camera access for your terminal in System Settings.
- **Windows:** allow camera access for desktop apps in Camera privacy settings.
- **Linux:** make sure your user has permission to access `/dev/video*` devices.

### Slow Rendering

Try lowering the target FPS in the settings screen or use a smaller terminal window. TermiCam caps capture scaling to reduce CPU load, but color output and detailed palettes can still be heavier in large terminals.

## Project Structure

- `cmd/termicam/main.go`: application entry point.
- `internal/tui/`: Bubble Tea model, update loop, views, and styling.
- `internal/video/`: platform-specific camera discovery and FFmpeg capture sessions.
- `internal/ascii/`: RGB24 frame to ASCII conversion.

## Built With

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - A powerful little TUI framework, used to power TermiCam's interface.
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** - Style definitions for nice terminal layouts, used for styling and rendering in the TUI.
- **[FFmpeg](https://github.com/ffmpeg/ffmpeg)** - A complete, cross-platform solution to record, convert and stream audio and video, used for getting a stream of the camera video.

## Acknowledgments

- **Algorithm Inspiration:** The core image-to-ASCII conversion logic, character mapping, and 8-bit color scaling in this project were heavily inspired by the fantastic [ascii-image-converter](https://github.com/TheZoraiz/ascii-image-converter) by [Zoraiz Hassan](https://github.com/TheZoraiz).
