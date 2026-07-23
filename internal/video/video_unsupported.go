//go:build !linux && !darwin && !windows

package video

import "fmt"

func ListDevices() ([]Device, error) {
	return nil, fmt.Errorf("video capture is only supported on Linux, macOS and Windows")
}

func GetDeviceResolution(device Device) (int, int, error) {
	return 640, 480, fmt.Errorf("video capture is only supported on Linux, macOS and Windows")
}

func GetDeviceFramerate(device Device, _, _, _ int) (float64, error) {
	return 0, fmt.Errorf("video capture is only supported on Linux, macOS and Windows")
}

func NewSession(device Device, _, _, width, height, fps int, _ float64) (*Session, error) {
	return nil, fmt.Errorf("video capture is only supported on Linux, macOS and Windows")
}
