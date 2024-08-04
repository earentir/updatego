// Package local provides functions to manage the Go installation locally
package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	goFullPath = filepath.Join(goExtractPathRoot, "go")
	targetPath := filepath.Join(goExtractPathRoot, "go-"+version)

	// Check if the target version exists locally
	if !utils.IsDirExists(targetPath) {
		fmt.Printf("Go version %s not found locally. Downloading...\n", version)

		filename := utils.BuildFilename(version)
		filePath, err := utils.DownloadAndVerifyFile(utils.GoDownloadURL + filename)
		if err != nil {
			fmt.Println("Error downloading the file:", err)
			os.Exit(1)
		}

		if err := os.Mkdir(targetPath, 0755); err != nil {
			fmt.Println("Error creating directory for the new Go version:", err)
			os.Exit(1)
		}

		fmt.Println("Extracting the Go version...")
		if err := utils.ExtractTarGz(filePath, targetPath, false); err != nil {
			fmt.Println("Error extracting the Go archive:", err)
			os.Exit(1)
		}
	}

	// Rename the current Go folder
	if utils.IsDirExists(goFullPath) {
		currentVersion, _ := utils.CheckGoVersion(goFullPath)
		parsedCurrentVersion, _ := utils.ParseGoVersion(currentVersion)
		currentBackupPath := filepath.Join(goExtractPathRoot, "go-"+parsedCurrentVersion)
		if err := os.Rename(goFullPath, currentBackupPath); err != nil {
			fmt.Println("Error renaming the current Go folder:", err)
			os.Exit(1)
		}
	}

	// Switch to the target version
	if err := os.Rename(targetPath, goFullPath); err != nil {
		fmt.Println("Error renaming the target Go folder:", err)
		os.Exit(1)
	}

	fmt.Printf("Switched to Go version %s successfully.\n", version)
}
