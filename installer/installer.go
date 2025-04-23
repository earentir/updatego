// Package installer provides functions to install Go versions
package installer

import (
	"fmt"
	"path/filepath"

	"updatego/config"
	"updatego/local"
	"updatego/utils"
)

var (
	tempDir = ""
)

// InstallGo installs the specified Go version
func InstallGo(version string, force, global, user bool, customPath string) {
	config.GlobalConfig.GoExtractPathRoot = utils.DetermineInstallPath(global, user, customPath)
	config.GlobalConfig.GoFullPath = filepath.Join(config.GlobalConfig.GoExtractPathRoot, "go")

	if !utils.IsWritable(config.GlobalConfig.GoExtractPathRoot) {
		fmt.Printf("Error: Installation directory %s is not writable\n", config.GlobalConfig.GoExtractPathRoot)
		return
	}

	if version == "" {
		version = utils.GetVersionToInstall("")
	}

	if utils.DirNotEmpty(config.GlobalConfig.GoFullPath) {
		handleExistingInstallation(version, force)
	} else {
		installNewVersion(version)
	}

	utils.SetEnvironmentVariables(config.GlobalConfig.GoFullPath)
}

func handleExistingInstallation(version string, force bool) {
	currentVersion, err := utils.CheckGoVersion(config.GlobalConfig.GoFullPath)
	if err != nil {
		fmt.Printf("Error checking current Go version: %v\n", err)
		fmt.Println("Proceeding with installation...")
		installNewVersion(version)
		return
	}

	parsedCurrentVersion, _ := utils.ParseGoVersion(currentVersion)
	fmt.Printf("Go is already installed. Current version: %s\n", parsedCurrentVersion)

	if parsedCurrentVersion == version && !force {
		fmt.Printf("Go version %s is already installed. Use the --force flag to reinstall it.\n", version)
		return
	}

	localPath := filepath.Join(config.GlobalConfig.GoExtractPathRoot, "go-"+version)
	if utils.IsDirExists(localPath) {
		fmt.Printf("Go version %s is already available locally. Switching to this version.\n", version)
		local.SwitchGoVersion(version)
		return
	}

	if force {
		backupPath := filepath.Join(config.GlobalConfig.GoExtractPathRoot, "go-"+parsedCurrentVersion)
		utils.BackupOldGo(backupPath, config.GlobalConfig.GoFullPath)
	}

	installNewVersion(version)
}

func installNewVersion(version string) {
	targetPath := filepath.Join(config.GlobalConfig.GoExtractPathRoot, "go-"+version)
	fmt.Printf("Installing Go version: %s\n", version)
	if err := installGoVersion(version, targetPath, false); err != nil {
		fmt.Printf("Error installing Go version %s: %v\n", version, err)
		return
	}

	fmt.Printf("Switching to the newly installed Go version: %s\n", version)
	local.SwitchGoVersion(version)
}

func installGoVersion(version, installPath string, isMainGoDir bool) error {
	filename := utils.BuildFilename(version)
	fmt.Printf("Downloading %s, writing to: %s\n", filename, filepath.Join(tempDir, filename))
	filePath, err := utils.DownloadFileWithProgress(utils.GoDownloadURL + filename)
	if err != nil {
		return fmt.Errorf("error downloading the file: %v", err)
	}

	fmt.Println("Extracting the new Go version...")
	if err := utils.ExtractTarGz(filePath, installPath, isMainGoDir); err != nil {
		return fmt.Errorf("error extracting the Go archive: %v", err)
	}

	return nil
}
