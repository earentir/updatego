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
	rows := make([]statusRow, 0, 10)

	if !utils.IsDirExists(goFullPath) {
		printStatusTable([]statusRow{{
			label:  "Directory",
			kind:   statusFail,
			detail: "missing (" + goFullPath + ")",
		}})
		return nil
	}
	rows = append(rows, statusRow{label: "Directory", kind: statusOK, detail: goFullPath})

	goVersion, err := utils.CheckGoVersion(goFullPath)
	if err != nil {
		rows = append(rows, statusRow{label: "Version", kind: statusFail, detail: err.Error()})
	} else {
		version, osArch := utils.ParseGoVersion(goVersion)
		rows = append(rows,
			statusRow{label: "Version", kind: statusOK, detail: version},
			statusRow{label: "OS/Arch", kind: statusInfo, detail: osArch},
		)
	}

	if utils.IsWritable(goFullPath) {
		rows = append(rows, statusRow{label: "Writable", kind: statusOK, detail: "yes"})
	} else {
		rows = append(rows, statusRow{label: "Writable", kind: statusWarn, detail: "no"})
	}

	rows = append(rows,
		statusRow{label: "Install type", kind: statusInfo, detail: utils.DetermineInstallType(goFullPath)},
		statusRow{label: "Extract root", kind: statusInfo, detail: config.GlobalConfig.GoExtractPathRoot},
	)

	if goroot := os.Getenv("GOROOT"); goroot == "" {
		// Modern Go derives GOROOT from the binary; unset is normal.
		rows = append(rows, statusRow{label: "GOROOT", kind: statusOK, detail: "unset (derived from binary)"})
	} else if goroot == goFullPath {
		rows = append(rows, statusRow{label: "GOROOT", kind: statusOK, detail: goroot})
	} else {
		rows = append(rows, statusRow{label: "GOROOT", kind: statusFail, detail: goroot + " (want " + goFullPath + ")"})
	}

	expectedGOPATH := filepath.Join(os.Getenv("HOME"), "go")
	if gopath := os.Getenv("GOPATH"); gopath == "" {
		rows = append(rows, statusRow{label: "GOPATH", kind: statusOK, detail: "unset (defaults to " + expectedGOPATH + ")"})
	} else if gopath == expectedGOPATH {
		rows = append(rows, statusRow{label: "GOPATH", kind: statusOK, detail: gopath})
	} else {
		rows = append(rows, statusRow{label: "GOPATH", kind: statusWarn, detail: gopath + " (default would be " + expectedGOPATH + ")"})
	}

	if utils.IsGoInPath(goFullPath) {
		rows = append(rows, statusRow{label: "PATH", kind: statusOK, detail: "go binary present"})
	} else {
		rows = append(rows, statusRow{label: "PATH", kind: statusFail, detail: "go binary not in PATH"})
	}

	toolchain := GOTOOLCHAINForStatus()
	kind := statusOK
	if strings.HasPrefix(toolchain, "unavailable") {
		kind = statusWarn
	}
	rows = append(rows, statusRow{label: "GOTOOLCHAIN", kind: kind, detail: toolchain})

	printStatusTable(rows)
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
