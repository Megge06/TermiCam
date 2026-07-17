//go:build !linux && !darwin

package video

import "fmt"

func ListDevices() ([]Device, error) {
	return nil, fmt.Errorf("video capture is only supported on Linux and macOS")
}

func GetDeviceResolution(device Device) (int, int, error) {
	return 640, 480, fmt.Errorf("video capture is only supported on Linux and macOS")
}

func NewSession(device Device, width, height, fps int) (*Session, error) {
	return nil, fmt.Errorf("video capture is only supported on Linux and macOS")
}
