//go:build windows

package video

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func ListDevices() ([]Device, error) {
	out, err := exec.Command("ffmpeg", "-hide_banner", "-f", "dshow", "-list_devices", "true", "-i", "dummy").CombinedOutput()
	devices := parseDirectShowDevices(string(out))
	if len(devices) > 0 {
		return devices, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list dshow devices: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return devices, nil
}

func GetDeviceResolution(device Device) (int, int, error) {
	if device.ID == "" {
		return 640, 480, fmt.Errorf("empty video device id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-hide_banner",
		"-nostdin",
		"-f", "dshow",
		"-list_options", "true",
		"-i", "video="+device.ID,
	).CombinedOutput()

	w, h, parseErr := parseDirectShowResolution(string(out))
	if parseErr == nil {
		return w, h, nil
	}

	if ctx.Err() != nil {
		return 640, 480, fmt.Errorf("failed to probe dshow resolution: %w", ctx.Err())
	}
	if err != nil {
		return 640, 480, fmt.Errorf("failed to probe dshow resolution: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return 640, 480, parseErr
}

func NewSession(device Device, captureWidth, captureHeight, width, height, fps int) (*Session, error) {
	if device.ID == "" {
		return nil, fmt.Errorf("empty video device id")
	}

	args := []string{
		"-hide_banner",
		"-loglevel", "error",
		"-nostdin",
		"-fflags", "nobuffer",
		"-flags", "low_delay",
		"-f", "dshow",
		"-video_size", fmt.Sprintf("%dx%d", captureWidth, captureHeight),
		"-framerate", fmt.Sprintf("%d", fps),
		"-i", "video=" + device.ID,
		"-vf", fmt.Sprintf("scale=%d:%d", width, height),
		"-r", fmt.Sprintf("%d", fps),
		"-f", "rawvideo",
		"-pix_fmt", "rgb24",
		"pipe:1",
	}

	return newSession(width, height, args)
}

func parseDirectShowDevices(output string) []Device {
	deviceLine := regexp.MustCompile(`^\s*"(.+)"\s*$`)
	altLine := regexp.MustCompile(`Alternative name\s+"(.+)"`)
	devices := make([]Device, 0)
	inVideoDevices := false
	var pending *Device

	flushPending := func() {
		if pending == nil {
			return
		}
		devices = append(devices, *pending)
		pending = nil
	}

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(stripFFmpegPrefix(line))
		if strings.Contains(trimmed, "DirectShow video devices") {
			flushPending()
			inVideoDevices = true
			continue
		}

		if strings.Contains(trimmed, "DirectShow audio devices") {
			flushPending()
			inVideoDevices = false
			continue
		}

		if !inVideoDevices {
			continue
		}

		if matches := deviceLine.FindStringSubmatch(trimmed); len(matches) >= 2 {
			flushPending()
			name := strings.TrimSpace(matches[1])
			pending = &Device{ID: name, Name: name}
			continue
		}

		if pending != nil {
			if matches := altLine.FindStringSubmatch(trimmed); len(matches) >= 2 {
				pending.ID = strings.TrimSpace(matches[1])
			}
		}
	}

	flushPending()

	return devices
}

func parseDirectShowResolution(output string) (int, int, error) {
	dimensions := regexp.MustCompile(`s=(\d+)x(\d+)`)
	bestW, bestH := 0, 0

	for _, matches := range dimensions.FindAllStringSubmatch(output, -1) {
		if len(matches) < 3 {
			continue
		}

		w, _ := strconv.Atoi(matches[1])
		h, _ := strconv.Atoi(matches[2])
		if w <= 0 || h <= 0 {
			continue
		}

		if w*h > bestW*bestH {
			bestW, bestH = w, h
		}
	}

	if bestW > 0 && bestH > 0 {
		return bestW, bestH, nil
	}

	return 640, 480, fmt.Errorf("could not parse resolution from dshow output")
}

func stripFFmpegPrefix(line string) string {
	closingBracket := strings.Index(line, "]")
	if closingBracket == -1 || closingBracket+1 >= len(line) {
		return line
	}

	return line[closingBracket+1:]
}
