// Package config holds the configuration for the installer
package config

import (
	"os"
	"path/filepath"
)

// Config holds the configuration for the installer
type Config struct {
	GoExtractPathRoot string
	GoFullPath        string
	TempDir           string
	Verbose           bool
}

// GlobalConfig is the global configuration for the installer
var GlobalConfig = &Config{
	GoExtractPathRoot: "/usr/local/",
	TempDir:           os.TempDir(),
	Verbose:           false,
}

func init() {
	GlobalConfig.GoFullPath = filepath.Join(GlobalConfig.GoExtractPathRoot, "go")
}
