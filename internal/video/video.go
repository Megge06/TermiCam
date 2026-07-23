package video

import (
	"context"
	"fmt"
	"io"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"sync"
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
	stderr *stderrBuffer
	cancel context.CancelFunc
	width  int
	height int
}

type stderrBuffer struct {
	mu    sync.Mutex
	data  []byte
	limit int
}

func newStderrBuffer(limit int) *stderrBuffer {
	return &stderrBuffer{limit: limit}
}

func (b *stderrBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.limit <= 0 {
		return len(p), nil
	}
	if len(p) >= b.limit {
		b.data = append(b.data[:0], p[len(p)-b.limit:]...)
		return len(p), nil
	}

	b.data = append(b.data, p...)
	if overflow := len(b.data) - b.limit; overflow > 0 {
		copy(b.data, b.data[overflow:])
		b.data = b.data[:b.limit]
	}

	return len(p), nil
}

func (b *stderrBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return strings.TrimSpace(string(b.data))
}

type captureMode struct {
	width      int
	height     int
	framerates []float64
}

func chooseCaptureFramerate(targetFPS, width, height int, modes []captureMode) float64 {
	candidates := make([]float64, 0)
	for _, mode := range modes {
		if mode.width == width && mode.height == height {
			candidates = append(candidates, mode.framerates...)
		}
	}

	if len(candidates) == 0 {
		for _, mode := range modes {
			candidates = append(candidates, mode.framerates...)
		}
	}

	return chooseNearestFramerate(targetFPS, candidates)
}

func chooseNearestFramerate(targetFPS int, candidates []float64) float64 {
	if targetFPS <= 0 {
		targetFPS = 30
	}

	target := float64(targetFPS)
	best := 0.0
	bestDiff := math.MaxFloat64

	for _, candidate := range candidates {
		if candidate <= 0 {
			continue
		}

		diff := math.Abs(candidate - target)
		if best == 0 || diff < bestDiff || (diff == bestDiff && candidate > best) {
			best = candidate
			bestDiff = diff
		}
	}

	return best
}

func formatFramerate(fps float64) string {
	return strconv.FormatFloat(fps, 'f', -1, 64)
}

func newSession(width, height int, args []string) (*Session, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// #nosec G204 -- Safe. The args array is constructed entirely inside internal platform-specific files using static system parameters.
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	stderr := newStderrBuffer(8192)
	cmd.Stderr = stderr

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
		stderr: stderr,
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
	if err == nil {
		return nil
	}

	if stderr := s.stderr.String(); stderr != "" {
		return fmt.Errorf("failed to read ffmpeg frame: %w: %s", err, stderr)
	}

	return err
}

// Close gracefully cancels the execution
func (s *Session) Close() error {
	s.cancel()

	// Wait for the process to exit to prevent zombie child processes
	_ = s.cmd.Wait()

	return s.stdout.Close()
}
