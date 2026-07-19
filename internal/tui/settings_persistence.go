package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// PersistedSettings mirrors the configurable fields in model.go
type PersistedSettings struct {
	Color    bool `json:"color"`
	Detailed bool `json:"detailed"`
	Mirror   bool `json:"mirror"`
	FPS      int  `json:"fps"`
}

// DefaultSettings returns the fallback settings if no config file exists.
func DefaultSettings() PersistedSettings {
	return PersistedSettings{
		Color:    false,
		Detailed: false,
		Mirror:   false,
		FPS:      30,
	}
}

// getConfigPath returns the platform-specific path
func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	// Store in a subdirectory specific to TermiCam
	appDir := filepath.Join(configDir, "termicam")
	return filepath.Join(appDir, "config.json"), nil
}

// LoadSettings attempts to read and parse the configuration file from disk
func LoadSettings() PersistedSettings {
	defaults := DefaultSettings()

	path, err := getConfigPath()
	if err != nil {
		return defaults
	}

	// #nosec G304 -- Path is resolved safely from os.UserConfigDir and does not take user input
	data, err := os.ReadFile(path)
	if err != nil {
		return defaults
	}

	var saved PersistedSettings
	if err := json.Unmarshal(data, &saved); err != nil {
		return defaults
	}

	// Simple validation for FPS to ensure it isn't invalid or zero
	if saved.FPS <= 0 {
		saved.FPS = defaults.FPS
	}

	return saved
}

// SaveSettings writes the current configuration to disk.
func SaveSettings(settings PersistedSettings) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	// Ensure the parent directory exists
	err = os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	// #nosec G304 -- Path is resolved safely from os.UserConfigDir and does not take user input
	return os.WriteFile(path, data, 0600)
}
