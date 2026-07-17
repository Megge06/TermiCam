package video

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

// Device identifies a camera as exposed by the host platform.
type Device struct {
	ID   string
	Name string
}

func (d Device) String() string {
	if d.Name == "" || d.Name == d.ID {
		return d.ID
	}

	return fmt.Sprintf("%s (%s)", d.Name, d.ID)
}

// Session manages the lifecycle of the background FFmpeg capture process.
type Session struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
	cancel context.CancelFunc
	width  int
	height int
}

func newSession(width, height int, args []string) (*Session, error) {
	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create ffmpeg stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start ffmpeg subprocess: %w", err)
	}

	return &Session{
		cmd:    cmd,
		stdout: stdout,
		cancel: cancel,
		width:  width,
		height: height,
	}, nil
}

// ReadFrame reads a single frame of video data into the buffer
func (s *Session) ReadFrame(dest []byte) error {
	expectedSize := s.width * s.height * 3
	if len(dest) < expectedSize {
		return fmt.Errorf("destination buffer too small: got %d bytes, need at least %d", len(dest), expectedSize)
	}

	_, err := io.ReadFull(s.stdout, dest[:expectedSize])
	return err
}

// Close gracefully cancels the execution
func (s *Session) Close() error {
	s.cancel()

	// Wait for the process to exit to prevent zombie child processes
	_ = s.cmd.Wait()

	return s.stdout.Close()
}
