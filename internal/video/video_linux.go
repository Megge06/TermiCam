//go:build linux

package video

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func ListDevices() ([]Device, error) {
	out, err := exec.Command("v4l2-ctl", "--list-devices").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	devices := make([]Device, 0)
	currentName := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "/dev/video") {
			name := currentName
			if name == "" {
				name = trimmed
			}

			devices = append(devices, Device{
				ID:   trimmed,
				Name: name,
			})
			continue
		}

		currentName = strings.TrimSuffix(trimmed, ":")
	}

	return devices, nil
}

func GetDeviceResolution(device Device) (int, int, error) {
	out, err := exec.Command("v4l2-ctl", "-d", device.ID, "--get-fmt-video").CombinedOutput()
	if err != nil {
		out, err = exec.Command("v4l2-ctl", "-d", device.ID, "--get-fmt-video-out").CombinedOutput()
		if err != nil {
			return 640, 480, err
		}
	}

	re := regexp.MustCompile(`Width/Height\s*:\s*(\d+)/(\d+)`)
	matches := re.FindStringSubmatch(string(out))
	if len(matches) < 3 {
		return 640, 480, fmt.Errorf("could not parse resolution from output")
	}

	w, _ := strconv.Atoi(matches[1])
	h, _ := strconv.Atoi(matches[2])

	return w, h, nil
}

func NewSession(device Device, _, _, width, height, fps int) (*Session, error) {
	if device.ID == "" {
		return nil, fmt.Errorf("empty video device id")
	}

	args := []string{
		"-nostdin",
		"-f", "v4l2",
		"-i", device.ID,
		"-vf", fmt.Sprintf("scale=%d:%d", width, height),
		"-r", fmt.Sprintf("%d", fps),
		"-f", "rawvideo",
		"-pix_fmt", "rgb24",
		"-",
	}

	return newSession(width, height, args)
}
