// Package installer provides functions to install Go versions
package installer

import (
	"fmt"
	"os"
	"path/filepath"

	"updatego/local"
	"updatego/utils"
)

var (
	tempDir           = ""
	goExtractPathRoot = "/usr/local/"
	goFullPath        = ""
)

// InstallGo installs the specified Go version
func InstallGo(version string, force, global, user bool, customPath string) {
	tempDir = os.TempDir()

	// Determine the installation path based on flags
	goExtractPathRoot = utils.DetermineInstallPath(global, user, customPath)

	if !utils.IsWritable(goExtractPathRoot) {
		fmt.Printf("Error: Installation directory %s is not writable\n", goExtractPathRoot)
		return
	}

	goFullPath = filepath.Join(goExtractPathRoot, "go")

	// Check if Go is already installed
	if utils.DirNotEmpty(goFullPath) {
		goVersion, err := utils.CheckGoVersion(goFullPath)
		if err == nil {
			parsedGoVersion, _ := utils.ParseGoVersion(goVersion)
			fmt.Printf("Go is already installed. Current version: %s\n", parsedGoVersion)

			if version == "" {
				fmt.Println("Please provide a version to install or use the --force flag to reinstall the current version.")
				return
			}

			if parsedGoVersion == version {
				if force {
					backupPath := filepath.Join(goExtractPathRoot, "go-"+parsedGoVersion)
					utils.BackupOldGo(backupPath, goFullPath)
				} else {
					fmt.Printf("Go version %s is already installed. Use the --force flag to reinstall it.\n", version)
					return
				}
			} else {
				// Check if the requested version is already available locally
				localPath := filepath.Join(goExtractPathRoot, "go-"+version)
				if utils.IsDirExists(localPath) {
					fmt.Printf("Go version %s is already available locally. Switching to this version.\n", version)
					local.SwitchGoVersion(version)
					return
				}
			}
		} else {
			fmt.Printf("Error checking current Go version: %v\n", err)
			fmt.Println("Proceeding with installation...")
		}
	} else {
		if version == "" {
			// No version provided, install the latest version in the goFullPath
			version = utils.GetVersionToInstall("")
			fmt.Printf("Installing latest Go version: %s\n", version)
			if err := installGoVersion(version, goFullPath, true); err != nil {
				fmt.Printf("Error installing Go version %s: %v\n", version, err)
				return
			}
		} else {
			// Specific version provided, install it in a go-version folder
			targetPath := filepath.Join(goExtractPathRoot, "go-"+version)
			fmt.Printf("Installing Go version: %s\n", version)
			if err := installGoVersion(version, targetPath, false); err != nil {
				fmt.Printf("Error installing Go version %s: %v\n", version, err)
				return
			}
		}
		return
	}

	// Install the specified version in a go-version folder
	targetPath := filepath.Join(goExtractPathRoot, "go-"+version)
	fmt.Printf("Installing Go version: %s\n", version)
	if err := installGoVersion(version, targetPath, false); err != nil {
		fmt.Printf("Error installing Go version %s: %v\n", version, err)
		return
	}

	if force {
		// Remove the old Go folder if --force flag is used
		fmt.Println("Removing the old Go folder...")
		if err := utils.RemoveGoFolder(goFullPath); err != nil {
			fmt.Printf("Error removing the old Go folder: %v\n", err)
			return
		}

		// Switch to the newly installed version
		fmt.Printf("Switching to the newly installed Go version: %s\n", version)
		local.SwitchGoVersion(version)
	}

	// Set environment variables
	utils.SetEnvironmentVariables(goFullPath)
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
