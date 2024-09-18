// Package update updates Go to the latest version
package update

import (
	"fmt"
	"os"
	"path/filepath"

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

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("Error checking downloaded file: %v\n", err)
		os.Exit(1)
	}
	if fileInfo.Size() == 0 {
		fmt.Println("Error: Downloaded file is empty")
		os.Exit(1)
	}

	goExtractPathRoot := "/usr/local/"
	goFullPath := filepath.Join(goExtractPathRoot, "go")

	if utils.DirNotEmpty(goFullPath) {
		goVersion, err := utils.CheckGoVersion(goFullPath)
		if err == nil {
			parsedGoVersion, _ := utils.ParseGoVersion(goVersion)
			backupPath := filepath.Join(goExtractPathRoot, "go-"+parsedGoVersion)
			if err := os.Rename(goFullPath, backupPath); err != nil {
				fmt.Printf("Error backing up old Go version: %v\n", err)
				fmt.Println("Proceeding with update without backup...")
			} else {
				fmt.Printf("Old Go version backed up to: %s\n", backupPath)
			}
		} else {
			fmt.Printf("Error checking current Go version: %v\n", err)
			fmt.Println("Proceeding with update...")
		}
	}

	fmt.Println("Extracting the new Go version...")
	if err := utils.ExtractTarGz(filePath, goFullPath, true); err != nil {
		fmt.Println("Error extracting the Go archive:", err)
		os.Exit(1)
	}

	fmt.Printf("Go has been successfully updated to version %s\n", version)
}
