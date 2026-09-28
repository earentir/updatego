package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const defaultExtractRoot = "/usr/local"

type fileConfig struct {
	ExtractRoot string `json:"extract_root"`
}

// ConfigFilePath returns the path to the user config file.
func ConfigFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "updatego", "config.json"), nil
}

// LoadExtractRoot reads extract_root from disk, or returns the default.
func LoadExtractRoot() (string, error) {
	path, err := ConfigFilePath()
	if err != nil {
		return defaultExtractRoot, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaultExtractRoot, nil
		}
		return defaultExtractRoot, err
	}
	var fc fileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return defaultExtractRoot, err
	}
	root := filepath.Clean(fc.ExtractRoot)
	if root == "" || root == "." {
		return defaultExtractRoot, nil
	}
	return root, nil
}

// SaveExtractRoot writes extract_root to the config file.
func SaveExtractRoot(extractRoot string) error {
	path, err := ConfigFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	root := filepath.Clean(extractRoot)
	data, err := json.MarshalIndent(fileConfig{ExtractRoot: root}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ApplyExtractRoot sets GlobalConfig paths from an extract root directory.
func ApplyExtractRoot(extractRoot string) {
	root := filepath.Clean(extractRoot)
	if root == "" {
		root = defaultExtractRoot
	}
	GlobalConfig.GoExtractPathRoot = root
	GlobalConfig.GoFullPath = filepath.Join(root, "go")
}

// LoadIntoGlobal loads extract_root from disk into GlobalConfig.
func LoadIntoGlobal() error {
	root, err := LoadExtractRoot()
	if err != nil {
		ApplyExtractRoot(defaultExtractRoot)
		return err
	}
	ApplyExtractRoot(root)
	return nil
}
