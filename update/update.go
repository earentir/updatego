// Package update updates Go to the latest version
package update

import (
	"fmt"
	"os"
	"path/filepath"

	"updatego/config"
	"updatego/utils"
)

// Go updates Go to the latest version
func Go() {
	fmt.Println("Downloading Go Data to get the latest version...")
	htmlContent, err := utils.DownloadHTML(utils.GoDownloadURL)
	if err != nil {
		fmt.Println("Error downloading HTML content:", err)
		os.Exit(1)
	}

	fmt.Println("Finding the latest version...")
	version, err := utils.FindVersion(htmlContent)
	if err != nil {
		fmt.Println("Error finding the latest version:", err)
		os.Exit(1)
	}
	fmt.Println("Latest version found:", version)

	filename := utils.BuildFilename(version)

	fmt.Println("Downloading the latest version:", filename, " From: ", utils.GoDownloadURL+filename)
	fileURL := utils.GoDownloadURL + filename
	filePath, err := utils.DownloadFileWithProgress(fileURL)
	if err != nil {
		fmt.Println("Error downloading the file:", err)
		os.Exit(1)
	}

	if err := utils.VerifyDownloadedFile(filePath); err != nil {
		fmt.Printf("Error verifying downloaded file: %v\n", err)
		os.Exit(1)
	}

	if utils.DirNotEmpty(config.GlobalConfig.GoFullPath) {
		backupCurrentVersion()
	}

	fmt.Println("Extracting the new Go version...")
	if err := utils.ExtractTarGz(filePath, config.GlobalConfig.GoFullPath, true); err != nil {
		fmt.Println("Error extracting the Go archive:", err)
		os.Exit(1)
	}

	fmt.Printf("Go has been successfully updated to version %s\n", version)
}

func backupCurrentVersion() {
	backupName := "unusable"
	goVersion, err := utils.CheckGoVersion(config.GlobalConfig.GoFullPath)
	if err != nil {
		fmt.Printf("Current Go installation cannot be executed (%v); moving it aside.\n", err)
	} else {
		parsedGoVersion, osArch := utils.ParseGoVersion(goVersion)
		if parsedGoVersion != "" && parsedGoVersion != "Unknown version" {
			backupName = parsedGoVersion
		}
		if !utils.IsHostPlatform(osArch) {
			fmt.Printf("Current Go is %s; this machine is %s. Moving it aside.\n", osArch, utils.HostOSArch())
		}
	}

	backupPath := filepath.Join(config.GlobalConfig.GoExtractPathRoot, "go-"+backupName)
	utils.BackupOldGo(backupPath, config.GlobalConfig.GoFullPath)
	fmt.Printf("Old Go version backed up to: %s\n", backupPath)
}
