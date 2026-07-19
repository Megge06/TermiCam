<img width="523" height="132" alt="TermiCam Ascii Art" src="https://github.com/user-attachments/assets/f97967de-3e74-4d5d-aea6-7d36c93c4d7c" />

# TermiCam

<p align="left">
  <a href="https://github.com/Megge06/TermiCam/releases">
    <img src="https://img.shields.io/github/v/release/Megge06/TermiCam?color=a4326b&label=release" alt="Latest Release">
  </a>
  <a href="https://github.com/Megge06/TermiCam/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/license-GPLv3-a4326b" alt="License GPLv3">
  </a>
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white" alt="Go Version">
  <img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-00f5d4?logo=gnubash&logoColor=black" alt="Platforms Supported">
  <img src="https://img.shields.io/badge/dependency-FFmpeg-FF007F?logo=ffmpeg&logoColor=white" alt="FFmpeg Required">
  </a>
</p>

A real-time ASCII camera for your terminal.

TermiCam is a Go TUI application that reads camera frames through FFmpeg and renders them as ASCII art in your terminal. It supports Linux, macOS, and Windows camera capture through platform-specific FFmpeg backends.

TermiCam supports both real and virtual camera inputs, as well as lightweight session recording and playback.

<img width="720" height="406" alt="Demonstration of TermiCam converting a live video feed into real-time ASCII art inside a terminal window" src="https://github.com/user-attachments/assets/e921eebd-24bd-4027-b8dd-75b086ec8039" />

---

## Requirements

- **Go** to build the app.
- **FFmpeg** for camera capture. `ffmpeg` must be available in your `PATH`.
- A terminal with enough space to render the camera view.

Platform-specific requirements:

- **Linux:** install `v4l2-ctl` from `v4l-utils`.
- **macOS:** grant camera permission to the terminal app you use to run TermiCam.
- **Windows:** use an FFmpeg build with DirectShow support. Most common Windows FFmpeg builds include this.

---

## Installation

### Pre-compiled Binaries (Recommended)

If you do not have Go installed, you can download a pre-compiled binary for your operating system and architecture directly from the [GitHub Releases](https://github.com/Megge06/TermiCam/releases) page.

> [!WARNING]
> FFmpeg is still needed when running the binaries, it is not compiled into the binaries.

#### Linux & macOS

1. Download the archive for your system (e.g., `termicam_Darwin_x86_64.tar.gz` or `termicam_Linux_arm64.tar.gz`).
2. Extract the archive and make the binary executable:

   ```sh
   tar -xzf termicam_*.tar.gz
   chmod +x termicam
   ```

3. (Optional) Move it to your local path to run it from anywhere:
   ```sh
   sudo mv termicam /usr/local/bin/
   ```

> [!TIP]
> **macOS Users:** Because the binary is not signed/notarized by Apple, you may see a "Developer Cannot Be Verified" warning on the first run. To allow execution, you can run:
> `xattr -d com.apple.quarantine termicam`

#### Windows

1. Download the `.zip` archive for your architecture (e.g., `termicam_Windows_x86_64.zip`).
2. Extract the folder.
3. Run `termicam.exe` from PowerShell or Command Prompt, or add the folder to your System Environment variables (`PATH`) to run it globally.

### Via Go

If you have Go installed on your machine, you can install TermiCam directly to your `$GOPATH/bin`:

```sh
go install github.com/Megge06/TermiCam/cmd/termicam@latest
```

---

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

---

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

---

## Usage

### Live Mode

Start the app:

```sh
go run ./cmd/termicam
```

The app starts on a settings screen. Configure the display options, proceed to device selection, choose a camera, and press Enter to start the ASCII camera view.

### Playback Mode

You can play back previously recorded `.tcam` video sessions directly from the command line:

```sh
go run ./cmd/termicam path/to/recording.tcam
```

This launches TermiCam directly in a lightweight streaming viewer that loops the file infinitely. Press **Esc** or **q** to close the program.

---

## Using as a Virtual Camera (OBS Studio)

If you want to use your live ASCII feed as a system webcam for video calls or other things, you can easily route TermiCam's output through **OBS Studio** without installing custom virtual camera drivers:

1. **Start TermiCam:** Run the application in your preferred terminal emulator.
2. **Open OBS Studio:** Download and launch [OBS Studio](https://obsproject.com/).
3. **Add a Capture Source:** Under the **Sources** panel, click the `+` button and add Window Capture.
4. **Select your Terminal:** Configure the source to target your open terminal window running TermiCam.
5. **Crop the Window Borders:** Adjust the size of the window in the preview to fill it completely with the output from TermiCam.
6. **Start the Virtual Camera:** Click **Start Virtual Camera** under the _Controls_ panel in the bottom-right of OBS.
7. **Select OBS as your Input:** In your video application, change your input device to **OBS Virtual Camera**.

---

## Keyboard Shortcuts

### Settings Screen

- **Up / Down (`j` / `k`):** Move between configuration items.
- **Space / Left / Right:** Toggle settings or activate the Target FPS input field.
- **Enter:** Save settings and proceed to device selection.
- **Esc (while editing FPS):** Cancel editing and restore previous value.

### Device Selection

- **Up / Down:** Navigate the list of detected video devices.
- **`/`:** Enter search mode to filter devices by name or path.
- **Space:** Check/uncheck a device.
- **Enter:** Connect to the selected/checked device and start the camera stream.
- **Esc / Backspace:** Return to the Settings screen.

### Camera View

- **`h`:** Toggle the HUD sidebar on and off for a minimal, full-screen ASCII display.
- **`r`:** Toggle recording on/off (only available in live mode).
- **Esc / Backspace:** Stop the active session and return to the Device Selection screen (or exit immediately if in Playback Mode).
- **`q` / Ctrl+C:** Terminate the application immediately.

---

## Development Commands

Run linting:

```sh
golangci-lint run
```

Format Go files:

```sh
gofmt -w .
```

Build for the current platform:

```sh
go build -o termicam ./cmd/termicam
```

---

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

---

## Project Structure

- `cmd/termicam/main.go`: application entry point.
- `internal/tui/`: Bubble Tea model, update loop, views, and styling.
- `internal/video/`: platform-specific camera discovery and FFmpeg capture sessions.
- `internal/ascii/`: RGB24 frame to ASCII conversion.

---

## Built With

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - A powerful little TUI framework, used to power TermiCam's interface.
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** - Style definitions for nice terminal layouts, used for styling and rendering in the TUI.
- **[FFmpeg](https://github.com/ffmpeg/ffmpeg)** - A complete, cross-platform solution to record, convert and stream audio and video, used for getting a stream of the camera video.

## Acknowledgments

- **Algorithm Inspiration:** The core image-to-ASCII conversion logic, character mapping, and 8-bit color scaling in this project were heavily inspired by the fantastic [ascii-image-converter](https://github.com/TheZoraiz/ascii-image-converter) by [Zoraiz Hassan](https://github.com/TheZoraiz).
