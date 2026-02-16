package cmd

import (
	"os"
	"path/filepath"
)

// getConfigDir returns ~/.config/zenkai (or platform equivalent).
func getConfigDir() (string, error) {

	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(base, "zenkai")

	// Ensure directory exists
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return "", err
	}

	return configPath, nil
}
