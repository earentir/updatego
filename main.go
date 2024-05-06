package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	cli "github.com/jawher/mow.cli"
)

const (
	goDownloadURL = "https://go.dev/dl/"
)

var (
	tempDir           = ""
	goExtractPathRoot = "/usr/local/"
	goFullPath        = ""
)

func main() {
	app := cli.App("go-installer", "A simple Go installer")

	installType := app.StringArg("TYPE", "global", "Installation type: user, global or <custom path>")

	app.Action = func() {
		tempDir = os.TempDir()

		switch *installType {
		case "user":
			homeDir, err := os.UserHomeDir()
			if err != nil {
				fmt.Println("Error getting user home directory:", err)
				os.Exit(1)
			}
			goExtractPathRoot = homeDir
		case "global":
			goExtractPathRoot = "/usr/local/"
		default:
			if *installType == "" {
				fmt.Println("Usage: program [user|global|<custom path>]")
				os.Exit(1)
			}
			goExtractPathRoot = *installType
		}

		if isWritable(goExtractPathRoot) {
			fmt.Printf("Install path is set to: %s\n", goExtractPathRoot)
		} else {
			fmt.Println("The provided install path is not valid or not writable.")
			os.Exit(1)
		}

		goFullPath = filepath.Join(goExtractPathRoot, "go")

		if dirNotEmpty(goFullPath) {
			goVersion, err := checkGoVersion(goFullPath)
			if err != nil {
				fmt.Println("Directory contains non-Go content. Do you want to continue? (yes/no):")
			} else {
				fmt.Printf("Existing Go version found: %s. Do you want to continue installing/updating? (yes/no): ", goVersion)
			}
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(response)) != "yes" {
				fmt.Println("Installation cancelled.")
				os.Exit(0)
			}
		}

		fmt.Println("Downloading Go Data to get the latest version...")
		htmlContent, err := downloadHTML(goDownloadURL)
		if err != nil {
			fmt.Println("Error downloading HTML content:", err)
			os.Exit(1)
		}

		fmt.Println("Finding the latest version...")
		version, err := findVersion(htmlContent)
		if err != nil {
			fmt.Println("Error finding the latest version:", err)
			os.Exit(1)
		}
		fmt.Println("Latest version found:", version)

		tempFile, err := os.CreateTemp(tempDir, "test_write_")
		if err != nil {
			fmt.Printf("No write permission in temp directory: %s\n", tempDir)
			os.Exit(1)
		} else {
			fmt.Printf("Write permission confirmed in temp directory: %s\n", tempDir)
			tempFile.Close()
			os.Remove(tempFile.Name())
		}

		filename := buildFilename(version)

		fmt.Println("Downloading the latest version:", filename, " From: ", goDownloadURL+filename)
		fileURL := goDownloadURL + filename
		filePath, err := downloadFile(fileURL)
		if err != nil {
			fmt.Println("Error downloading the file:", err)
			os.Exit(1)
		}

		fmt.Println("Removing the old Go folder...")
		if err := removeGoFolder(goFullPath); err != nil {
			fmt.Println("Error removing the old Go folder:", err)
			os.Exit(1)
		}

		fmt.Println("Extracting the new Go version...")
		if err := extractTarGz(filePath, goExtractPathRoot); err != nil {
			fmt.Println("Error extracting the Go archive:", err)
			os.Exit(1)
		}
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func findVersion(htmlContent string) (string, error) {
	regex := regexp.MustCompile(`go(\d+\.\d+\.\d+)\.linux-amd64\.tar\.gz`)
	matches := regex.FindStringSubmatch(htmlContent)
	if len(matches) < 2 {
		err := errors.New("No version found")
		return "", err
	}
	return matches[1], nil
}

func buildFilename(version string) string {
	return "go" + version + ".linux-amd64.tar.gz"
}

func removeGoFolder(path string) error {
	return os.RemoveAll(path)
}

func isWritable(path string) bool {
	tmpFilePath := filepath.Join(path, ".tmp-check")
	defer os.Remove(tmpFilePath)

	file, err := os.Create(tmpFilePath)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

func dirNotEmpty(path string) bool {
	files, err := os.ReadDir(path)
	return err == nil && len(files) > 0
}

func checkGoVersion(path string) (string, error) {
	goVersionPath := filepath.Join(path, "bin", "go")
	output, err := exec.Command(goVersionPath, "version").Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
