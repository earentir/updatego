// Package local provides functions to manage the Go installation locally
package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"updatego/config"
	"updatego/utils"
)

var (
	goExtractPathRoot = "/usr/local/"
	goFullPath        = ""
)

func init() {
	goFullPath = filepath.Join(goExtractPathRoot, "go")
}

// CheckGoStatus checks the status of the Go installation
func CheckGoStatus() {
	goFullPath = filepath.Join(goExtractPathRoot, "go")

	// Check if Go directory exists
	if !utils.IsDirExists(goFullPath) {
		fmt.Println("Go directory does not exist. ❌")
		return
	}
	fmt.Println("Go directory exists. ✅")

	// Check Go version
	goVersion, err := utils.CheckGoVersion(goFullPath)
	if err != nil {
		fmt.Println("Error checking Go version: ❌", err)
	} else {
		version, osArch := utils.ParseGoVersion(goVersion)
		fmt.Printf("Go version: %s ✅\n", version)
		fmt.Printf("OS/Arch: %s ✅\n", osArch)
	}

	// Check if Go is writable
	if utils.IsWritable(goFullPath) {
		fmt.Println("Go directory is writable. ✅")
	} else {
		fmt.Println("Go directory is not writable. ❌")
	}

	// Check if install is user or global
	installType := utils.DetermineInstallType(goFullPath)
	fmt.Printf("Install type: %s ✅\n", installType)

	// Check GOROOT environment variable
	if os.Getenv("GOROOT") == goFullPath {
		fmt.Println("GOROOT environment variable is set correctly. ✅")
	} else {
		fmt.Println("GOROOT environment variable is not set correctly. ❌")
	}

	// Check GOPATH environment variable
	expectedGOPATH := filepath.Join(os.Getenv("HOME"), "go")
	if os.Getenv("GOPATH") == expectedGOPATH {
		fmt.Println("GOPATH environment variable is set correctly. ✅")
	} else {
		fmt.Println("GOPATH environment variable is not set correctly. ❌")
	}

	// Check if `go` is in PATH
	if utils.IsGoInPath(goFullPath) {
		fmt.Println("`go` binary is in PATH. ✅")
	} else {
		fmt.Println("`go` binary is not in PATH. ❌")
	}
}

// PrintLatestGoVersion prints the latest Go version available
func PrintLatestGoVersion() {
	version, err := utils.GetLatestVersion()
	if err != nil {
		fmt.Println("Error finding the latest version:", err)
		os.Exit(1)
	}
	fmt.Println("Latest version available:", version)
}

// ListLocalVersions lists all local Go versions
func ListLocalVersions() {
	goFullPath = filepath.Join(goExtractPathRoot, "go")

	// List the current Go version
	if goVersion, err := utils.CheckGoVersion(goFullPath); err == nil {
		version, _ := utils.ParseGoVersion(goVersion)
		fmt.Printf("Current Go version: %s\n", version)
	} else {
		fmt.Println("No current Go version found.")
	}

	// List other local Go versions
	err := filepath.Walk(goExtractPathRoot, func(path string, info os.FileInfo, err error) error {
		if strings.HasPrefix(info.Name(), "go-") && info.IsDir() {
			version := strings.TrimPrefix(info.Name(), "go-")
			fmt.Printf("Local Go version: %s\n", version)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error listing local Go versions:", err)
	}
}

// SwitchGoVersion switches to a specific Go version
func SwitchGoVersion(version string) {
	targetPath := filepath.Join(config.GlobalConfig.GoExtractPathRoot, "go-"+version)

	if isAlreadyOnVersion(version) {
		return
	}

	ensureVersionExists(version, targetPath)

	backupCurrentVersion()

	if err := os.Rename(targetPath, config.GlobalConfig.GoFullPath); err != nil {
		fmt.Printf("Error switching to Go version %s: %v\n", version, err)
		return
	}

	fmt.Printf("Switched to Go version %s successfully.\n", version)
}

func isAlreadyOnVersion(version string) bool {
	currentVersion, err := utils.CheckGoVersion(config.GlobalConfig.GoFullPath)
	if err == nil {
		parsedCurrentVersion, _ := utils.ParseGoVersion(currentVersion)
		if parsedCurrentVersion == version {
			fmt.Printf("Already using Go version %s\n", version)
			return true
		}
	}
	return false
}

func ensureVersionExists(version, targetPath string) {
	if !utils.IsDirExists(targetPath) {
		fmt.Printf("Go version %s not found locally. Downloading...\n", version)
		downloadAndExtractVersion(version, targetPath)
	}
}

func downloadAndExtractVersion(version, targetPath string) {
	filename := utils.BuildFilename(version)
	filePath, err := utils.DownloadAndVerifyFile(utils.GoDownloadURL + filename)
	if err != nil {
		fmt.Printf("Error downloading Go version %s: %v\n", version, err)
		return
	}

	if err := os.MkdirAll(targetPath, 0755); err != nil {
		fmt.Printf("Error creating directory for Go version %s: %v\n", version, err)
		return
	}

	fmt.Println("Extracting the Go version...")
	if err := utils.ExtractTarGz(filePath, targetPath, false); err != nil {
		fmt.Printf("Error extracting Go version %s: %v\n", version, err)
		return
	}
}

func backupCurrentVersion() {
	if utils.IsDirExists(config.GlobalConfig.GoFullPath) {
		currentVersion, err := utils.CheckGoVersion(config.GlobalConfig.GoFullPath)
		if err != nil {
			fmt.Printf("Error checking current Go version: %v\n", err)
			return
		}
		parsedCurrentVersion, _ := utils.ParseGoVersion(currentVersion)
		currentBackupPath := filepath.Join(config.GlobalConfig.GoExtractPathRoot, "go-"+parsedCurrentVersion)
		if err := os.Rename(config.GlobalConfig.GoFullPath, currentBackupPath); err != nil {
			fmt.Printf("Error backing up current Go version: %v\n", err)
			return
		}
	}
}
