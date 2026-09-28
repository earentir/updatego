package utils

import (
	"os"
	"path/filepath"

	"updatego/config"
)

// ResolveInstallRoot picks the extract root for an install command.
func ResolveInstallRoot(global, user bool, customPath string) (string, error) {
	if global {
		return "/usr/local", nil
	}
	if user {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		localRoot := filepath.Join(homeDir, ".local")
		if err := os.MkdirAll(localRoot, 0755); err != nil {
			return "", err
		}
		return localRoot, nil
	}
	if customPath != "" {
		return customPath, nil
	}
	return config.LoadExtractRoot()
}
