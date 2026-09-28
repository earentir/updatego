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

// CheckGoStatus checks the status of the Go installation.
func CheckGoStatus() error {
	goFullPath := config.GlobalConfig.GoFullPath

	if !utils.IsDirExists(goFullPath) {
		fmt.Println("Go directory does not exist.")
		return nil
	}
	fmt.Println("Go directory exists.")

	goVersion, err := utils.CheckGoVersion(goFullPath)
	if err != nil {
		fmt.Println("Error checking Go version:", err)
	} else {
		version, osArch := utils.ParseGoVersion(goVersion)
		fmt.Printf("Go version: %s\n", version)
		fmt.Printf("OS/Arch: %s\n", osArch)
	}

	if utils.IsWritable(goFullPath) {
		fmt.Println("Go directory is writable.")
	} else {
		fmt.Println("Go directory is not writable.")
	}

	installType := utils.DetermineInstallType(goFullPath)
	fmt.Printf("Install type: %s\n", installType)
	fmt.Printf("Extract root: %s\n", config.GlobalConfig.GoExtractPathRoot)

	if os.Getenv("GOROOT") == goFullPath {
		fmt.Println("GOROOT environment variable is set correctly.")
	} else {
		fmt.Println("GOROOT environment variable is not set correctly.")
	}

	expectedGOPATH := filepath.Join(os.Getenv("HOME"), "go")
	if os.Getenv("GOPATH") == expectedGOPATH {
		fmt.Println("GOPATH environment variable is set correctly.")
	} else {
		fmt.Println("GOPATH environment variable is not set correctly.")
	}

	if utils.IsGoInPath(goFullPath) {
		fmt.Println("`go` binary is in PATH.")
	} else {
		fmt.Println("`go` binary is not in PATH.")
	}
	return nil
}

// PrintLatestGoVersion prints the latest Go version available.
func PrintLatestGoVersion() error {
	version, err := utils.GetLatestVersion()
	if err != nil {
		return err
	}
	fmt.Println("Latest version available:", version)
	return nil
}

// ListLocalVersions lists all local Go versions.
func ListLocalVersions() error {
	goFullPath := config.GlobalConfig.GoFullPath
	root := config.GlobalConfig.GoExtractPathRoot

	if goVersion, err := utils.CheckGoVersion(goFullPath); err == nil {
		version, _ := utils.ParseGoVersion(goVersion)
		fmt.Printf("Current Go version: %s\n", version)
	} else {
		fmt.Println("No current Go version found.")
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), "go-") {
			version := strings.TrimPrefix(info.Name(), "go-")
			fmt.Printf("Local Go version: %s\n", version)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error listing local Go versions: %w", err)
	}
	return nil
}

// SwitchGoVersion switches to a specific Go version.
func SwitchGoVersion(version string) error {
	if err := utils.RefuseWindows(); err != nil {
		return err
	}

	targetPath := filepath.Join(config.GlobalConfig.GoExtractPathRoot, "go-"+version)

	if isAlreadyOnVersion(version) {
		return nil
	}

	if err := ensureVersionExists(version, targetPath); err != nil {
		return err
	}
	if !utils.GoTreeReady(targetPath) {
		return fmt.Errorf("Go version %s is not available at %s", version, targetPath)
	}

	if err := backupCurrentVersion(targetPath); err != nil {
		return err
	}

	if err := os.Rename(targetPath, config.GlobalConfig.GoFullPath); err != nil {
		return fmt.Errorf("error switching to Go version %s: %w", version, err)
	}

	fmt.Printf("Switched to Go version %s successfully.\n", version)
	return nil
}

func isAlreadyOnVersion(version string) bool {
	currentVersion, err := utils.CheckGoVersion(config.GlobalConfig.GoFullPath)
	if err != nil {
		return false
	}
	parsedCurrentVersion, osArch := utils.ParseGoVersion(currentVersion)
	if parsedCurrentVersion != version {
		return false
	}
	if !utils.IsHostPlatform(osArch) {
		fmt.Printf("Current Go %s is %s; this machine is %s. Replacing it.\n", version, osArch, utils.HostOSArch())
		return false
	}
	fmt.Printf("Already using Go version %s\n", version)
	return true
}

func ensureVersionExists(version, targetPath string) error {
	if utils.GoTreeReady(targetPath) {
		return nil
	}
	fmt.Printf("Go version %s not found locally. Downloading...\n", version)
	return downloadAndExtractVersion(version, targetPath)
}

func downloadAndExtractVersion(version, targetPath string) error {
	filename := utils.BuildFilename(version)
	filePath, err := utils.DownloadArchive(utils.GoDownloadURL + filename)
	if err != nil {
		return fmt.Errorf("error downloading Go version %s: %w", version, err)
	}

	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("error creating directory for Go version %s: %w", version, err)
	}

	fmt.Println("Extracting the Go version...")
	if err := utils.ExtractTarGz(filePath, targetPath, false); err != nil {
		return fmt.Errorf("error extracting Go version %s: %w", version, err)
	}
	if !utils.GoTreeReady(targetPath) {
		return fmt.Errorf("extracted tree at %s is not a usable Go installation", targetPath)
	}
	return nil
}

func backupCurrentVersion(targetPath string) error {
	if !utils.IsDirExists(config.GlobalConfig.GoFullPath) {
		return nil
	}

	backupName := "unusable"
	currentVersion, err := utils.CheckGoVersion(config.GlobalConfig.GoFullPath)
	if err != nil {
		fmt.Printf("Current Go installation cannot be executed (%v); moving it aside.\n", err)
	} else {
		parsedCurrentVersion, osArch := utils.ParseGoVersion(currentVersion)
		if parsedCurrentVersion != "" && parsedCurrentVersion != "Unknown version" {
			backupName = parsedCurrentVersion
		}
		if !utils.IsHostPlatform(osArch) {
			fmt.Printf("Current Go is %s; this machine is %s. Moving it aside.\n", osArch, utils.HostOSArch())
		}
	}

	baseName := "go-" + backupName
	if filepath.Join(config.GlobalConfig.GoExtractPathRoot, baseName) == targetPath {
		baseName = "go-" + backupName + "-previous"
	}
	currentBackupPath, err := utils.AllocateBackupDir(config.GlobalConfig.GoExtractPathRoot, baseName)
	if err != nil {
		return err
	}
	if err := utils.BackupOldGo(currentBackupPath, config.GlobalConfig.GoFullPath); err != nil {
		return err
	}
	fmt.Printf("Moved previous installation to %s\n", currentBackupPath)
	return nil
}
