//go:build darwin

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
	out, err := exec.Command("ffmpeg", "-hide_banner", "-f", "avfoundation", "-list_devices", "true", "-i", "").CombinedOutput()
	devices := parseAVFoundationDevices(string(out))
	if len(devices) > 0 {
		return devices, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list avfoundation devices: %w: %s", err, strings.TrimSpace(string(out)))
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
		"-f", "avfoundation",
		"-framerate", "30",
		"-i", fmt.Sprintf("%s:none", device.ID),
		"-frames:v", "1",
		"-f", "null",
		"-",
	).CombinedOutput()

	w, h, parseErr := parseAVFoundationResolution(string(out))
	if parseErr == nil {
		return w, h, nil
	}

	if ctx.Err() != nil {
		return 640, 480, fmt.Errorf("failed to probe avfoundation resolution: %w", ctx.Err())
	}
	if err != nil {
		return 640, 480, fmt.Errorf("failed to probe avfoundation resolution: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return 640, 480, parseErr
}

func NewSession(device Device, width, height, fps int) (*Session, error) {
	if device.ID == "" {
		return nil, fmt.Errorf("empty video device id")
	}

	args := []string{
		"-hide_banner",
		"-loglevel", "error",
		"-nostdin",
		"-fflags", "nobuffer",
		"-flags", "low_delay",
		"-f", "avfoundation",
		"-framerate", fmt.Sprintf("%d", fps),
		"-i", fmt.Sprintf("%s:none", device.ID),
		"-vf", fmt.Sprintf("scale=%d:%d", width, height),
		"-r", fmt.Sprintf("%d", fps),
		"-f", "rawvideo",
		"-pix_fmt", "rgb24",
		"pipe:1",
	}

	return newSession(width, height, args)
}

func parseAVFoundationDevices(output string) []Device {
	deviceLine := regexp.MustCompile(`\[(\d+)\]\s+(.+)$`)
	devices := make([]Device, 0)
	inVideoDevices := false

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "AVFoundation video devices:") {
			inVideoDevices = true
			continue
		}

		if strings.Contains(trimmed, "AVFoundation audio devices:") {
			inVideoDevices = false
			continue
		}

		if !inVideoDevices {
			continue
		}

		matches := deviceLine.FindStringSubmatch(trimmed)
		if len(matches) < 3 {
			continue
		}

		devices = append(devices, Device{
			ID:   matches[1],
			Name: strings.TrimSpace(matches[2]),
		})
	}

	return devices
}

func parseAVFoundationResolution(output string) (int, int, error) {
	dimensions := regexp.MustCompile(`^(\d+)x(\d+)(?:\s|$)`)

	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "Video:") {
			continue
		}

		for _, field := range strings.Split(line, ",") {
			matches := dimensions.FindStringSubmatch(strings.TrimSpace(field))
			if len(matches) < 3 {
				continue
			}

			w, _ := strconv.Atoi(matches[1])
			h, _ := strconv.Atoi(matches[2])
			if w > 0 && h > 0 {
				return w, h, nil
			}
		}
	}

	return 640, 480, fmt.Errorf("could not parse resolution from avfoundation output")
}
