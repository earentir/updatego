// Package update updates Go to the latest version
package update

import (
	"fmt"

	"updatego/config"
	"updatego/utils"
)

// Go updates Go to the latest version in the configured extract root.
func Go() error {
	if err := utils.RefuseWindows(); err != nil {
		return err
	}

	fmt.Println("Downloading Go Data to get the latest version...")
	htmlContent, err := utils.DownloadHTML(utils.GoDownloadURL)
	if err != nil {
		return fmt.Errorf("error downloading HTML content: %w", err)
	}

	fmt.Println("Finding the latest version...")
	version, err := utils.FindVersion(htmlContent)
	if err != nil {
		return fmt.Errorf("error finding the latest version: %w", err)
	}
	fmt.Println("Latest version found:", version)

	filename := utils.BuildFilename(version)
	fileURL := utils.GoDownloadURL + filename
	fmt.Println("Downloading the latest version:", filename, " From:", fileURL)

	filePath, err := utils.DownloadArchive(fileURL)
	if err != nil {
		return fmt.Errorf("error downloading the file: %w", err)
	}

	if utils.DirNotEmpty(config.GlobalConfig.GoFullPath) {
		if err := backupCurrentVersion(); err != nil {
			return err
		}
	}

	fmt.Println("Extracting the new Go version...")
	if err := utils.ExtractTarGz(filePath, config.GlobalConfig.GoFullPath, true); err != nil {
		return fmt.Errorf("error extracting the Go archive: %w", err)
	}
	if !utils.GoTreeReady(config.GlobalConfig.GoFullPath) {
		return fmt.Errorf("updated tree at %s is not a usable Go installation", config.GlobalConfig.GoFullPath)
	}

	fmt.Printf("Go has been successfully updated to version %s\n", version)
	return nil
}

func backupCurrentVersion() error {
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

	baseName := "go-" + backupName
	backupPath, err := utils.AllocateBackupDir(config.GlobalConfig.GoExtractPathRoot, baseName)
	if err != nil {
		return err
	}
	if err := utils.BackupOldGo(backupPath, config.GlobalConfig.GoFullPath); err != nil {
		return err
	}
	fmt.Printf("Old Go version backed up to: %s\n", backupPath)
	return nil
}
