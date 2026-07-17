# TermiCam

A real-time ASCII camera for your terminal.

## Requirements

- **Go 1.26.5+** to build the app.
- **FFmpeg** for camera capture.
- **Linux:** `v4l2-ctl` from `v4l-utils`.
- **macOS:** install FFmpeg with `brew install ffmpeg` and grant camera permission to the terminal app you use to run TermiCam.

## Built With

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - A powerful little TUI framework, used to power TermiCam's interface.
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** - Style definitions for nice terminal layouts, used for styling and rendering in the TUI.
- **[FFmpeg](https://github.com/ffmpeg/ffmpeg)** - A complete, cross-platform solution to record, convert and stream audio and video, used for getting a stream of the camera video.

## Acknowledgments

- **Algorithm Inspiration:** The core image-to-ASCII conversion logic, character mapping, and 8-bit color scaling in this project were heavily inspired by the fantastic [ascii-image-converter](https://github.com/TheZoraiz/ascii-image-converter) by [Zoraiz Hassan](https://github.com/TheZoraiz).
